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
	var recs []*playlistItemRecord

	err := r.db.Where("playlist_youtube_id IN ?", playlistIDs).Order("added_at").Find(&recs).Error
	if err != nil {
		return nil, err
	}

	return mapSlice(recs, (*playlistItemRecord).toDomain), nil
}

func (r *libraryRepository) FindVideos(ids []library.VideoID) (map[library.VideoID]*library.Video, error) {
	videos := map[library.VideoID]*library.Video{}
	if len(ids) == 0 {
		return videos, nil
	}

	var recs []*videoRecord
	if err := r.db.Where("youtube_id IN ?", ids).Find(&recs).Error; err != nil {
		return nil, err
	}

	for _, rec := range recs {
		videos[library.VideoID(rec.YoutubeID)] = rec.toDomain()
	}

	return videos, nil
}

func (r *libraryRepository) FindItemsByVideos(ids []library.VideoID) ([]*library.PlaylistItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var recs []*playlistItemRecord
	if err := r.db.Where("video_youtube_id IN ?", ids).Find(&recs).Error; err != nil {
		return nil, err
	}

	return mapSlice(recs, (*playlistItemRecord).toDomain), nil
}

// サブクエリで絞り込むことで、何千もの ID を IN 句で送らずに済む。
func (r *libraryRepository) FindVideosInPlaylists(playlistIDs ...library.PlaylistID) (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&playlistItemRecord{}).Select("video_youtube_id").Where("playlist_youtube_id IN ?", playlistIDs)

	return r.findVideosIn(listed)
}

func (r *libraryRepository) FindListedVideos() (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&playlistItemRecord{}).Select("video_youtube_id")

	return r.findVideosIn(listed)
}

// ApplyItems は外れたアイテムを、YouTube のアイテム ID（一意）で削除する。
func (r *libraryRepository) ApplyItems(playlistID library.PlaylistID, added []*library.PlaylistItem, removed []*library.PlaylistItem, videos []*library.Video) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(removed) > 0 {
			itemIDs := make([]string, 0, len(removed))
			for _, i := range removed {
				itemIDs = append(itemIDs, string(i.ItemID))
			}

			if err := tx.Where("item_id IN ?", itemIDs).Delete(&playlistItemRecord{}).Error; err != nil {
				return err
			}
		}

		if len(added) > 0 {
			if err := tx.CreateInBatches(mapSlice(added, toPlaylistItemRecord), 500).Error; err != nil {
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

	return r.db.CreateInBatches(mapSlice(snapshots, toVideoSnapshotRecord), 500).Error
}

func (r *libraryRepository) FindSnapshots(id library.VideoID, since time.Time) ([]*library.VideoSnapshot, error) {
	var recs []*videoSnapshotRecord

	err := r.db.Where("video_youtube_id = ? AND recorded_at >= ?", id, since).Order("recorded_at").Find(&recs).Error
	if err != nil {
		return nil, err
	}

	return mapSlice(recs, (*videoSnapshotRecord).toDomain), nil
}

func (r *libraryRepository) FindEarliestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundarySnapshots(ids, since, "recorded_at")
}

func (r *libraryRepository) FindLatestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundarySnapshots(ids, since, "recorded_at DESC")
}

func (r *libraryRepository) SaveChannels(channels []*library.Channel) error {
	if len(channels) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "updated_at"}),
	}).CreateInBatches(mapSlice(channels, toChannelRecord), 500).Error
}

func (r *libraryRepository) FindChannels(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	channels := map[library.ChannelID]*library.Channel{}
	if len(ids) == 0 {
		return channels, nil
	}

	var recs []*channelRecord
	if err := r.db.Where("youtube_id IN ?", ids).Find(&recs).Error; err != nil {
		return nil, err
	}

	for _, rec := range recs {
		channels[library.ChannelID(rec.YoutubeID)] = rec.toDomain()
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
	}).Create(mapSlice(categories, toCategoryRecord)).Error
}

func (r *libraryRepository) FindCategories() (map[library.CategoryID]*library.Category, error) {
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

func (r *libraryRepository) findVideosIn(ids *gorm.DB) (map[library.VideoID]*library.Video, error) {
	var recs []*videoRecord
	if err := r.db.Where("youtube_id IN (?)", ids).Find(&recs).Error; err != nil {
		return nil, err
	}

	videos := map[library.VideoID]*library.Video{}
	for _, rec := range recs {
		videos[library.VideoID(rec.YoutubeID)] = rec.toDomain()
	}

	return videos, nil
}

// findBoundarySnapshots は各動画の、指定した順で最初の状態を返す。
// (video, recorded_at) のインデックスに対して DISTINCT ON（PostgreSQL）を使う。
func (r *libraryRepository) findBoundarySnapshots(ids []library.VideoID, since time.Time, order string) (map[library.VideoID]*library.VideoSnapshot, error) {
	snapshots := map[library.VideoID]*library.VideoSnapshot{}
	if len(ids) == 0 {
		return snapshots, nil
	}

	var recs []*videoSnapshotRecord

	err := r.db.Select("DISTINCT ON (video_youtube_id) *").
		Where("video_youtube_id IN ? AND recorded_at >= ?", ids, since).
		Order("video_youtube_id, " + order).
		Find(&recs).Error
	if err != nil {
		return nil, err
	}

	for _, rec := range recs {
		snapshots[library.VideoID(rec.VideoYoutubeID)] = rec.toDomain()
	}

	return snapshots, nil
}

// 呼び出し側は完全な行（既存の値に新しい値をマージしたもの）を渡すので、
// 競合時には全カラムを上書きする。ただし GORM は UpdateAll でも id と created_at を
// 上書きしないので、ドメインのモデルが持たないこの 2 つは既存の値のまま残る。
func saveVideos(db *gorm.DB, videos []*library.Video) error {
	if len(videos) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(mapSlice(videos, toVideoRecord), 500).Error
}
