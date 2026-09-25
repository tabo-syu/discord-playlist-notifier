package repository

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LibraryRepository stores the contents of watched YouTube playlists.
type LibraryRepository interface {
	FindItems(playlistYoutubeIds ...string) ([]*domain.PlaylistItem, error)
	FindVideos(videoIds []string) (map[string]*domain.Video, error)
	// FindVideosInPlaylists returns the videos of every item in the given playlists.
	FindVideosInPlaylists(playlistYoutubeIds ...string) (map[string]*domain.Video, error)
	// ApplyItems replaces the stored items of a playlist with the given ones
	// and saves the videos, in one transaction.
	ApplyItems(playlistYoutubeId string, added []*domain.PlaylistItem, removed []*domain.PlaylistItem, videos []*domain.Video) error
}

type libraryRepository struct {
	db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) *libraryRepository {
	return &libraryRepository{db}
}

func (r *libraryRepository) FindItems(playlistYoutubeIds ...string) ([]*domain.PlaylistItem, error) {
	var items []*domain.PlaylistItem
	err := r.db.Where("playlist_youtube_id IN ?", playlistYoutubeIds).Order("added_at").Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *libraryRepository) FindVideos(videoIds []string) (map[string]*domain.Video, error) {
	videos := map[string]*domain.Video{}
	if len(videoIds) == 0 {
		return videos, nil
	}

	var rows []*domain.Video
	if err := r.db.Where("youtube_id IN ?", videoIds).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, v := range rows {
		videos[v.YoutubeID] = v
	}

	return videos, nil
}

// Filtering with a subquery avoids sending thousands of IDs in an IN clause.
func (r *libraryRepository) FindVideosInPlaylists(playlistYoutubeIds ...string) (map[string]*domain.Video, error) {
	listed := r.db.Model(&domain.PlaylistItem{}).Select("video_youtube_id").Where("playlist_youtube_id IN ?", playlistYoutubeIds)

	return r.findVideosIn(listed)
}

func (r *libraryRepository) findVideosIn(videoIds *gorm.DB) (map[string]*domain.Video, error) {
	var rows []*domain.Video
	if err := r.db.Where("youtube_id IN (?)", videoIds).Find(&rows).Error; err != nil {
		return nil, err
	}

	videos := map[string]*domain.Video{}
	for _, v := range rows {
		videos[v.YoutubeID] = v
	}

	return videos, nil
}

func (r *libraryRepository) ApplyItems(playlistYoutubeId string, added []*domain.PlaylistItem, removed []*domain.PlaylistItem, videos []*domain.Video) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(removed) > 0 {
			if err := tx.Delete(&removed).Error; err != nil {
				return err
			}
		}
		if len(added) > 0 {
			if err := tx.CreateInBatches(added, 500).Error; err != nil {
				return err
			}
		}

		return saveVideos(tx, videos)
	})
}

// Callers pass complete rows (existing values merged with the new ones), so
// every column is overwritten on conflict.
func saveVideos(db *gorm.DB, videos []*domain.Video) error {
	if len(videos) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(videos, 500).Error
}
