package command

import (
	"discord-playlist-notifier/repository"

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

func delete(repository repository.YouTubeRepository, playlistId string) string {
	return "delete"
}
