package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// WrappedTitle is the heading of a yearly summary.
func WrappedTitle(playlistTitle string, year int) string {
	return fmt.Sprintf("🎁 %s の %d 年まとめ", playlistTitle, year)
}

// Wrapped formats the body of a yearly summary (without the heading).
func Wrapped(w *library.Wrapped, loc *time.Location, target Target) string {
	if w.Added == 0 {
		return fmt.Sprintf("%d 年に追加された曲はありません。\n", w.Year)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "追加した曲: %s 曲（前の年 %s 曲）\n", Number(w.Added), Number(w.PreviousYearAdded))
	fmt.Fprintf(&sb, "一番追加した月: %d 月（%s 曲）\n", int(w.BusiestMonth.Month.Month()), Number(w.BusiestMonth.Count))
	if w.First != nil {
		fmt.Fprintf(&sb, "最初の一曲: %s（%s）\n", VideoLink(w.First.Video, target), w.First.Item.AddedAt.In(loc).Format("01/02"))
		fmt.Fprintf(&sb, "最後の一曲: %s（%s）\n", VideoLink(w.Last.Video, target), w.Last.Item.AddedAt.In(loc).Format("01/02"))
	}

	if len(w.TopViewed) > 0 {
		sb.WriteString("\n**この年に追加した曲の再生数ランキング**\n")
		for i, v := range w.TopViewed {
			fmt.Fprintf(&sb, "%d. %s（%s 回）\n", i+1, VideoLink(v, target), Number(v.Views))
		}
	}

	return sb.String()
}
