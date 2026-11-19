package controller

import (
	"go-astro-url-shortener/internal/delivery/http/middleware"
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

	var ownerID *uint
	if id, ok := middleware.GetUserID(ctx); ok {
		ownerID = &id
	}

	response, err := c.UseCase.Create(ctx.UserContext(), ownerID, request)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.LinkResponse]{Data: response})
}

func (c *LinkController) ListMine(ctx *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(ctx)

	response, err := c.UseCase.ListByUser(ctx.UserContext(), userID, ctx.QueryInt("page", 1), ctx.QueryInt("size", 20))
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.ListLinksResponse]{Data: response})
}

func (c *LinkController) Delete(ctx *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(ctx)

	if err := c.UseCase.Delete(ctx.UserContext(), userID, ctx.Params("code")); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *LinkController) Redirect(ctx *fiber.Ctx) error {
	url, err := c.UseCase.Resolve(ctx.UserContext(), ctx.Params("code"))
	if err != nil {
		return err
	}

	return ctx.Redirect(url, fiber.StatusFound)
}

func (c *LinkController) Get(ctx *fiber.Ctx) error {
	response, err := c.UseCase.Get(ctx.UserContext(), ctx.Params("code"))
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.LinkResponse]{Data: response})
}
