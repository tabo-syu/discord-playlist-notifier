package registerer

import (
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/handler/command"

	"github.com/bwmarrin/discordgo"
)

type Registerer interface {
	Register(guildId string) error
	Unregister()
}

type registerer struct {
	session       *discordgo.Session
	commands      []command.Command
	guildCommands map[string][]*discordgo.ApplicationCommand
}

func NewRegisterer(s *discordgo.Session, cs []command.Command) Registerer {
	return &registerer{s, cs, map[string][]*discordgo.ApplicationCommand{}}
}

func (r *registerer) Register(guildId string) error {
	for _, command := range r.commands {
		registered, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, guildId, command.GetCommand())
		if err != nil {
			return errs.ErrDiscordCommandCouldNotCreate
		}

		r.guildCommands[guildId] = append(r.guildCommands[guildId], registered)
	}

	return nil
}

func (r *registerer) Unregister() {
	for guildId, commands := range r.guildCommands {
		for _, command := range commands {
			err := r.session.ApplicationCommandDelete(r.session.State.User.ID, guildId, command.ID)
			if err != nil {
				continue
			}
		}
	}
}
