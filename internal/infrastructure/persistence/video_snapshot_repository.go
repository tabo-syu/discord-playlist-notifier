package persistence

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
)

type videoSnapshotRepository struct {
	db *gorm.DB
}

func NewVideoSnapshotRepository(db *gorm.DB) *videoSnapshotRepository {
	return &videoSnapshotRepository{db}
}

func (r *videoSnapshotRepository) Add(snapshots []*library.VideoSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	return r.db.CreateInBatches(mapSlice(snapshots, toVideoSnapshotRecord), 500).Error
}

func (r *videoSnapshotRepository) Find(id library.VideoID, since time.Time) ([]*library.VideoSnapshot, error) {
	var recs []*videoSnapshotRecord

	err := r.db.Where("video_youtube_id = ? AND recorded_at >= ?", id, since).Order("recorded_at").Find(&recs).Error
	if err != nil {
		return nil, err
	}

	return mapSlice(recs, (*videoSnapshotRecord).toDomain), nil
}

func (r *videoSnapshotRepository) FindEarliest(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundary(ids, since, "recorded_at")
}

func (r *videoSnapshotRepository) FindLatest(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	return r.findBoundary(ids, since, "recorded_at DESC")
}

// findBoundary は各動画の、指定した順で最初の状態を返す。
// (video, recorded_at) のインデックスに対して DISTINCT ON（PostgreSQL）を使う。
func (r *videoSnapshotRepository) findBoundary(ids []library.VideoID, since time.Time, order string) (map[library.VideoID]*library.VideoSnapshot, error) {
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
