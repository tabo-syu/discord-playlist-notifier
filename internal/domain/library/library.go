// Package library は通知対象の YouTube プレイリストの中身（アイテムと動画）を
// 持つ。ギルドごとの設定である通知設定とは違い、中身は YouTube の
// プレイリストごとに 1 つだけ保存し、全てのギルドで共有する。
//
// YouTube の ID の型は、人が入力した値（コマンドやツールの引数）からは ParseXxx を
// 通して作る。YouTube API の応答と DB から読んだ値は、発行元の YouTube が正しさを
// 保証しているので、型の変換で作る。
package library

import (
	"regexp"
	"strings"
)

// 形式が公開されていない YouTube の ID に使える文字。長さや接頭辞は ID の種類ごとに
// まちまちなので、文字だけを確かめる。
var idCharsPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// PlaylistID は YouTube のプレイリストの ID。
type PlaylistID string

// ParsePlaylistID は人が入力したプレイリストの ID を読み取る。前後の空白は
// 取り除く。ID に使えない文字があれば ErrInvalidPlaylistID を返す。
func ParsePlaylistID(s string) (PlaylistID, error) {
	s = strings.TrimSpace(s)
	if !idCharsPattern.MatchString(s) {
		return "", ErrInvalidPlaylistID
	}

	return PlaylistID(s), nil
}

// VideoID は YouTube の動画の ID。
type VideoID string

// ItemID はプレイリストのアイテムの ID で、その動画の ID とは異なる。
type ItemID string
