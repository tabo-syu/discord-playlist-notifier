package playlist_notifier

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
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
	var message string
	switch c.playlist.Unregister(guildId, playlistId) {
	case nil:
		message = "指定されたプレイリストを削除しました。"
	case domain.ErrDBRecordNotFound:
		message = "通知登録されていないプレイリストです。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
