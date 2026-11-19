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

func (r *LinkRepository) FindByCodeAndUser(db *gorm.DB, code string, userID uint, link *entity.Link) error {
	return db.Where("code = ? AND user_id = ?", code, userID).Take(link).Error
}

func (r *LinkRepository) IncrementClicks(db *gorm.DB, id uint) error {
	return db.Model(&entity.Link{}).
		Where("id = ?", id).
		UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}

func (r *LinkRepository) FindByUser(db *gorm.DB, userID uint, limit, offset int) ([]entity.Link, error) {
	var links []entity.Link
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&links).Error
	return links, err
}

func (r *LinkRepository) CountByUser(db *gorm.DB, userID uint) (int64, error) {
	var total int64
	err := db.Model(&entity.Link{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

func (r *LinkRepository) DeleteByCodeAndUser(db *gorm.DB, code string, userID uint) (int64, error) {
	res := db.Where("code = ? AND user_id = ?", code, userID).Delete(&entity.Link{})
	return res.RowsAffected, res.Error
}

func (r *LinkRepository) AddClicks(db *gorm.DB, batch map[uint]int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for id, n := range batch {
			if err := tx.Model(&entity.Link{}).
				Where("id = ?", id).
				UpdateColumn("clicks", gorm.Expr("clicks + ?", n)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
