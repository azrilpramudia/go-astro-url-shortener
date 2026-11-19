package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"
	"go-astro-url-shortener/internal/model/converter"
	"go-astro-url-shortener/internal/repository"
	"go-astro-url-shortener/internal/util"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	codeLength     = 7
	maxCodeRetries = 5
)

type LinkUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	LinkRepository *repository.LinkRepository
	BaseURL        string
}

func NewLinkUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	linkRepository *repository.LinkRepository, baseURL string) *LinkUseCase {
	return &LinkUseCase{
		DB:             db,
		Log:            log,
		Validate:       validate,
		LinkRepository: linkRepository,
		BaseURL:        strings.TrimRight(baseURL, "/"),
	}
}

func (c *LinkUseCase) Create(ctx context.Context, userID *uint, request *model.CreateLinkRequest) (*model.LinkResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("invalid request body: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request: url is required and must be a valid URL")
	}

	parsed, err := url.Parse(request.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "url must start with http:// or https://")
	}

	if request.ExpiresAt != nil && !request.ExpiresAt.After(time.Now()) {
		return nil, fiber.NewError(fiber.StatusBadRequest, "expires_at must be in the future")
	}

	db := c.DB.WithContext(ctx)

	for i := 0; i < maxCodeRetries; i++ {
		code, err := util.RandomCode(codeLength)
		if err != nil {
			c.Log.Warnf("failed to generate code: %+v", err)
			return nil, fiber.ErrInternalServerError
		}

		link := &entity.Link{
			Code:        code,
			OriginalURL: request.URL,
			ExpiresAt:   request.ExpiresAt,
			UserID:      userID,
		}

		err = c.LinkRepository.Create(db, link)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			continue
		}
		if err != nil {
			c.Log.Warnf("failed to create link: %+v", err)
			return nil, fiber.ErrInternalServerError
		}

		return converter.LinkToResponse(link, c.BaseURL), nil
	}

	c.Log.Warn("failed to generate unique code after retries")
	return nil, fiber.ErrInternalServerError
}

func (c *LinkUseCase) Resolve(ctx context.Context, code string) (string, error) {
	db := c.DB.WithContext(ctx)

	link := new(entity.Link)
	if err := c.LinkRepository.FindByCode(db, code, link); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fiber.NewError(fiber.StatusNotFound, "link not found")
		}
		c.Log.Warnf("failed to find link: %+v", err)
		return "", fiber.ErrInternalServerError
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return "", fiber.NewError(fiber.StatusGone, "link has expired")
	}

	if err := c.LinkRepository.IncrementClicks(db, link.ID); err != nil {
		c.Log.Warnf("failed to increment clicks for %s: %+v", code, err)
	}

	return link.OriginalURL, nil
}

func (c *LinkUseCase) Get(ctx context.Context, code string) (*model.LinkResponse, error) {
	link := new(entity.Link)
	if err := c.LinkRepository.FindByCode(c.DB.WithContext(ctx), code, link); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "link not found")
		}
		c.Log.Warnf("failed to find link: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.LinkToResponse(link, c.BaseURL), nil
}

func (c *LinkUseCase) ListByUser(ctx context.Context, userID uint, page, size int) (*model.ListLinksResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	db := c.DB.WithContext(ctx)

	total, err := c.LinkRepository.CountByUser(db, userID)
	if err != nil {
		c.Log.Warnf("failed to count links: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	links, err := c.LinkRepository.FindByUser(db, userID, size, (page-1)*size)
	if err != nil {
		c.Log.Warnf("failed to list links: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	items := make([]model.LinkResponse, 0, len(links))
	for i := range links {
		items = append(items, *converter.LinkToResponse(&links[i], c.BaseURL))
	}

	return &model.ListLinksResponse{Items: items, Total: total, Page: page, Size: size}, nil
}

func (c *LinkUseCase) Delete(ctx context.Context, userID uint, code string) error {
	affected, err := c.LinkRepository.DeleteByCodeAndUser(c.DB.WithContext(ctx), code, userID)
	if err != nil {
		c.Log.Warnf("failed to delete link: %+v", err)
		return fiber.ErrInternalServerError
	}
	if affected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "link not found")
	}
	return nil
}
