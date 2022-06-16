package registerer

import (
	"discord-playlist-notifier/command"
	"discord-playlist-notifier/errs"
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
	commands []*command.Command
}

func NewRegisterer(session *discordgo.Session, guild string, commands []*command.Command) *registerer {
	return &registerer{session, guild, commands}
}

func (r *registerer) Register() error {
	for i, command := range r.commands {
		registered, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, r.guild, command.Info)
		if err != nil {
			return errs.ErrCommandCouldNotCreate
		}

		r.commands[i].Info = registered
		fmt.Println("Command Registerd:", command.Info.Name)
	}

	return nil
}

func (r *registerer) Unregister() error {
	for _, command := range r.commands {
		err := r.session.ApplicationCommandDelete(r.session.State.User.ID, r.guild, command.Info.ID)
		if err != nil {
			return errs.ErrCommandCouldNotDelete
		}

		fmt.Println("Command Deleted:", command.Info.Name)
	}

	return nil
}
