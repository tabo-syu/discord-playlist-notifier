package command

import (
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/repository"
	"errors"
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
	if errors.Is(err, errs.ErrPlaylistCouldNotFound) {
		return "該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"
	}
	if err != nil {
		return "YouTube API のサービス状況を確認してください。"
	}

	// プレイリストは一度に一個しか登録できない
	playlist := playlists[0]
	playlist.Mention = needMention

	var message string
	switch db.AddPlaylist(guildId, playlist) {
	case nil:
		message = fmt.Sprintf("通知登録しました！\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case errs.ErrRecordAlreadyCreated:
		message = "既に通知登録されているプレイリストです。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
