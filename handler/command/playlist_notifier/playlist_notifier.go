package playlist_notifier

import (
	cmd "discord-playlist-notifier/handler/command"
	"discord-playlist-notifier/service"

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

	mentionOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        "mention",
		Description: "このサーバーに参加しているユーザー全員にメンションするか否か。",
		Required:    true,
	}
)

func NewPlaylistNotifier(p service.PlaylistService) cmd.Command {
	return &playlistNotifier{
		&discordgo.ApplicationCommand{
			Name:        "playlist-notifier",
			Description: "テキストチャンネルに YouTube のプレイリストの更新を通知します。",
			Options: []*discordgo.ApplicationCommandOption{
				listSubCommand,
				addSubCommand,
				updateSubCommand,
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

func (c *playlistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string) string {
	subcommand := data.Options[0]

	var message string
	switch subcommand.Name {
	case "list":
		message = c.list(guildId)
	case "add":
		options := cmd.ParseArguments(subcommand.Options)
		message = c.add(
			guildId,
			options[playlistIdOption.Name].StringValue(),
			options[mentionOption.Name].BoolValue(),
		)
	case "update":
		options := cmd.ParseArguments(subcommand.Options)
		message = c.update(
			options[playlistIdOption.Name].StringValue(),
			options[mentionOption.Name].BoolValue(),
		)
	case "delete":
		options := cmd.ParseArguments(subcommand.Options)
		message = c.delete(
			guildId,
			options[playlistIdOption.Name].StringValue(),
		)
	case "source":
		message = c.source()
	}

	return message
}
