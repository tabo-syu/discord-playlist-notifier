// Package fake はテストのために、リポジトリと外部サービスをメモリ上の偽物に
// 差し替える。本物の DB と同じく、読み書きのたびに値をコピーし、呼び出し側が
// 持っているポインタを書き換えても保存した値は変わらないようにする。
//
// テストからだけ使う。本番のコードから import しないこと。
package fake

import (
	"errors"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Err は偽物に失敗させたときに返すエラー。
var Err = errors.New("fake error")

func clone[T any](v *T) *T {
	c := *v
	return &c
}

// 偽物が本物のインターフェースを満たしていることを確かめる
var (
	_ guild.Repository        = (*GuildRepository)(nil)
	_ subscription.Repository = (*SubscriptionRepository)(nil)
	_ library.Repository      = (*LibraryRepository)(nil)
	_ library.YouTube         = (*YouTube)(nil)
	_ application.Notifier    = (*Notifier)(nil)
)
