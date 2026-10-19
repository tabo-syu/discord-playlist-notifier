package subscription

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

type Repository interface {
	Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error)
	Add(subscription *Subscription) error
	// Update は通知設定の全てのフィールドを保存する。保存していないか、論理削除
	// した通知設定には ErrNotSubscribed を返す。
	Update(subscription *Subscription) error
	// FindAll は通知設定が 1 つもないときに ErrNotSubscribed を返す。
	FindAll() ([]*Subscription, error)
	// FindByGuild はギルドに通知設定がないときに ErrNotSubscribed を返す。
	FindByGuild(guildID guild.DiscordID) ([]*Subscription, error)
	DeleteAll(subscriptions []*Subscription) error
}
