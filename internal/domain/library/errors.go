package library

import "errors"

var (
	// ErrPlaylistNotFound はプレイリストが存在しないか非公開であることを表す。
	ErrPlaylistNotFound = errors.New("playlist not found")
	// ErrVideoNotFound は動画が指定したプレイリストのどれにも含まれていないか、
	// 詳細をまだ取得していないことを表す。
	ErrVideoNotFound = errors.New("video not found")
)
