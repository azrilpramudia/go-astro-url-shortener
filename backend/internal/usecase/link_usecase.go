package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/gateway/cache"
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
	DB                  *gorm.DB
	Log                 *logrus.Logger
	Validate            *validator.Validate
	LinkRepository      *repository.LinkRepository
	AnalyticsRepository *repository.AnalyticsRepository
	Cache               *cache.LinkCache
	BaseURL             string
}

func NewLinkUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	linkRepository *repository.LinkRepository, analyticsRepository *repository.AnalyticsRepository,
	linkCache *cache.LinkCache, baseURL string) *LinkUseCase {
	return &LinkUseCase{
		DB:                  db,
		Log:                 log,
		Validate:            validate,
		LinkRepository:      linkRepository,
		AnalyticsRepository: analyticsRepository,
		Cache:               linkCache,
		BaseURL:             strings.TrimRight(baseURL, "/"),
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

func (c *LinkUseCase) Resolve(
	ctx context.Context,
	code string,
	analytics model.RedirectAnalytics,
) (string, error) {
	notFound := fiber.NewError(fiber.StatusNotFound, "link not found")

	if !util.IsValidCode(code) {
		return "", notFound
	}

	entry, lookup := c.Cache.Get(ctx, code)
	switch lookup {
	case cache.NegativeHit:
		return "", notFound
	case cache.Miss:
		link := new(entity.Link)
		if err := c.LinkRepository.FindByCode(c.DB.WithContext(ctx), code, link); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Cache.SetNotFound(ctx, code)
				return "", notFound
			}
			c.Log.Warnf("failed to find link: %+v", err)
			return "", fiber.ErrInternalServerError
		}
		entry = &cache.LinkEntry{ID: link.ID, URL: link.OriginalURL, ExpiresAt: link.ExpiresAt}
		c.Cache.Set(ctx, code, entry)
	}

	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		return "", fiber.NewError(fiber.StatusGone, "link has expired")
	}

	c.countClick(ctx, entry.ID)
	event := &entity.LinkAnalyticsEvent{
		LinkID:       entry.ID,
		VisitorHash:  analytics.VisitorHash,
		ReferrerHost: analytics.ReferrerHost,
		DeviceType:   analytics.DeviceType,
		CountryCode:  analytics.CountryCode,
		OccurredAt:   time.Now().UTC(),
	}
	if err := c.AnalyticsRepository.Create(c.DB.WithContext(ctx), event); err != nil {
		c.Log.Warnf("failed to record analytics for link %d: %+v", entry.ID, err)
	}
	return entry.URL, nil
}

func (c *LinkUseCase) countClick(ctx context.Context, linkID uint) {
	if err := c.Cache.IncrementClicks(ctx, linkID); err == nil {
		return
	}
	if err := c.LinkRepository.IncrementClicks(c.DB.WithContext(ctx), linkID); err != nil {
		c.Log.Warnf("failed to increment clicks for link %d: %+v", linkID, err)
	}
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

func (c *LinkUseCase) GetAnalytics(
	ctx context.Context,
	userID uint,
	code string,
	days int,
) (*model.LinkAnalyticsResponse, error) {
	if days != 7 && days != 30 && days != 90 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "days must be 7, 30, or 90")
	}

	link := new(entity.Link)
	if err := c.LinkRepository.FindByCodeAndUser(c.DB.WithContext(ctx), code, userID, link); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "link not found")
		}
		c.Log.Warnf("failed to find link for analytics: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	until := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
	from := until.AddDate(0, 0, -days)
	summary, err := c.AnalyticsRepository.Summary(c.DB.WithContext(ctx), link.ID, from, until)
	if err != nil {
		c.Log.Warnf("failed to load analytics for link %d: %+v", link.ID, err)
		return nil, fiber.ErrInternalServerError
	}

	clicksByDate := make(map[string]int64, len(summary.ClicksByDay))
	for _, day := range summary.ClicksByDay {
		clicksByDate[day.Date] = day.Clicks
	}

	clicksByDay := make([]model.AnalyticsDailyClicks, 0, days)
	for date := from; date.Before(until); date = date.AddDate(0, 0, 1) {
		day := date.Format("2006-01-02")
		clicksByDay = append(clicksByDay, model.AnalyticsDailyClicks{
			Date:   day,
			Clicks: clicksByDate[day],
		})
	}

	toBreakdown := func(items []repository.AnalyticsBreakdown) []model.AnalyticsBreakdown {
		result := make([]model.AnalyticsBreakdown, 0, len(items))
		for _, item := range items {
			result = append(result, model.AnalyticsBreakdown{Name: item.Name, Count: item.Count})
		}
		return result
	}

	return &model.LinkAnalyticsResponse{
		Days:           days,
		Since:          from,
		TotalClicks:    summary.TotalClicks,
		UniqueVisitors: summary.UniqueVisitors,
		ClicksByDay:    clicksByDay,
		Referrers:      toBreakdown(summary.Referrers),
		Countries:      toBreakdown(summary.Countries),
		Devices:        toBreakdown(summary.Devices),
	}, nil
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
	c.Cache.Delete(ctx, code)

	return nil
}
