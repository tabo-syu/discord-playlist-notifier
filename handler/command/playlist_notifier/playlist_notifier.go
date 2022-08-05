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
	playlistId := ""
	mention := false

	subcommand := data.Options[0]
	if len(subcommand.Options) > 0 {
		options := cmd.ParseArguments(subcommand.Options)
		playlistId = options["playlist-id"].StringValue()
		mention = options["mention"].BoolValue()
	}

	var message string
	switch subcommand.Name {
	case "list":
		message = c.list(guildId)
	case "add":
		message = c.add(guildId, playlistId, mention)
	case "update":
		message = c.update(playlistId, mention)
	case "delete":
		message = c.delete(playlistId)
	case "source":
		message = c.source()
	}

	return message
}
