package controller

import (
	"time"

	"go-astro-url-shortener/internal/delivery/http/middleware"
	"go-astro-url-shortener/internal/gateway/geoip"
	"go-astro-url-shortener/internal/model"
	"go-astro-url-shortener/internal/usecase"
	"go-astro-url-shortener/internal/util"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type LinkController struct {
	Log           *logrus.Logger
	UseCase       *usecase.LinkUseCase
	GeoIP         *geoip.Locator
	VisitorSecret string
	VisitorAge    time.Duration
	CookieSecure  bool
}

func NewLinkController(
	log *logrus.Logger,
	useCase *usecase.LinkUseCase,
	geoIP *geoip.Locator,
	visitorSecret string,
	visitorAge time.Duration,
	cookieSecure bool,
) *LinkController {
	return &LinkController{
		Log:           log,
		UseCase:       useCase,
		GeoIP:         geoIP,
		VisitorSecret: visitorSecret,
		VisitorAge:    visitorAge,
		CookieSecure:  cookieSecure,
	}
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

	response, err := c.UseCase.ListByUser(
		ctx.UserContext(),
		userID,
		ctx.QueryInt("page", 1),
		ctx.QueryInt("size", 20),
		model.LinkListFilters{
			Query:  ctx.Query("q"),
			Status: ctx.Query("status"),
			Tag:    ctx.Query("tag"),
		},
	)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.ListLinksResponse]{Data: response})
}

func (c *LinkController) Update(ctx *fiber.Ctx) error {
	request := new(model.UpdateLinkRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("failed to parse update request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	userID, _ := middleware.GetUserID(ctx)
	response, err := c.UseCase.Update(ctx.UserContext(), userID, ctx.Params("code"), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.LinkResponse]{Data: response})
}

func (c *LinkController) Delete(ctx *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(ctx)

	if err := c.UseCase.Delete(ctx.UserContext(), userID, ctx.Params("code")); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *LinkController) Redirect(ctx *fiber.Ctx) error {
	visitorToken := ctx.Cookies(util.VisitorCookieName)
	newVisitorToken := ""
	if !util.IsVisitorToken(visitorToken) {
		token, err := util.NewVisitorToken()
		if err != nil {
			c.Log.Errorf("failed to create analytics visitor token: %v", err)
			return fiber.ErrInternalServerError
		}
		visitorToken = token
		newVisitorToken = token
	}

	analytics := model.RedirectAnalytics{
		VisitorHash:  util.HashVisitorToken(c.VisitorSecret, visitorToken),
		ReferrerHost: util.ReferrerHost(ctx.Get("Referer")),
		DeviceType:   util.DeviceType(ctx.Get("User-Agent")),
	}
	if c.GeoIP != nil {
		analytics.CountryCode = c.GeoIP.CountryCode(ctx.IP())
	}

	url, err := c.UseCase.Resolve(ctx.UserContext(), ctx.Params("code"), analytics)
	if err != nil {
		return err
	}

	if newVisitorToken != "" {
		ctx.Cookie(&fiber.Cookie{
			Name:     util.VisitorCookieName,
			Value:    newVisitorToken,
			Path:     "/",
			MaxAge:   int(c.VisitorAge.Seconds()),
			Expires:  time.Now().Add(c.VisitorAge),
			HTTPOnly: true,
			Secure:   c.CookieSecure,
			SameSite: fiber.CookieSameSiteLaxMode,
		})
	}

	return ctx.Redirect(url, fiber.StatusFound)
}

func (c *LinkController) Analytics(ctx *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(ctx)
	response, err := c.UseCase.GetAnalytics(
		ctx.UserContext(),
		userID,
		ctx.Params("code"),
		ctx.QueryInt("days", 30),
	)
	if err != nil {
		return err
	}
	response.LocationAvailable = c.GeoIP != nil

	return ctx.JSON(model.WebResponse[*model.LinkAnalyticsResponse]{Data: response})
}

func (c *LinkController) Get(ctx *fiber.Ctx) error {
	response, err := c.UseCase.Get(ctx.UserContext(), ctx.Params("code"))
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.LinkResponse]{Data: response})
}
