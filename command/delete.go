package command

import "github.com/bwmarrin/discordgo"

var delete = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "delete",
	Description: "通知するプレイリストを削除します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func deleteFunc(playlistId string) string {
	return "delete"
}
