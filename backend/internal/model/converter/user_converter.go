package converter

import (
	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"
)

func UserToResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}
}
