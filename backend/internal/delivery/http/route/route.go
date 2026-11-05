package route

import (
	"go-astro-url-shortener/internal/delivery/http/controller"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RouteConfig struct {
	App            *fiber.App
	LinkController *controller.LinkController
}

func (c *RouteConfig) Setup() {
	c.App.Get("/api/hello", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	createLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		LimitReached: func(ctx *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests, try again later")
		},
	})

	c.App.Post("/api/links", createLimiter, c.LinkController.Create)
	c.App.Get("/:code", c.LinkController.Redirect)
}
