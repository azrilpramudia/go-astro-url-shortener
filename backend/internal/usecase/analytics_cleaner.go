package usecase

import (
	"context"
	"time"

	"go-astro-url-shortener/internal/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AnalyticsCleaner struct {
	Repository   *repository.AnalyticsRepository
	DB           *gorm.DB
	Log          *logrus.Logger
	Retention    time.Duration
	CleanupEvery time.Duration
}

func NewAnalyticsCleaner(
	repository *repository.AnalyticsRepository,
	db *gorm.DB,
	log *logrus.Logger,
	retention time.Duration,
) *AnalyticsCleaner {
	return &AnalyticsCleaner{
		Repository:   repository,
		DB:           db,
		Log:          log,
		Retention:    retention,
		CleanupEvery: 24 * time.Hour,
	}
}

func (c *AnalyticsCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(c.CleanupEvery)
	defer ticker.Stop()

	c.Cleanup(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Cleanup(ctx)
		}
	}
}

func (c *AnalyticsCleaner) Cleanup(ctx context.Context) {
	cutoff := time.Now().UTC().Add(-c.Retention)
	deleted, err := c.Repository.DeleteBeforeContext(ctx, c.DB, cutoff)
	if err != nil {
		c.Log.Warnf("failed to remove expired analytics events: %v", err)
		return
	}
	if deleted > 0 {
		c.Log.Infof("removed %d analytics events older than %s", deleted, cutoff.Format(time.RFC3339))
	}
}
