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

func add(db repository.DBRepository, youtube repository.YouTubeRepository, playlistId string, needMention bool) string {
	playlists, err := youtube.GetPlaylists(playlistId)
	if err != nil {
		return "error"
	}

	// プレイリストは一度に一個しか登録できない
	playlist := playlists[0]
	playlist.Mention = needMention

	return "hoge"
	// return fmt.Sprintf("%s, %#v", playlistId, needMention)
}
