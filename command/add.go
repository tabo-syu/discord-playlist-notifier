package command

import "github.com/bwmarrin/discordgo"

var add = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "通知するプレイリストを追加します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
		mentionOption,
	},
}

func addFunc(playlistId string, needMention bool) string {
	return "add"
}
