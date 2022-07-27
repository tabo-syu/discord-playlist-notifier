package command

import (
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/repository"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var listSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "list",
	Description: "通知するプレイリストを一覧表示します。",
}

func list(db repository.DBRepository, guildId string) string {
	playlists, err := db.GetPlaylistsByDiscordId(guildId)
	if errors.Is(err, errs.ErrRecordCouldNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	message := "通知登録されているプレイリスト一覧\n"
	for _, playlist := range *playlists {
		message += fmt.Sprintf("https://www.youtube.com/playlist?list=%s\n", playlist.YoutubeID)
	}

	return message
}
