package playlist_notifier

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

var addSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "通知するプレイリストを追加します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func (c *PlaylistNotifier) add(guildId string, channelId string, playlistId string) string {
	var message string
	switch c.playlist.Register(guildId, channelId, playlistId) {
	case nil:
		message = fmt.Sprintf("通知登録しました！\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case domain.ErrYouTubePlaylistCouldNotFound:
		message = "該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"
	case domain.ErrDBRecordAlreadyCreated:
		message = "既に通知登録されているプレイリストです。"
	case domain.ErrYouTubeGeneralError:
		message = "YouTube API のサービス状況を確認してください。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
