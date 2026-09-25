package domain

import (
	"time"

	"gorm.io/gorm"
)

type Guild struct {
	gorm.Model
	DiscordID string
	// foreign
	Playlists []Playlist
}

type Playlist struct {
	gorm.Model
	YoutubeID     string
	SendChannelID string
	Title         string
	// foreign
	GuildID uint
	Guild   Guild
	Videos  []Video
}

type Video struct {
	gorm.Model
	YoutubeID        string
	Title            string
	Views            uint64
	Thumbnail        string
	ChannelName      string
	ChannelIcon      string
	PublishedAt      time.Time
	OwnerPublishedAt time.Time
	// foreign
	PlaylistID uint
	Playlist   Playlist
}

// Privacy statuses of a YouTube video as seen through a playlist item.
const (
	PrivacyPublic   = "public"
	PrivacyUnlisted = "unlisted"
	PrivacyPrivate  = "private"
	PrivacyDeleted  = "deleted"
)

// PlaylistItem is one entry of a YouTube playlist. Unlike Playlist, which is a
// per-guild notification setting, items are stored once per YouTube playlist.
type PlaylistItem struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PlaylistYoutubeID string `gorm:"index"`
	ItemID            string `gorm:"uniqueIndex"`
	VideoYoutubeID    string `gorm:"index"`
	// When the video was added to the playlist
	AddedAt time.Time
}

// YouTubeVideo holds the latest known state of a video that appears in a
// watched playlist. Details (title, views, ...) are kept after the video
// becomes private or deleted, so they can still be shown.
type YouTubeVideo struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	YoutubeID     string `gorm:"uniqueIndex"`
	Title         string
	ChannelID     string
	ChannelName   string
	ChannelIcon   string
	Thumbnail     string
	Duration      time.Duration
	Views         uint64
	PrivacyStatus string
	PublishedAt   time.Time
}

func (YouTubeVideo) TableName() string {
	return "youtube_videos"
}

// HasDetails reports whether the video details have been fetched at least once.
func (v *YouTubeVideo) HasDetails() bool {
	return v.Title != ""
}

// Available reports whether the video can be watched and shown to users.
func (v *YouTubeVideo) Available() bool {
	return v.HasDetails() && (v.PrivacyStatus == PrivacyPublic || v.PrivacyStatus == PrivacyUnlisted)
}
