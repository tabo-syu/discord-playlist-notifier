package subscription

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

type Repository interface {
	// Add は、同じギルドに同じプレイリストの通知設定（論理削除したものを除く）が
	// あれば保存せず ErrAlreadySubscribed を返す。同時に呼んでも 1 つしか保存しない。
	// 「同じギルドに同じプレイリストは 1 つだけ」は 1 つの通知設定の中では確かめられない
	// （集合に対する）条件なので、集約ではなくリポジトリが守る。
	Add(subscription *Subscription) error
	// Update は通知設定の全てのフィールドを保存する。保存していないか、論理削除
	// した通知設定には ErrNotSubscribed を返す。
	Update(subscription *Subscription) error
	// FindByGuildAndPlaylist は、ギルドとプレイリストの組で通知設定を 1 つ引く。
	// 通知設定がないときは ErrNotSubscribed を返す。
	FindByGuildAndPlaylist(guildID guild.DiscordID, playlistID library.PlaylistID) (*Subscription, error)
	// FindAll は通知設定が 1 つもないときに ErrNotSubscribed を返す。
	FindAll() ([]*Subscription, error)
	// FindByGuild はギルドに通知設定がないときに ErrNotSubscribed を返す。
	FindByGuild(guildID guild.DiscordID) ([]*Subscription, error)
	DeleteAll(subscriptions []*Subscription) error
}
