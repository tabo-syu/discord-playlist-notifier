package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *categoryRepository {
	return &categoryRepository{db}
}

func (r *categoryRepository) FindAll() (map[library.CategoryID]*library.Category, error) {
	var recs []*categoryRecord
	if err := r.db.Find(&recs).Error; err != nil {
		return nil, err
	}

	categories := map[library.CategoryID]*library.Category{}
	for _, rec := range recs {
		categories[library.CategoryID(rec.YoutubeID)] = rec.toDomain()
	}

	return categories, nil
}

func (r *categoryRepository) Save(categories []*library.Category) error {
	if len(categories) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "assignable", "updated_at"}),
	}).Create(mapSlice(categories, toCategoryRecord)).Error
}
