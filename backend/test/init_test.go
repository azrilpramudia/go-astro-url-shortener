package test

import (
	"strings"

	"go-astro-url-shortener/internal/config"
	"go-astro-url-shortener/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	app              *fiber.App
	db               *gorm.DB
	redisClient      *redis.Client
	flusher          *usecase.ClickFlusher
	analyticsCleaner *usecase.AnalyticsCleaner
	viperConfig      *viper.Viper
)

func init() {
	viperConfig = config.NewViper()
	viperConfig.Set("DB_NAME", viperConfig.GetString("DB_NAME")+"_test")
	viperConfig.Set("REDIS_DB", 1)
	if viperConfig.GetString("JWT_SECRET") == "" {
		viperConfig.Set("JWT_SECRET", strings.Repeat("t", 48))
	}
	viperConfig.Set("RATE_LIMIT_CREATE", 1000)
	viperConfig.Set("RATE_LIMIT_AUTH", 1000)

	log := config.NewLogger(viperConfig)
	log.SetLevel(logrus.WarnLevel)

	db = config.NewDatabase(viperConfig, log)
	redisClient = config.NewRedis(viperConfig, log)
	app = config.NewFiber(viperConfig)

	flusher, analyticsCleaner = config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		Redis:    redisClient,
		App:      app,
		Log:      log,
		Validate: config.NewValidator(),
		Config:   viperConfig,
	})
}
