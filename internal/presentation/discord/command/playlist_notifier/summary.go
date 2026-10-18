package playlist_notifier

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

var summarySubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "summary",
	Description: "通知登録されているプレイリストの曲数、追加ペース、再生数ランキングを表示します。",
	Options: []*discordgo.ApplicationCommandOption{
		optionalPlaylistIdOption,
	},
}

func (c *PlaylistNotifier) summary(guildId string, playlistId string) string {
	subscriptions, err := c.subscriptions.FindByGuild(guild.DiscordID(guildId))
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	var sections []string
	for _, sub := range subscriptions {
		if playlistId != "" && sub.YoutubeID != library.PlaylistID(playlistId) {
			continue
		}

		stats, err := c.library.Stats(sub.YoutubeID)
		if err != nil {
			return "エラー！システムに問題があります！"
		}
		sections = append(sections, view.Summary(sub.Title, stats, c.location))
	}
	if len(sections) == 0 {
		return "通知登録されていないプレイリストです。"
	}

	return view.Truncate(strings.Join(sections, "\n"), view.MAX_MESSAGE_LENGTH)
}
