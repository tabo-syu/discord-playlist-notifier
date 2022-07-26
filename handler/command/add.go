package command

import (
	"discord-playlist-notifier/repository"
	"fmt"

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

func add(db repository.DBRepository, youtube repository.YouTubeRepository, guildId string, playlistId string, needMention bool) string {
	playlists, err := youtube.GetPlaylists(playlistId)
	if err != nil {
		return "error"
	}

	// プレイリストは一度に一個しか登録できない
	playlist := playlists[0]
	playlist.Mention = needMention

	if err = db.AddPlaylist(guildId, playlist); err != nil {
		fmt.Println(err)
		return "Error!"
	}

	return fmt.Sprintf("登録しました！\nhttps://www.youtube.com/playlist?list=%s", playlistId)
}
