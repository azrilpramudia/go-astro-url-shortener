package config

import (
	"time"

	"go-astro-url-shortener/internal/delivery/http/controller"
	"go-astro-url-shortener/internal/delivery/http/middleware"
	"go-astro-url-shortener/internal/delivery/http/route"
	"go-astro-url-shortener/internal/repository"
	"go-astro-url-shortener/internal/usecase"
	"go-astro-url-shortener/internal/util"

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
	secret := cfg.Config.GetString("JWT_SECRET")
	if len(secret) < 32 {
		cfg.Log.Fatal("JWT_SECRET must be set and at least 32 characters long")
	}
	tokens := util.NewTokenManager(secret, time.Duration(cfg.Config.GetInt("JWT_TTL_HOURS"))*time.Hour)

	cfg.App.Use(middleware.RequestLogger(cfg.Log))
	cfg.App.Use(recover.New())
	cfg.App.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Config.GetString("WEB_CORS_ORIGIN"),
		AllowMethods:     "GET,POST,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept",
		AllowCredentials: true,
	}))

	// repository
	linkRepository := repository.NewLinkRepository(cfg.Log)
	userRepository := repository.NewUserRepository(cfg.Log)

	// usecase
	linkUseCase := usecase.NewLinkUseCase(
		cfg.DB, cfg.Log, cfg.Validate, linkRepository,
		cfg.Config.GetString("APP_BASE_URL"),
	)
	userUseCase := usecase.NewUserUseCase(cfg.DB, cfg.Log, cfg.Validate, userRepository, tokens)

	// controller
	linkController := controller.NewLinkController(cfg.Log, linkUseCase)
	userController := controller.NewUserController(
		cfg.Log, userUseCase, cfg.Config.GetBool("COOKIE_SECURE"), tokens.TTL,
	)

	// route
	routeConfig := route.RouteConfig{
		App:            cfg.App,
		LinkController: linkController,
		UserController: userController,
		RequireAuth:    middleware.Auth(tokens, true),
		OptionalAuth:   middleware.Auth(tokens, false),
		CreateLimit:    cfg.Config.GetInt("RATE_LIMIT_CREATE"),
		AuthLimit:      cfg.Config.GetInt("RATE_LIMIT_AUTH"),
	}
	routeConfig.Setup()
}
