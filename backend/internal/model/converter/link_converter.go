package converter

import (
	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"
)

func LinkToResponse(link *entity.Link, baseURL string) *model.LinkResponse {
	tags := make([]string, 0, len(link.Tags))
	for _, tag := range link.Tags {
		tags = append(tags, tag.Name)
	}

	return &model.LinkResponse{
		ID:          link.ID,
		Code:        link.Code,
		ShortURL:    baseURL + "/" + link.Code,
		OriginalURL: link.OriginalURL,
		Clicks:      link.Clicks,
		ExpiresAt:   link.ExpiresAt,
		CreatedAt:   link.CreatedAt,
		Tags:        tags,
	}
}
