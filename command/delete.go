package command

import (
	"discord-playlist-notifier/service"

	"github.com/bwmarrin/discordgo"
)

var deleteSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "delete",
	Description: "通知するプレイリストを削除します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func delete(service service.YouTubeService, playlistId string) string {
	return "delete"
}
