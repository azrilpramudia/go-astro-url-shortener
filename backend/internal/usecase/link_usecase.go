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

func (c *LinkUseCase) Create(ctx context.Context, request *model.CreateLinkRequest) (*model.LinkResponse, error) {
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
