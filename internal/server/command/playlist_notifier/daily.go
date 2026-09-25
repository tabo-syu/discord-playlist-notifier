package playlist_notifier

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

var (
	dailyEnabledOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        "enabled",
		Description: "True で毎日投稿します。False で投稿をやめます。",
		Required:    true,
	}

	dailySubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "daily",
		Description: "プレイリストからランダムに選んだ「今日の一曲」を毎日投稿するか設定します。",
		Options: []*discordgo.ApplicationCommandOption{
			playlistIdOption,
			dailyEnabledOption,
		},
	}
)

func (c *PlaylistNotifier) daily(guildId string, playlistId string, enabled bool) string {
	err := c.playlist.SetDailyPick(guildId, playlistId, enabled)
	switch {
	case err == nil && enabled:
		return fmt.Sprintf("毎日「今日の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case err == nil:
		return fmt.Sprintf("「今日の一曲」の投稿をやめました。\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case errors.Is(err, domain.ErrDBRecordNotFound):
		return "通知登録されていないプレイリストです。"
	default:
		return "エラー！システムに問題があります！"
	}
}
