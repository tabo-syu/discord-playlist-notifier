package playlist_notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/server/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

type playlistNotifier struct {
	command  *discordgo.ApplicationCommand
	playlist service.PlaylistService
}

var (
	playlistIdOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "playlist-id",
		Description: "YouTube のプレイリストページの URL 末尾に付く ID を入力します。",
		Required:    true,
	}
)

func NewPlaylistNotifier(p service.PlaylistService) command.Command {
	return &playlistNotifier{
		&discordgo.ApplicationCommand{
			Name:        "playlist-notifier",
			Description: "テキストチャンネルに YouTube のプレイリストの更新を通知します。",
			Options: []*discordgo.ApplicationCommandOption{
				listSubCommand,
				addSubCommand,
				deleteSubCommand,
				sourceSubCommand,
			},
		},
		p,
	}
}

func (c *playlistNotifier) GetCommand() *discordgo.ApplicationCommand {
	return c.command
}

func (c *playlistNotifier) SetCommand(cmd *discordgo.ApplicationCommand) {
	c.command = cmd
}

func (c *playlistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string {
	subcommand := data.Options[0]

	var message string
	switch subcommand.Name {
	case listSubCommand.Name:
		message = c.list(guildId)
	case addSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		message = c.add(
			guildId,
			channelId,
			options[playlistIdOption.Name].StringValue(),
		)
	case deleteSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		message = c.delete(
			guildId,
			options[playlistIdOption.Name].StringValue(),
		)
	case sourceSubCommand.Name:
		message = c.source()
	}

	return message
}
