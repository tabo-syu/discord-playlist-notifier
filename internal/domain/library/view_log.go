package library

import "time"

// ViewLog is the view count of a video at one point in time. A row is added
// every time the views of the watched videos are refreshed.
type ViewLog struct {
	ID             uint    `gorm:"primarykey"`
	VideoYoutubeID VideoID `gorm:"index:idx_view_logs_video_recorded,priority:1"`
	Views          ViewCount
	RecordedAt     time.Time `gorm:"index:idx_view_logs_video_recorded,priority:2"`
}

func (ViewLog) TableName() string {
	return "view_logs"
}

// NewViewLog records the current views of the video.
func NewViewLog(v *Video, at time.Time) *ViewLog {
	return &ViewLog{VideoYoutubeID: v.YoutubeID, Views: v.Views, RecordedAt: at}
}
