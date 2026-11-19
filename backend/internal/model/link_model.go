package model

import "time"

type CreateLinkRequest struct {
	URL        string     `json:"url" validate:"required,url,max=2048"`
	CustomCode string     `json:"custom_code,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
}

type UpdateLinkRequest struct {
	URL       string     `json:"url" validate:"required,url,max=2048"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
}

type LinkListFilters struct {
	Query  string
	Status string
	Tag    string
}

type LinkResponse struct {
	ID          uint       `json:"id"`
	Code        string     `json:"code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Clicks      int64      `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	Tags        []string   `json:"tags,omitempty"`
}

type ListLinksResponse struct {
	Items []LinkResponse `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}
