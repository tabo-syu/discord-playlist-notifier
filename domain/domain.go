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
	Videos        []Video
	SendChannelID string
	// foreign
	GuildID uint
	Guild   Guild
}

type Video struct {
	gorm.Model
	YoutubeID   string
	Title       string
	PublishedAt time.Time
	// foreign
	PlaylistID uint
	Playlist   Playlist
}
