package subscription

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

type Repository interface {
	Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error)
	Add(subscription *Subscription) error
	// Update saves every field, including UpdatedAt.
	Update(subscription *Subscription) error
	// UpdatePickInterval saves only the pick interval. It must not touch
	// UpdatedAt, which is the base time of new video notifications.
	UpdatePickInterval(subscription *Subscription) error
	// FindAll returns domain.ErrDBRecordNotFound when there is no subscription.
	FindAll() ([]*Subscription, error)
	// FindByGuild returns domain.ErrDBRecordNotFound when the guild has no subscription.
	FindByGuild(guildID guild.DiscordID) ([]*Subscription, error)
	DeleteAll(subscriptions []*Subscription) error
}
