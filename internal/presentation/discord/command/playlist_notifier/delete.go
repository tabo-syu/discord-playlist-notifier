package playlist_notifier

import (
	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

var deleteSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "delete",
	Description: "通知するプレイリストを削除します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func (c *PlaylistNotifier) delete(guildId string, playlistId string) string {
	id, err := library.ParsePlaylistID(playlistId)
	if err != nil {
		return INVALID_PLAYLIST_ID_MESSAGE
	}

	var message string
	//nolint:errorlint // 完全一致での比較は既知の挙動（CLAUDE.md）
	switch c.subscriptions.Unsubscribe(guild.DiscordID(guildId), id) {
	case nil:
		message = "指定されたプレイリストを削除しました。"
	case subscription.ErrNotSubscribed:
		message = "通知登録されていないプレイリストです。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
