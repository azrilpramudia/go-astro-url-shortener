package main

import (
	"fmt"

	"go-astro-url-shortener/internal/config"
)

func main() {
	viper := config.NewViper()
	log := config.NewLogger(viper)
	db := config.NewDatabase(viper, log)
	validate := config.NewValidator()
	app := config.NewFiber(viper)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viper,
	})

	port := viper.GetInt("WEB_PORT")
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
