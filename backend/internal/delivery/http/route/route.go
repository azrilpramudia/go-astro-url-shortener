package route

import (
	"time"

	"go-astro-url-shortener/internal/delivery/http/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RouteConfig struct {
	App            *fiber.App
	LinkController *controller.LinkController
	UserController *controller.UserController
	RequireAuth    fiber.Handler
	OptionalAuth   fiber.Handler
	CreateLimit    int
	AuthLimit      int
}

func newLimiter(max int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: time.Minute,
		LimitReached: func(ctx *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests, try again later")
		},
	})
}

func (c *RouteConfig) Setup() {
	authLimiter := newLimiter(c.AuthLimit)

	c.App.Get("/api/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	// auth
	c.App.Post("/api/auth/register", authLimiter, c.UserController.Register)
	c.App.Post("/api/auth/login", authLimiter, c.UserController.Login)
	c.App.Post("/api/auth/logout", c.UserController.Logout)
	c.App.Get("/api/auth/me", c.RequireAuth, c.UserController.Me)

	// links
	c.App.Post("/api/links", newLimiter(c.CreateLimit), c.OptionalAuth, c.LinkController.Create)
	c.App.Get("/api/me/links", c.RequireAuth, c.LinkController.ListMine)
	c.App.Get("/api/me/links/:code/analytics", c.RequireAuth, c.LinkController.Analytics)
	c.App.Get("/api/links/:code", c.LinkController.Get)
	c.App.Delete("/api/links/:code", c.RequireAuth, c.LinkController.Delete)
	c.App.Get("/:code", c.LinkController.Redirect)
}
