package playlist_notifier

import (
	"errors"
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"github.com/bwmarrin/discordgo"
)

var listSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "list",
	Description: "通知するプレイリストを一覧表示します。",
}

func (c *PlaylistNotifier) list(guildId string) string {
	playlists, err := c.playlist.FindByGuild(guildId)
	if errors.Is(err, domain.ErrDBRecordCouldNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	message := "通知登録されているプレイリスト一覧\n"
	for _, playlist := range playlists {
		message += fmt.Sprintf("https://www.youtube.com/playlist?list=%s\n", playlist.YoutubeID)
	}

	return message
}
