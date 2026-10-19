package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *videoRepository {
	return &videoRepository{db}
}

func (r *videoRepository) Find(ids []library.VideoID) (map[library.VideoID]*library.Video, error) {
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

// サブクエリで絞り込むことで、何千もの ID を IN 句で送らずに済む。
func (r *videoRepository) FindInPlaylists(playlistIDs ...library.PlaylistID) (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&playlistItemRecord{}).Select("video_youtube_id").Where("playlist_youtube_id IN ?", playlistIDs)

	return r.findIn(listed)
}

func (r *videoRepository) FindListed() (map[library.VideoID]*library.Video, error) {
	listed := r.db.Model(&playlistItemRecord{}).Select("video_youtube_id")

	return r.findIn(listed)
}

// 呼び出し側は完全な行（既存の値に新しい値をマージしたもの）を渡すので、
// 競合時には全カラムを上書きする。ただし GORM は UpdateAll でも id と created_at を
// 上書きしないので、ドメインのモデルが持たないこの 2 つは既存の値のまま残る。
func (r *videoRepository) Save(videos []*library.Video) error {
	if len(videos) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		UpdateAll: true,
	}).CreateInBatches(mapSlice(videos, toVideoRecord), 500).Error
}

func (r *videoRepository) findIn(ids *gorm.DB) (map[library.VideoID]*library.Video, error) {
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
