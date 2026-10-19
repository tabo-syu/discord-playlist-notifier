package playlist_notifier

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
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
	id, err := library.ParsePlaylistID(playlistId)
	if err != nil {
		return INVALID_PLAYLIST_ID_MESSAGE
	}

	var message string
	//nolint:errorlint // 完全一致での比較は既知の挙動（CLAUDE.md）
	switch c.subscriptions.Subscribe(guild.DiscordID(guildId), subscription.ChannelID(channelId), id) {
	case nil:
		message = fmt.Sprintf("通知登録しました！\nhttps://www.youtube.com/playlist?list=%s", id)
	case library.ErrPlaylistNotFound:
		message = "該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"
	case subscription.ErrAlreadySubscribed:
		message = "既に通知登録されているプレイリストです。"
	case application.ErrYouTubeUnavailable:
		message = "YouTube API のサービス状況を確認してください。"
	default:
		message = "エラー！システムに問題があります！"
	}

	return message
}
