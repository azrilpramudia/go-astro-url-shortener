package controller

import (
	"go-astro-url-shortener/internal/model"
	"go-astro-url-shortener/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type LinkController struct {
	Log     *logrus.Logger
	UseCase *usecase.LinkUseCase
}

func NewLinkController(log *logrus.Logger, useCase *usecase.LinkUseCase) *LinkController {
	return &LinkController{Log: log, UseCase: useCase}
}

func (c *LinkController) Create(ctx *fiber.Ctx) error {
	request := new(model.CreateLinkRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("failed to parse request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	response, err := c.UseCase.Create(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.LinkResponse]{Data: response})
}

func (c *LinkController) Redirect(ctx *fiber.Ctx) error {
	url, err := c.UseCase.Resolve(ctx.UserContext(), ctx.Params("code"))
	if err != nil {
		return err
	}

	return ctx.Redirect(url, fiber.StatusFound)
}
