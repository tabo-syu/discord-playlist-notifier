package errs

import "errors"

var (
	// Discord
	ErrCommandCouldNotCreate = errors.New("command could not create")
	ErrCommandCouldNotDelete = errors.New("command could not delete")

	// YouTube
	ErrPlaylistCouldNotFound = errors.New("playlist could not found")

	// DB
	ErrRecordAlreadyCreated = errors.New("record already created")
	ErrRecordCouldNotFound  = errors.New("record could not found")
)
