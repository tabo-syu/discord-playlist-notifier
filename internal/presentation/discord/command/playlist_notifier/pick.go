package playlist_notifier

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Choice value to stop posting, since Discord does not allow an empty value
const PICK_OFF = "off"

var (
	pickIntervalOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "interval",
		Description: "投稿する間隔",
		Required:    true,
		Choices: []*discordgo.ApplicationCommandOptionChoice{
			{Name: "毎日", Value: subscription.PickDaily},
			{Name: "毎週（月曜）", Value: subscription.PickWeekly},
			{Name: "毎月（1 日）", Value: subscription.PickMonthly},
			{Name: "やめる", Value: PICK_OFF},
		},
	}

	pickSubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "pick",
		Description: "プレイリストからランダムに選んだ一曲を、決まった間隔で投稿するか設定します。",
		Options: []*discordgo.ApplicationCommandOption{
			playlistIdOption,
			pickIntervalOption,
		},
	}
)

func (c *PlaylistNotifier) pick(guildId string, playlistId string, choice string) string {
	interval := subscription.PickInterval(choice)
	if choice == PICK_OFF {
		interval = subscription.PickNone
	}

	err := c.subscriptions.SetPickInterval(guild.DiscordID(guildId), library.PlaylistID(playlistId), interval)
	switch {
	case errors.Is(err, domain.ErrDBRecordNotFound):
		return "通知登録されていないプレイリストです。"
	case err != nil:
		return "エラー！システムに問題があります！"
	}

	url := fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistId)
	switch interval {
	case subscription.PickDaily:
		return "毎日 12:00 に「今日の一曲」を投稿します！\n" + url
	case subscription.PickWeekly:
		return "毎週月曜の 12:00 に「今週の一曲」を投稿します！\n" + url
	case subscription.PickMonthly:
		return "毎月 1 日の 12:00 に「今月の一曲」を投稿します！\n" + url
	default:
		return "一曲の投稿をやめました。\n" + url
	}
}
