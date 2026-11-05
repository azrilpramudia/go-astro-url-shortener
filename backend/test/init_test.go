package test

import (
	"go-astro-url-shortener/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	app         *fiber.App
	db          *gorm.DB
	viperConfig *viper.Viper
)

func init() {
	viperConfig = config.NewViper()
	viperConfig.Set("DB_NAME", viperConfig.GetString("DB_NAME")+"_test")

	log := config.NewLogger(viperConfig)
	log.SetLevel(logrus.WarnLevel)

	db = config.NewDatabase(viperConfig, log)
	app = config.NewFiber(viperConfig)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: config.NewValidator(),
		Config:   viperConfig,
	})
}
