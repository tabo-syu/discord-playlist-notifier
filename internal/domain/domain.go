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

// Video holds the latest known state of a video that appears in a watched
// playlist, once per video. Details (title, views, ...) are kept after the
// video becomes private or deleted, so they can still be shown.
type Video struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	YoutubeID     string `gorm:"uniqueIndex"`
	Title         string
	Views         uint64
	PrivacyStatus string
	// When the video itself was published
	PublishedAt time.Time
	// The highest view milestone already reached. Nil until the first time
	// the views are known, so that existing views are not notified.
	ViewMilestone *uint64
}

// HasDetails reports whether the video details have been fetched at least once.
func (v *Video) HasDetails() bool {
	return v.Title != ""
}

// Available reports whether the video can be watched and shown to users.
func (v *Video) Available() bool {
	return v.HasDetails() && (v.PrivacyStatus == PrivacyPublic || v.PrivacyStatus == PrivacyUnlisted)
}

// Thumbnail returns the URL of the video thumbnail, which YouTube derives from the video ID.
func (v *Video) Thumbnail() string {
	return "https://i.ytimg.com/vi/" + v.YoutubeID + "/hqdefault.jpg"
}

// LiveVideo is information fetched from YouTube right before a video is
// shown. The channel is not stored, and the views are fresher than the stored ones.
type LiveVideo struct {
	ChannelName string
	ChannelIcon string
	Views       uint64
}
