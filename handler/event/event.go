package event

import (
	"discord-playlist-notifier/service"
)

type Event interface {
	GuildCreate(guildId string) error
	GuildDelete(guildId string) error
}

type event struct {
	guild service.GuildService
}

func NewEvent(g service.GuildService) Event {
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
