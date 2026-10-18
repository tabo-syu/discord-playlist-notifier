package view

import (
	"fmt"
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Random はランダムに選んだ曲の一覧を整形する。
func Random(picked []*library.PlaylistVideo) string {
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
