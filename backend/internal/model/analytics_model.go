package model

import "time"

type RedirectAnalytics struct {
	VisitorHash  string
	ReferrerHost string
	DeviceType   string
	CountryCode  string
}

type AnalyticsDailyClicks struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

type AnalyticsBreakdown struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type LinkAnalyticsResponse struct {
	Days              int                    `json:"days"`
	Since             time.Time              `json:"since"`
	TotalClicks       int64                  `json:"total_clicks"`
	UniqueVisitors    int64                  `json:"unique_visitors"`
	LocationAvailable bool                   `json:"location_available"`
	ClicksByDay       []AnalyticsDailyClicks `json:"clicks_by_day"`
	Referrers         []AnalyticsBreakdown   `json:"referrers"`
	Countries         []AnalyticsBreakdown   `json:"countries"`
	Devices           []AnalyticsBreakdown   `json:"devices"`
}
