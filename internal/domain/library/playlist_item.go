package library

import "time"

// PlaylistItem is one entry of a YouTube playlist.
type PlaylistItem struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PlaylistYoutubeID PlaylistID `gorm:"index"`
	ItemID            ItemID     `gorm:"uniqueIndex"`
	VideoYoutubeID    VideoID    `gorm:"index"`
	// When the video was added to the playlist
	AddedAt time.Time
}

func (PlaylistItem) TableName() string {
	return "playlist_items"
}
