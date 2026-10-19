package guild

import "errors"

var (
	// ErrAlreadyRegistered はボットが参加したギルドがもう記録されていることを表す。
	ErrAlreadyRegistered = errors.New("guild is already registered")
	// ErrNotRegistered はギルドが記録されていないことを表す。
	ErrNotRegistered = errors.New("guild is not registered")
)
