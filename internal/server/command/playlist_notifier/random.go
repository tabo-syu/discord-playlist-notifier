package playlist_notifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
)

const MAX_RANDOM_COUNT = 5

var (
	minRandomCount = 1.0

	randomCountOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "count",
		Description: "選ぶ曲数（1〜5、省略すると 1）",
		MinValue:    &minRandomCount,
		MaxValue:    MAX_RANDOM_COUNT,
	}

	randomSubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "random",
		Description: "通知登録されているプレイリストから、ランダムに曲を選びます。",
		Options: []*discordgo.ApplicationCommandOption{
			randomCountOption,
		},
	}
)

func (c *PlaylistNotifier) random(guildId string, count int) string {
	playlists, err := c.playlist.FindByGuild(guildId)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	var ids []string
	for _, playlist := range playlists {
		ids = append(ids, playlist.YoutubeID)
	}
	picked, err := c.library.RandomVideos(ids, count)
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	return formatRandom(picked)
}

func formatRandom(picked []*service.PlaylistVideo) string {
	if len(picked) == 0 {
		return "選べる曲がまだありません。少し待ってからもう一度試してください。"
	}

	var sb strings.Builder
	sb.WriteString("🎲 ランダムに選びました！\n")
	for _, p := range picked {
		fmt.Fprintf(&sb, "\n**%s**\nhttps://www.youtube.com/watch?v=%s&list=%s\n", p.Video.Title, p.Video.YoutubeID, p.PlaylistID)
	}

	return sb.String()
}
