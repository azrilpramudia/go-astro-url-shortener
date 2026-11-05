package repository

import (
	"go-astro-url-shortener/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LinkRepository struct {
	Log *logrus.Logger
}

func NewLinkRepository(log *logrus.Logger) *LinkRepository {
	return &LinkRepository{Log: log}
}

func (r *LinkRepository) Create(db *gorm.DB, link *entity.Link) error {
	return db.Create(link).Error
}

func (r *LinkRepository) FindByCode(db *gorm.DB, code string, link *entity.Link) error {
	return db.Where("code = ?", code).Take(link).Error
}

func (r *LinkRepository) IncrementClicks(db *gorm.DB, id uint) error {
	return db.Model(&entity.Link{}).
		Where("id = ?", id).
		UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}
