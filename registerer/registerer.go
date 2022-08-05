package registerer

import (
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/handler/command"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Registerer interface {
	Register() error
	Unregister() error
}

type registerer struct {
	session  *discordgo.Session
	guild    string
	commands []command.Command
}

func NewRegisterer(session *discordgo.Session, guild string, commands []command.Command) Registerer {
	return &registerer{session, guild, commands}
}

func (r *registerer) Register() error {
	for i, command := range r.commands {
		registered, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, r.guild, command.GetCommand())
		if err != nil {
			return errs.ErrDiscordCommandCouldNotCreate
		}

		r.commands[i].SetCommand(registered)
		fmt.Println("Command Registerd:", command.GetCommand().Name)
	}

	return nil
}

func (r *registerer) Unregister() error {
	for _, command := range r.commands {
		err := r.session.ApplicationCommandDelete(r.session.State.User.ID, r.guild, command.GetCommand().ID)
		if err != nil {
			return errs.ErrDiscordCommandCouldNotDelete
		}

		fmt.Println("Command Deleted:", command.GetCommand().Name)
	}

	return nil
}
