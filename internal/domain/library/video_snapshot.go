package library

import "time"

// VideoSnapshot はある時点での動画の状態。通知対象の動画を更新するたびに、
// 全ての動画について 1 行ずつ追加する。
type VideoSnapshot struct {
	ID             uint      `gorm:"primarykey"`
	VideoYoutubeID VideoID   `gorm:"index:idx_video_snapshots_video_recorded,priority:1"`
	RecordedAt     time.Time `gorm:"index:idx_video_snapshots_video_recorded,priority:2"`
	Title          string
	PrivacyStatus  PrivacyStatus
	Views          ViewCount
	Likes          uint64
	Comments       uint64
}

func (VideoSnapshot) TableName() string {
	return "video_snapshots"
}

// NewVideoSnapshot は動画の現在の状態を記録する。
func NewVideoSnapshot(v *Video, at time.Time) *VideoSnapshot {
	return &VideoSnapshot{
		VideoYoutubeID: v.YoutubeID,
		RecordedAt:     at,
		Title:          v.Title,
		PrivacyStatus:  v.PrivacyStatus,
		Views:          v.Views,
		Likes:          v.Likes,
		Comments:       v.Comments,
	}
}

// VideoHistory は動画と、記録された状態（古い順）。
type VideoHistory struct {
	Video     *Video
	Snapshots []*VideoSnapshot
}
