package domain

import "errors"

var (
	// Discord
	ErrDiscordGeneralError          = errors.New("discord general error")
	ErrDiscordCommandCouldNotCreate = errors.New("command could not be created")
	ErrDiscordCommandCouldNotDelete = errors.New("command could not be deleted")

	// YouTube
	ErrYouTubeGeneralError     = errors.New("youtube general error")
	ErrYouTubePlaylistNotFound = errors.New("playlist not found")

	// DB
	ErrDBGeneralError         = errors.New("db general error")
	ErrDBRecordAlreadyCreated = errors.New("record already created")
	ErrDBRecordNotFound       = errors.New("record not found")
)
