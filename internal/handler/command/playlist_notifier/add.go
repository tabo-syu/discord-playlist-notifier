package playlist_notifier

import (
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/errs"

	"github.com/bwmarrin/discordgo"
)

var addSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "通知するプレイリストを追加します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func (c *playlistNotifier) add(guildId string, channelId string, playlistId string) string {
	var message string
	switch c.playlist.Register(guildId, channelId, playlistId) {
	case nil:
		message = fmt.Sprintf("通知登録しました！\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case errs.ErrYouTubePlaylistCouldNotFound:
		message = "該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"
	case errs.ErrDBRecordAlreadyCreated:
		message = "既に通知登録されているプレイリストです。"
	case errs.ErrYouTubeGeneralError:
		message = "YouTube API のサービス状況を確認してください。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
