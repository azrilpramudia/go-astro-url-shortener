package model

import "time"

type CreateLinkRequest struct {
	URL       string     `json:"url" validate:"required,url,max=2048"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type LinkResponse struct {
	ID          uint       `json:"id"`
	Code        string     `json:"code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Clicks      int64      `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
