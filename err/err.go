package err

import "errors"

var (
	ErrCommandNotFound = errors.New("command not found")
)
