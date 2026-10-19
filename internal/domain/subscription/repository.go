package subscription

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

type Repository interface {
	Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error)
	Add(subscription *Subscription) error
	// Update は通知設定の全てのフィールドを保存する。
	Update(subscription *Subscription) error
	// FindAll は通知設定が 1 つもないときに domain.ErrDBRecordNotFound を返す。
	FindAll() ([]*Subscription, error)
	// FindByGuild はギルドに通知設定がないときに domain.ErrDBRecordNotFound を返す。
	FindByGuild(guildID guild.DiscordID) ([]*Subscription, error)
	DeleteAll(subscriptions []*Subscription) error
}
