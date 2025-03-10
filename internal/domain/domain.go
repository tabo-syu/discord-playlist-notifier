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
	// Information about who added the video to the playlist
	AddedByChannelID   string
	AddedByChannelName string
	AddedByChannelIcon string
	// foreign
	PlaylistID uint
	Playlist   Playlist
}
