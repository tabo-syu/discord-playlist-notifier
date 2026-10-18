package view

import (
	"fmt"
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// History は動画の再生数、高評価数、コメント数の推移を整形する。
func History(h *library.VideoHistory, days int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "📈 **%s** の推移（直近 %d 日）\n", h.Video.Title, days)

	if len(h.Snapshots) < 2 {
		sb.WriteString("まだ記録が足りません。再生数などは 6 時間ごとに記録しています。\n")
	} else {
		first, last := h.Snapshots[0], h.Snapshots[len(h.Snapshots)-1]
		fmt.Fprintf(&sb, "再生数: %s\n", change(uint64(first.Views), uint64(last.Views)))
		fmt.Fprintf(&sb, "高評価: %s\n", change(first.Likes, last.Likes))
		fmt.Fprintf(&sb, "コメント: %s\n", change(first.Comments, last.Comments))
	}
	fmt.Fprintf(&sb, "https://www.youtube.com/watch?v=%s", h.Video.YoutubeID)

	return sb.String()
}

// change は before から after への値の変化を整形する。例: 1,000 → 1,200（+200）
func change(before, after uint64) string {
	sign := "+"
	diff := after - before
	if after < before {
		sign = "-"
		diff = before - after
	}

	return fmt.Sprintf("%s → %s（%s%s）", Number(before), Number(after), sign, Number(diff))
}
