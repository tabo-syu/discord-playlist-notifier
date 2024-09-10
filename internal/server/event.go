package server

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
)

type event struct {
	guild *service.GuildService
}

func NewEvent(g *service.GuildService) *event {
	return &event{g}
}

func (e *event) GuildCreate(guildId string) error {
	err := e.guild.Register(guildId)
	if err != nil {
		return err
	}

	return nil
}

func (e *event) GuildDelete(guildId string) error {
	err := e.guild.Unregister(guildId)
	if err != nil {
		return err
	}

	return nil
}
