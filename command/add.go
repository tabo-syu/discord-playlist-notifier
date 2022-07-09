package command

import (
	"discord-playlist-notifier/service"

	"github.com/bwmarrin/discordgo"
)

var addSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "通知するプレイリストを追加します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
		mentionOption,
	},
}

func add(service service.YouTubeService, playlistId string, needMention bool) string {
	res, err := service.GetPlaylist(playlistId)
	if err != nil {
		return ""
	}

	return res
	// return fmt.Sprintf("%s, %#v", playlistId, needMention)
}
