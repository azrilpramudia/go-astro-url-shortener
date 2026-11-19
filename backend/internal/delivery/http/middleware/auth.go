package middleware

import (
	"go-astro-url-shortener/internal/util"

	"github.com/gofiber/fiber/v2"
)

const (
	CookieName = "token"
	userIDKey  = "user_id"
)

func Auth(tokens *util.TokenManager, required bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if token := ctx.Cookies(CookieName); token != "" {
			if id, err := tokens.Parse(token); err == nil {
				ctx.Locals(userIDKey, id)
				return ctx.Next()
			}
		}
		if required {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		return ctx.Next()
	}
}

func GetUserID(ctx *fiber.Ctx) (uint, bool) {
	id, ok := ctx.Locals(userIDKey).(uint)
	return id, ok
}
