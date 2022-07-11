package domain

import "gorm.io/gorm"

type Guild struct {
	gorm.Model
	DiscordID string
	Playlists []Playlist
}

type Playlist struct {
	gorm.Model
	YoutubeID string
	Mention   bool
	Videos    []Video
	// foreign key
	GuildID uint
}

type Video struct {
	gorm.Model
	YoutubeID string
	Title     string
	// foreign key
	PlaylistID uint
}
