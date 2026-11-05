package config

import (
	"go-astro-url-shortener/internal/delivery/http/controller"
	"go-astro-url-shortener/internal/delivery/http/middleware"
	"go-astro-url-shortener/internal/delivery/http/route"
	"go-astro-url-shortener/internal/repository"
	"go-astro-url-shortener/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(cfg *BootstrapConfig) {
	cfg.App.Use(middleware.RequestLogger(cfg.Log))
	cfg.App.Use(recover.New())
	cfg.App.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Config.GetString("WEB_CORS_ORIGIN"),
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))
	// repository
	linkRepository := repository.NewLinkRepository(cfg.Log)

	// usecase
	linkUseCase := usecase.NewLinkUseCase(
		cfg.DB, cfg.Log, cfg.Validate, linkRepository,
		cfg.Config.GetString("APP_BASE_URL"),
	)

	// controller
	linkController := controller.NewLinkController(cfg.Log, linkUseCase)

	// route
	routeConfig := route.RouteConfig{
		App:            cfg.App,
		LinkController: linkController,
	}
	routeConfig.Setup()
}
