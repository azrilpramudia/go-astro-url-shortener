package entity

import "time"

type LinkAnalyticsEvent struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	LinkID       uint      `gorm:"column:link_id;not null"`
	VisitorHash  string    `gorm:"column:visitor_hash;size:64;not null"`
	ReferrerHost string    `gorm:"column:referrer_host;size:255;not null"`
	DeviceType   string    `gorm:"column:device_type;size:16;not null"`
	CountryCode  string    `gorm:"column:country_code;size:2;not null"`
	OccurredAt   time.Time `gorm:"column:occurred_at;not null"`
}

func (e *LinkAnalyticsEvent) TableName() string {
	return "link_analytics_events"
}
