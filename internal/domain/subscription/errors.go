package subscription

import "errors"

var (
	// ErrAlreadySubscribed はギルドがそのプレイリストをもう通知登録していることを表す。
	ErrAlreadySubscribed = errors.New("playlist is already subscribed")
	// ErrNotSubscribed はギルドがそのプレイリストを（または 1 つも）通知登録して
	// いないことを表す。
	ErrNotSubscribed = errors.New("playlist is not subscribed")
)
