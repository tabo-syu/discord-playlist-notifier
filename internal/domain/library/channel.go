package library

import "time"

// ChannelID is the ID YouTube gives to a channel.
type ChannelID string

// Channel is a YouTube channel that published a video in a watched playlist.
// Videos refer to it by ID. The icon is not stored; notifications fetch it
// right before they are sent.
type Channel struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	YoutubeID ChannelID `gorm:"uniqueIndex"`
	// The name at the last refresh of one of its videos
	Title string
}

func (Channel) TableName() string {
	return "channels"
}
