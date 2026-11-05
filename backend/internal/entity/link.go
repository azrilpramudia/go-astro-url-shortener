package entity

import "time"

type Link struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement"`
	Code        string     `gorm:"column:code;uniqueIndex"`
	OriginalURL string     `gorm:"column:original_url"`
	Clicks      int64      `gorm:"column:clicks"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
}

func (l *Link) TableName() string {
	return "links"
}
