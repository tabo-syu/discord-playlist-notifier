package playlist_notifier

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
	"github.com/tabo-syu/discord-playlist-notifier/internal/view"
)

var statsSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "stats",
	Description: "通知登録されているプレイリストの統計を表示します。",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        playlistIdOption.Name,
			Description: "統計を見るプレイリストの ID（省略するとすべて）",
		},
	},
}

func (c *PlaylistNotifier) stats(guildId string, playlistId string) string {
	playlists, err := c.playlist.FindByGuild(guildId)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}

	var sections []string
	for _, playlist := range playlists {
		if playlistId != "" && playlist.YoutubeID != playlistId {
			continue
		}

		contents, err := c.library.Contents([]string{playlist.YoutubeID})
		if err != nil {
			return "エラー！システムに問題があります！"
		}
		stats := service.ComputeStats(contents, time.Now(), c.location)
		sections = append(sections, formatStats(playlist, stats, c.location))
	}
	if len(sections) == 0 {
		return "通知登録されていないプレイリストです。"
	}

	return view.Truncate(strings.Join(sections, "\n"), view.MAX_MESSAGE_LENGTH)
}

func formatStats(playlist *domain.Playlist, s *service.PlaylistStats, loc *time.Location) string {
	if s.Total == 0 {
		return fmt.Sprintf("📊 **%s**\nまだ曲の情報がありません。少し待ってからもう一度試してください。\n", playlist.Title)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "📊 **%s**\n", playlist.Title)
	fmt.Fprintf(&sb, "曲数: %s 曲（観られる %s / 非公開 %s / 削除 %s）\n", view.Number(s.Total), view.Number(s.Available), view.Number(s.Private), view.Number(s.Deleted))
	fmt.Fprintf(&sb, "追加ペース: 今月 %d 曲 / 過去 30 日 %d 曲\n", s.AddedThisMonth, s.AddedLast30Days)
	fmt.Fprintf(&sb, "最初の追加: %s\n", s.FirstAddedAt.In(loc).Format("2006/01/02"))

	if len(s.TopViewed) > 0 {
		sb.WriteString("\n**再生数ランキング**\n")
		for i, v := range s.TopViewed {
			fmt.Fprintf(&sb, "%d. %s（%s 回）\n", i+1, view.VideoLink(v, view.InMessage), view.Number(v.Views))
		}
	}

	sb.WriteString("\n**月別の追加数**\n```\n")
	max := 0
	for _, m := range s.Monthly {
		if m.Count > max {
			max = m.Count
		}
	}
	for _, m := range s.Monthly {
		bar := 0
		if max > 0 {
			bar = (m.Count*20 + max - 1) / max
		}
		fmt.Fprintf(&sb, "%s %-20s %d\n", m.Month.Format("2006/01"), strings.Repeat("█", bar), m.Count)
	}
	sb.WriteString("```\n")

	return sb.String()
}
