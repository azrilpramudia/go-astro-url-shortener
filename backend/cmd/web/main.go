package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-astro-url-shortener/internal/config"
)

func main() {
	viper := config.NewViper()
	log := config.NewLogger(viper)
	db := config.NewDatabase(viper, log)
	redisClient := config.NewRedis(viper, log)
	validate := config.NewValidator()
	app := config.NewFiber(viper)

	flusher := config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		Redis:    redisClient,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viper,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go flusher.Run(ctx)

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
