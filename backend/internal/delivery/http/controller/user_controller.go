package controller

import (
	"time"

	"go-astro-url-shortener/internal/delivery/http/middleware"
	"go-astro-url-shortener/internal/model"
	"go-astro-url-shortener/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	Log          *logrus.Logger
	UseCase      *usecase.UserUseCase
	CookieSecure bool
	TokenTTL     time.Duration
}

func NewUserController(log *logrus.Logger, useCase *usecase.UserUseCase, cookieSecure bool, ttl time.Duration) *UserController {
	return &UserController{Log: log, UseCase: useCase, CookieSecure: cookieSecure, TokenTTL: ttl}
}

func (c *UserController) setCookie(ctx *fiber.Ctx, value string, expires time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     middleware.CookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   c.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
	request := new(model.RegisterRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	result, err := c.UseCase.Register(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	c.setCookie(ctx, result.Token, time.Now().Add(c.TokenTTL))
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.UserResponse]{Data: result.User})
}

func (c *UserController) Login(ctx *fiber.Ctx) error {
	request := new(model.LoginRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	result, err := c.UseCase.Login(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	c.setCookie(ctx, result.Token, time.Now().Add(c.TokenTTL))
	return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: result.User})
}

func (c *UserController) Logout(ctx *fiber.Ctx) error {
	c.setCookie(ctx, "", time.Now().Add(-time.Hour))
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *UserController) Me(ctx *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(ctx)

	user, err := c.UseCase.Current(ctx.UserContext(), userID)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: user})
}
