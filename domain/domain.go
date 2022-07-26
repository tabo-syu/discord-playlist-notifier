package domain

import "gorm.io/gorm"

type Guild struct {
	gorm.Model
	DiscordID string
	// foreign
	Playlists []Playlist
}

type Playlist struct {
	gorm.Model
	YoutubeID string
	Mention   bool
	Videos    []Video
	// foreign
	GuildID uint
	Guild   Guild
}

type Video struct {
	gorm.Model
	YoutubeID string
	Title     string
	// foreign
	PlaylistID uint
	Playlist   Playlist
}
