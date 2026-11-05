package converter

import (
	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"
)

func LinkToResponse(link *entity.Link, baseURL string) *model.LinkResponse {
	return &model.LinkResponse{
		ID:          link.ID,
		Code:        link.Code,
		ShortURL:    baseURL + "/" + link.Code,
		OriginalURL: link.OriginalURL,
		Clicks:      link.Clicks,
		ExpiresAt:   link.ExpiresAt,
		CreatedAt:   link.CreatedAt,
	}
}
