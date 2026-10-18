package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Summary はプレイリストの曲数、追加ペース、再生数ランキングを整形する。
func Summary(title string, s *library.Stats, loc *time.Location) string {
	if s.Total == 0 {
		return fmt.Sprintf("📊 **%s**\nまだ曲の情報がありません。少し待ってからもう一度試してください。\n", title)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "📊 **%s**\n", title)
	fmt.Fprintf(&sb, "曲数: %s 曲（観られる %s / 非公開 %s / 削除 %s）\n", Number(s.Total), Number(s.Available), Number(s.Private), Number(s.Deleted))
	fmt.Fprintf(&sb, "追加ペース: 今月 %d 曲 / 過去 30 日 %d 曲\n", s.AddedThisMonth, s.AddedLast30Days)
	fmt.Fprintf(&sb, "最初の追加: %s\n", s.FirstAddedAt.In(loc).Format("2006/01/02"))

	if len(s.TopViewed) > 0 {
		sb.WriteString("\n**再生数ランキング**\n")
		for i, v := range s.TopViewed {
			fmt.Fprintf(&sb, "%d. %s（%s 回）\n", i+1, VideoLink(v, InMessage), Number(v.Views))
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
