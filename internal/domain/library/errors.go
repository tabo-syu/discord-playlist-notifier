package library

import "errors"

var (
	// ErrPlaylistNotFound はプレイリストが存在しないか非公開であることを表す。
	ErrPlaylistNotFound = errors.New("playlist not found")
	// ErrVideoNotFound は動画が指定したプレイリストのどれにも含まれていないか、
	// 詳細をまだ取得していないことを表す。
	ErrVideoNotFound = errors.New("video not found")
	// ErrInvalidPlaylistID は入力がプレイリストの ID として読めないことを表す。
	ErrInvalidPlaylistID = errors.New("invalid playlist id")
	// ErrInvalidVideoID は入力が動画の ID とも URL とも読めないことを表す。
	ErrInvalidVideoID = errors.New("invalid video id")
	// ErrInvalidChannelID は入力がチャンネルの ID として読めないことを表す。
	ErrInvalidChannelID = errors.New("invalid channel id")
)
