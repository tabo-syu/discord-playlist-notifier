package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Trending は再生数が伸びている曲の一覧を整形する（見出しを除く）。
func Trending(t *library.Trending, loc *time.Location) string {
	if !t.To.After(t.From) {
		return "まだ記録が足りません。再生数は 6 時間ごとに記録しています。\n"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s〜%s の伸び\n", t.From.In(loc).Format("01/02"), t.To.In(loc).Format("01/02"))
	if len(t.Videos) == 0 {
		sb.WriteString("この期間に再生数が伸びた曲はありません。\n")
		return sb.String()
	}
	for i, v := range t.Videos {
		total := v.Snapshots[len(v.Snapshots)-1].Views
		fmt.Fprintf(&sb, "#%d %s（+%s 回 ・ 計 %s 回）\n", i+1, VideoLink(v.Video, InMessage), Number(v.Increase), Number(total))
	}

	return sb.String()
}
