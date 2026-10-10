package library

import "time"

// VideoSnapshot is the state of a video at one point in time. A row is added
// for every video each time the watched videos are refreshed.
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

// NewVideoSnapshot records the current state of the video.
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
