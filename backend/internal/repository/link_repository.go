package repository

import (
	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"

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

func (r *LinkRepository) FindByUser(
	db *gorm.DB,
	userID uint,
	filters model.LinkListFilters,
	limit,
	offset int,
) ([]entity.Link, error) {
	var links []entity.Link
	err := applyLinkFilters(db, userID, filters).
		Preload("Tags").
		Order("links.created_at DESC, links.id DESC").
		Limit(limit).Offset(offset).
		Find(&links).Error
	return links, err
}

func (r *LinkRepository) CountByUser(db *gorm.DB, userID uint, filters model.LinkListFilters) (int64, error) {
	var total int64
	err := applyLinkFilters(db, userID, filters).Count(&total).Error
	return total, err
}

func (r *LinkRepository) ReplaceTags(db *gorm.DB, linkID uint, tags []string) error {
	if err := db.Where("link_id = ?", linkID).Delete(&entity.LinkTag{}).Error; err != nil {
		return err
	}
	if len(tags) == 0 {
		return nil
	}

	entities := make([]entity.LinkTag, 0, len(tags))
	for _, name := range tags {
		entities = append(entities, entity.LinkTag{LinkID: linkID, Name: name})
	}
	return db.Create(&entities).Error
}

func (r *LinkRepository) DeleteByCodeAndUser(db *gorm.DB, code string, userID uint) (int64, error) {
	res := db.Where("code = ? AND user_id = ?", code, userID).Delete(&entity.Link{})
	return res.RowsAffected, res.Error
}

func applyLinkFilters(db *gorm.DB, userID uint, filters model.LinkListFilters) *gorm.DB {
	query := db.Model(&entity.Link{}).Where("links.user_id = ?", userID)
	if filters.Query != "" {
		pattern := "%" + filters.Query + "%"
		query = query.Where("(links.code ILIKE ? OR links.original_url ILIKE ?)", pattern, pattern)
	}
	switch filters.Status {
	case "active":
		query = query.Where("(links.expires_at IS NULL OR links.expires_at > NOW())")
	case "expired":
		query = query.Where("links.expires_at IS NOT NULL AND links.expires_at <= NOW()")
	}
	if filters.Tag != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM link_tags WHERE link_tags.link_id = links.id AND link_tags.name = ?)",
			filters.Tag,
		)
	}
	return query
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
