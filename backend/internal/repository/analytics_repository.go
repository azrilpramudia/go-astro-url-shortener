package repository

import (
	"context"
	"time"

	"go-astro-url-shortener/internal/entity"

	"gorm.io/gorm"
)

type AnalyticsRepository struct{}

type DailyClicks struct {
	Date   string
	Clicks int64
}

type AnalyticsBreakdown struct {
	Name  string
	Count int64
}

type LinkAnalytics struct {
	TotalClicks    int64
	UniqueVisitors int64
	ClicksByDay    []DailyClicks
	Referrers      []AnalyticsBreakdown
	Countries      []AnalyticsBreakdown
	Devices        []AnalyticsBreakdown
}

func NewAnalyticsRepository() *AnalyticsRepository {
	return &AnalyticsRepository{}
}

func (r *AnalyticsRepository) Create(db *gorm.DB, event *entity.LinkAnalyticsEvent) error {
	return db.Create(event).Error
}

func (r *AnalyticsRepository) Summary(
	db *gorm.DB,
	linkID uint,
	from time.Time,
	until time.Time,
) (*LinkAnalytics, error) {
	summary := &LinkAnalytics{
		ClicksByDay: make([]DailyClicks, 0),
		Referrers:   make([]AnalyticsBreakdown, 0),
		Countries:   make([]AnalyticsBreakdown, 0),
		Devices:     make([]AnalyticsBreakdown, 0),
	}

	base := func() *gorm.DB {
		return db.Table("link_analytics_events").
			Where("link_id = ? AND occurred_at >= ? AND occurred_at < ?", linkID, from, until)
	}

	if err := base().
		Select("COUNT(*)").
		Scan(&summary.TotalClicks).Error; err != nil {
		return nil, err
	}

	if err := base().
		Select("COUNT(DISTINCT visitor_hash)").
		Scan(&summary.UniqueVisitors).Error; err != nil {
		return nil, err
	}

	if err := base().
		Select("TO_CHAR(DATE_TRUNC('day', occurred_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS date, COUNT(*) AS clicks").
		Group("date").
		Order("date ASC").
		Scan(&summary.ClicksByDay).Error; err != nil {
		return nil, err
	}

	if err := base().
		Select("COALESCE(NULLIF(referrer_host, ''), 'Direct') AS name, COUNT(*) AS count").
		Group("name").
		Order("count DESC, name ASC").
		Limit(10).
		Scan(&summary.Referrers).Error; err != nil {
		return nil, err
	}

	if err := base().
		Where("country_code <> ''").
		Select("country_code AS name, COUNT(*) AS count").
		Group("country_code").
		Order("count DESC, name ASC").
		Limit(10).
		Scan(&summary.Countries).Error; err != nil {
		return nil, err
	}

	if err := base().
		Select("device_type AS name, COUNT(*) AS count").
		Group("device_type").
		Order("count DESC, name ASC").
		Scan(&summary.Devices).Error; err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *AnalyticsRepository) DeleteBefore(db *gorm.DB, cutoff time.Time) (int64, error) {
	result := db.Where("occurred_at < ?", cutoff).Delete(&entity.LinkAnalyticsEvent{})
	return result.RowsAffected, result.Error
}

func (r *AnalyticsRepository) DeleteBeforeContext(ctx context.Context, db *gorm.DB, cutoff time.Time) (int64, error) {
	return r.DeleteBefore(db.WithContext(ctx), cutoff)
}
