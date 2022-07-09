package command

import (
	"discord-playlist-notifier/repository"

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

func add(repository repository.YouTubeRepository, playlistId string, needMention bool) string {
	res, err := repository.GetPlaylist(playlistId)
	if err != nil {
		return ""
	}

	return res
	// return fmt.Sprintf("%s, %#v", playlistId, needMention)
}
