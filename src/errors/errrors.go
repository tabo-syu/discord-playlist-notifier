package errors

import "errors"

var (
	ErrNotFoundAtYouTube           = errors.New("not found at youtube")
	ErrNotFoundAtDatabase          = errors.New("not found at database")
	ErrAlreadyRegisteredAtDatabase = errors.New("already registerd at database")
)
