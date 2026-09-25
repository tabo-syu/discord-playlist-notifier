package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type libraryRepository struct {
	db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) *libraryRepository {
	return &libraryRepository{db}
}

func (r *libraryRepository) FindItems(playlistIDs ...library.PlaylistID) ([]*library.PlaylistItem, error) {
	var items []*library.PlaylistItem
	err := r.db.Where("playlist_youtube_id IN ?", playlistIDs).Order("added_at").Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *libraryRepository) FindVideos(ids []library.VideoID) (map[library.VideoID]*library.Video, error) {
	videos := map[library.VideoID]*library.Video{}
	if len(ids) == 0 {
		return videos, nil
	}

	var rows []*library.Video
	if err := r.db.Where("youtube_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, v := range rows {
		videos[v.YoutubeID] = v
	}

	return videos, nil
}

func (r *libraryRepository) FindItemsByVideos(ids []library.VideoID) ([]*library.PlaylistItem, error) {
	var items []*library.PlaylistItem
	if len(ids) == 0 {
		return items, nil
	}
	if err := r.db.Where("video_youtube_id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// Filtering with a subquery avoids sending thousands of IDs in an IN clause.
func (r *libraryRepository) FindVideosInPlaylists(playlistIDs ...library.PlaylistID) (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&library.PlaylistItem{}).Select("video_youtube_id").Where("playlist_youtube_id IN ?", playlistIDs)

	return r.findVideosIn(listed)
}

func (r *libraryRepository) FindListedVideos() (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&library.PlaylistItem{}).Select("video_youtube_id")

	return r.findVideosIn(listed)
}

func (r *libraryRepository) findVideosIn(ids *gorm.DB) (map[library.VideoID]*library.Video, error) {
	var rows []*library.Video
	if err := r.db.Where("youtube_id IN (?)", ids).Find(&rows).Error; err != nil {
		return nil, err
	}

	videos := map[library.VideoID]*library.Video{}
	for _, v := range rows {
		videos[v.YoutubeID] = v
	}

	return videos, nil
}

func (r *libraryRepository) ApplyItems(playlistID library.PlaylistID, added []*library.PlaylistItem, removed []*library.PlaylistItem, videos []*library.Video) error {
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

func (r *libraryRepository) SaveVideos(videos []*library.Video) error {
	return saveVideos(r.db, videos)
}

// Callers pass complete rows (existing values merged with the new ones), so
// every column is overwritten on conflict.
func saveVideos(db *gorm.DB, videos []*library.Video) error {
	if len(videos) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(videos, 500).Error
}
