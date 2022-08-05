package errs

import "errors"

var (
	// Discord
	ErrDiscordGeneralError          = errors.New("discord general error")
	ErrDiscordCommandCouldNotCreate = errors.New("command could not create")
	ErrDiscordCommandCouldNotDelete = errors.New("command could not delete")

	// YouTube
	ErrYouTubeGeneralError          = errors.New("youtube general error")
	ErrYouTubePlaylistCouldNotFound = errors.New("playlist could not found")

	// DB
	ErrDBGeneralError         = errors.New("db general error")
	ErrDBRecordAlreadyCreated = errors.New("record already created")
	ErrDBRecordCouldNotFound  = errors.New("record could not found")
)
