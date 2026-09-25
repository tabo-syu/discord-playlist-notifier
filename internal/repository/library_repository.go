package repository

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LibraryRepository stores the contents of watched YouTube playlists.
type LibraryRepository interface {
	FindItems(playlistYoutubeIds ...string) ([]*domain.PlaylistItem, error)
	FindVideos(videoIds []string) (map[string]*domain.YouTubeVideo, error)
	// FindListedVideoIds returns the IDs of all videos that are in at least one playlist.
	FindListedVideoIds() ([]string, error)
	// ApplyItems replaces the stored items of a playlist with the given ones
	// and saves the videos, in one transaction.
	ApplyItems(playlistYoutubeId string, added []*domain.PlaylistItem, removed []*domain.PlaylistItem, videos []*domain.YouTubeVideo) error
	SaveVideos(videos []*domain.YouTubeVideo) error
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

func (r *libraryRepository) FindVideos(videoIds []string) (map[string]*domain.YouTubeVideo, error) {
	videos := map[string]*domain.YouTubeVideo{}
	if len(videoIds) == 0 {
		return videos, nil
	}

	var rows []*domain.YouTubeVideo
	if err := r.db.Where("youtube_id IN ?", videoIds).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, v := range rows {
		videos[v.YoutubeID] = v
	}

	return videos, nil
}

func (r *libraryRepository) FindListedVideoIds() ([]string, error) {
	var ids []string
	err := r.db.Model(&domain.PlaylistItem{}).Distinct("video_youtube_id").Pluck("video_youtube_id", &ids).Error
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *libraryRepository) ApplyItems(playlistYoutubeId string, added []*domain.PlaylistItem, removed []*domain.PlaylistItem, videos []*domain.YouTubeVideo) error {
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

func (r *libraryRepository) SaveVideos(videos []*domain.YouTubeVideo) error {
	return saveVideos(r.db, videos)
}

// Callers pass complete rows (existing values merged with the new ones), so
// every column is overwritten on conflict.
func saveVideos(db *gorm.DB, videos []*domain.YouTubeVideo) error {
	if len(videos) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(videos, 500).Error
}
