package errs

import "errors"

var (
	ErrCommandCouldNotCreate = errors.New("command could not create")
	ErrCommandCouldNotDelete = errors.New("command could not delete")
)
