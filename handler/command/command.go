package command

import (
	"discord-playlist-notifier/repository"

	"github.com/bwmarrin/discordgo"
)

type Handle func(*discordgo.ApplicationCommandInteractionData, repository.DBRepository, repository.YouTubeRepository) string
type Command struct {
	Info   *discordgo.ApplicationCommand
	Handle Handle
}

var PlaylistNotifier = Command{
	Info: &discordgo.ApplicationCommand{
		Name:        "playlist-notifier",
		Description: "テキストチャンネルに YouTube のプレイリストの更新を通知します。",
		Options: []*discordgo.ApplicationCommandOption{
			addSubCommand,
			updateSubCommand,
			deleteSubCommand,
			sourceSubCommand,
		},
	},
	Handle: handle,
}

func handle(data *discordgo.ApplicationCommandInteractionData, db repository.DBRepository, youtube repository.YouTubeRepository) string {
	message := ""

	playlistId := ""
	mention := false

	subcommand := data.Options[0]
	if len(subcommand.Options) > 0 {
		options := parseCommandArguments(subcommand.Options)
		playlistId = options["playlist-id"].StringValue()
		mention = options["mention"].BoolValue()
	}
	switch subcommand.Name {
	case "add":
		message = add(db, youtube, playlistId, mention)
	case "update":
		message = update(db, playlistId, mention)
	case "delete":
		message = delete(db, playlistId)
	case "source":
		message = source()
	}

	return message
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

type Options map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseCommandArguments(args []*discordgo.ApplicationCommandInteractionDataOption) Options {
	options := make(Options, len(args))
	for _, arg := range args {
		options[arg.Name] = arg
	}

	return options
}
