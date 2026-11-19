package repository

import (
	"go-astro-url-shortener/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository struct {
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) *UserRepository {
	return &UserRepository{Log: log}
}

func (r *UserRepository) Create(db *gorm.DB, user *entity.User) error {
	return db.Create(user).Error
}

func (r *UserRepository) FindByEmail(db *gorm.DB, email string, user *entity.User) error {
	return db.Where("email = ?", email).Take(user).Error
}

func (r *UserRepository) FindByID(db *gorm.DB, id uint, user *entity.User) error {
	return db.Take(user, id).Error
}
