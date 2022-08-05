package command

import (
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/service"
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

func add(playlist service.PlaylistService, guildId string, playlistId string, needMention bool) string {
	var message string
	switch playlist.Register(guildId, playlistId, needMention) {
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
