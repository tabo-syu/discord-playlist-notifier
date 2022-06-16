package command

import "github.com/bwmarrin/discordgo"

type Handle func(*discordgo.ApplicationCommandInteractionData) string
type Command struct {
	Info   *discordgo.ApplicationCommand
	Handle Handle
}

var PlaylistNotifier = Command{
	Info: &discordgo.ApplicationCommand{
		Name:        "playlist-notifier",
		Description: "テキストチャンネルに YouTube のプレイリストの更新を通知します。",
		Options: []*discordgo.ApplicationCommandOption{
			add,
			update,
			delete,
			source,
		},
	},
	Handle: handle,
}

func handle(data *discordgo.ApplicationCommandInteractionData) string {
	message := ""

	subcommand := data.Options[0].Name
	switch subcommand {
	case "add":
		message = addFunc("hoge", true)
	case "update":
		message = updateFunc("hoge", true)
	case "delete":
		message = deleteFunc("hoge")
	case "source":
		message = sourceFunc()
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
