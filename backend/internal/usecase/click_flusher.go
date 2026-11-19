package usecase

import (
	"context"
	"time"

	"go-astro-url-shortener/internal/gateway/cache"
	"go-astro-url-shortener/internal/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ClickFlusher struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Cache          *cache.LinkCache
	LinkRepository *repository.LinkRepository
	Interval       time.Duration
}

func NewClickFlusher(db *gorm.DB, log *logrus.Logger, linkCache *cache.LinkCache,
	linkRepository *repository.LinkRepository, interval time.Duration) *ClickFlusher {
	return &ClickFlusher{DB: db, Log: log, Cache: linkCache, LinkRepository: linkRepository, Interval: interval}
}

func (f *ClickFlusher) Run(ctx context.Context) {
	ticker := time.NewTicker(f.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := f.Flush(ctx); err != nil {
				f.Log.Warnf("failed to flush clicks: %v", err)
			}
		}
	}
}

func (f *ClickFlusher) Flush(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	locked, err := f.Cache.AcquireFlushLock(ctx, 30*time.Second)
	if err != nil || !locked {
		return err
	}
	defer f.Cache.ReleaseFlushLock(context.Background())

	batch, err := f.Cache.LoadClickBatch(ctx)
	if err != nil || len(batch) == 0 {
		return err
	}

	if err := f.LinkRepository.AddClicks(f.DB.WithContext(ctx), batch); err != nil {
		return err
	}

	return f.Cache.CommitClickBatch(ctx)
}
