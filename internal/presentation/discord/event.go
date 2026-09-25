package discord

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
)

type event struct {
	guild *application.GuildService
}

func NewEvent(g *application.GuildService) *event {
	return &event{g}
}

func (e *event) GuildCreate(guildId string) error {
	err := e.guild.Register(guild.DiscordID(guildId))
	if err != nil {
		return err
	}

	return nil
}

func (e *event) GuildDelete(guildId string) error {
	err := e.guild.Unregister(guild.DiscordID(guildId))
	if err != nil {
		return err
	}

	return nil
}
