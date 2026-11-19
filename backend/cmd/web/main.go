package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-astro-url-shortener/internal/config"
	"go-astro-url-shortener/internal/gateway/geoip"
)

func main() {
	viper := config.NewViper()
	log := config.NewLogger(viper)
	db := config.NewDatabase(viper, log)
	redisClient := config.NewRedis(viper, log)
	validate := config.NewValidator()
	app := config.NewFiber(viper)
	geoIP, err := geoip.Open(viper.GetString("GEOIP_DB_PATH"), log)
	if err != nil {
		log.Fatalf("failed to open GeoIP database: %v", err)
	}
	if geoIP != nil {
		defer func() {
			if err := geoIP.Close(); err != nil {
				log.Warnf("failed to close GeoIP database: %v", err)
			}
		}()
	}

	flusher, analyticsCleaner := config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		Redis:    redisClient,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viper,
		GeoIP:    geoIP,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go flusher.Run(ctx)
	go analyticsCleaner.Run(ctx)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", viper.GetInt("WEB_PORT"))); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Warnf("server shutdown: %v", err)
	}
	if err := flusher.Flush(shutdownCtx); err != nil {
		log.Warnf("final click flush: %v", err)
	}
}
