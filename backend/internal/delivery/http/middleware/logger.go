package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RequestLogger(log *logrus.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()

		status := ctx.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			}
		}

		log.WithFields(logrus.Fields{
			"method":   ctx.Method(),
			"path":     ctx.Path(),
			"status":   status,
			"duration": time.Since(start).String(),
			"ip":       ctx.IP(),
		}).Info("request")

		return err
	}
}
