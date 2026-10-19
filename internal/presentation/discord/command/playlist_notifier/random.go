package playlist_notifier

import (
	"errors"

	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

const MAX_RANDOM_COUNT = 5

var (
	minRandomCount = 1.0

	randomCountOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "count",
		Description: "選ぶ本数（1〜5、省略すると 1）",
		MinValue:    &minRandomCount,
		MaxValue:    MAX_RANDOM_COUNT,
	}

	randomSubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "random",
		Description: "通知登録されているプレイリストから、ランダムに動画を選びます。",
		Options: []*discordgo.ApplicationCommandOption{
			randomCountOption,
		},
	}
)

func (c *PlaylistNotifier) random(guildId string, count int) string {
	subscriptions, err := c.subscriptions.FindByGuild(guild.DiscordID(guildId))
	if errors.Is(err, subscription.ErrNotSubscribed) {
		return "通知登録されているプレイリストが存在しません。"
	}

	if err != nil {
		return "エラー！システムに問題があります！"
	}

	var ids []library.PlaylistID
	for _, sub := range subscriptions {
		ids = append(ids, sub.YoutubeID)
	}

	picked, err := c.library.RandomVideos(ids, count)
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	return view.Random(picked)
}
