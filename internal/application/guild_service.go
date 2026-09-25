// Package application holds the use cases of the bot. They load aggregates
// through the repositories, let the domain decide, and save the results.
package application

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

type GuildService struct {
	guilds        guild.Repository
	subscriptions subscription.Repository
}

func NewGuildService(g guild.Repository, s subscription.Repository) *GuildService {
	return &GuildService{g, s}
}

// Register records a guild the bot has joined.
func (s *GuildService) Register(id guild.DiscordID) error {
	exists, err := s.guilds.Exists(id)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrDBRecordAlreadyCreated
	}

	return s.guilds.Add(guild.New(id))
}

// Unregister deletes a guild the bot has left, together with its subscriptions.
func (s *GuildService) Unregister(id guild.DiscordID) error {
	exists, err := s.guilds.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrDBRecordNotFound
	}

	g, err := s.guilds.FindByDiscordID(id)
	if err != nil {
		return err
	}
	subscriptions, err := s.subscriptions.FindByGuild(id)
	if err != nil {
		return err
	}

	err = s.guilds.Delete(g)
	if err != nil {
		return err
	}

	return s.subscriptions.DeleteAll(subscriptions)
}
