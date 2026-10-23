package discord

import (
	"sync"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"

	"github.com/bwmarrin/discordgo"
)

type guildID string

type registrar struct {
	session       *discordgo.Session
	commands      []command.Command
	guildCommands map[guildID][]*discordgo.ApplicationCommand
	// GuildCreate events are handled concurrently, so guildCommands is guarded
	mu sync.Mutex
}

func NewRegisterer(s *discordgo.Session, cs []command.Command) *registrar {
	return &registrar{s, cs, map[guildID][]*discordgo.ApplicationCommand{}, sync.Mutex{}}
}

func (r *registrar) Register(guildId string) error {
	for _, command := range r.commands {
		registered, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, guildId, command.GetCommand())
		if err != nil {
			return domain.ErrDiscordCommandCouldNotCreate
		}

		r.mu.Lock()
		r.guildCommands[guildID(guildId)] = append(r.guildCommands[guildID(guildId)], registered)
		r.mu.Unlock()
	}

	return nil
}

func (r *registrar) Unregister() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for guildId, commands := range r.guildCommands {
		for _, command := range commands {
			err := r.session.ApplicationCommandDelete(r.session.State.User.ID, string(guildId), command.ID)
			if err != nil {
				continue
			}
		}
	}
}
