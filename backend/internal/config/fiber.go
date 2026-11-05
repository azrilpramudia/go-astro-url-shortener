package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func NewFiber(v *viper.Viper) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      v.GetString("APP_NAME"),
		Prefork:      v.GetBool("WEB_PREFORK"),
		ErrorHandler: NewErrorHandler(),
	})
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		return ctx.Status(code).JSON(fiber.Map{"errors": err.Error()})
	}
}
