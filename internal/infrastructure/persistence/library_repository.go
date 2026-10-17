package persistence

import (
	"time"

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

// サブクエリで絞り込むことで、何千もの ID を IN 句で送らずに済む。
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

func (r *libraryRepository) AddSnapshots(snapshots []*library.VideoSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	return r.db.CreateInBatches(snapshots, 500).Error
}

func (r *libraryRepository) FindSnapshots(id library.VideoID, since time.Time) ([]*library.VideoSnapshot, error) {
	var snapshots []*library.VideoSnapshot
	err := r.db.Where("video_youtube_id = ? AND recorded_at >= ?", id, since).Order("recorded_at").Find(&snapshots).Error
	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

func (r *libraryRepository) FindEarliestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundarySnapshots(ids, since, "recorded_at")
}

func (r *libraryRepository) FindLatestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundarySnapshots(ids, since, "recorded_at DESC")
}

// findBoundarySnapshots は各動画の、指定した順で最初の状態を返す。
// (video, recorded_at) のインデックスに対して DISTINCT ON（PostgreSQL）を使う。
func (r *libraryRepository) findBoundarySnapshots(ids []library.VideoID, since time.Time, order string) (map[library.VideoID]*library.VideoSnapshot, error) {
	snapshots := map[library.VideoID]*library.VideoSnapshot{}
	if len(ids) == 0 {
		return snapshots, nil
	}

	var rows []*library.VideoSnapshot
	err := r.db.Select("DISTINCT ON (video_youtube_id) *").
		Where("video_youtube_id IN ? AND recorded_at >= ?", ids, since).
		Order("video_youtube_id, " + order).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, s := range rows {
		snapshots[s.VideoYoutubeID] = s
	}

	return snapshots, nil
}

func (r *libraryRepository) SaveChannels(channels []*library.Channel) error {
	if len(channels) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "updated_at"}),
	}).CreateInBatches(channels, 500).Error
}

func (r *libraryRepository) FindChannels(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	channels := map[library.ChannelID]*library.Channel{}
	if len(ids) == 0 {
		return channels, nil
	}

	var rows []*library.Channel
	if err := r.db.Where("youtube_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, c := range rows {
		channels[c.YoutubeID] = c
	}

	return channels, nil
}

func (r *libraryRepository) SaveCategories(categories []*library.Category) error {
	if len(categories) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "assignable", "updated_at"}),
	}).Create(categories).Error
}

func (r *libraryRepository) FindCategories() (map[library.CategoryID]*library.Category, error) {
	var rows []*library.Category
	if err := r.db.Find(&rows).Error; err != nil {
		return nil, err
	}

	categories := map[library.CategoryID]*library.Category{}
	for _, c := range rows {
		categories[c.YoutubeID] = c
	}

	return categories, nil
}

// 呼び出し側は完全な行（既存の値に新しい値をマージしたもの）を渡すので、
// 競合時には全カラムを上書きする。
func saveVideos(db *gorm.DB, videos []*library.Video) error {
	if len(videos) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(videos, 500).Error
}
