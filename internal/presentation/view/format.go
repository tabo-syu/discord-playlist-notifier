// Package view はデータを Discord に表示する日本語の文言に整形する。
// スラッシュコマンドと定期投稿で共有する。
package view

import (
	"fmt"
	"strings"
)

// Discord はこれより長いメッセージを拒否する
const MAX_MESSAGE_LENGTH = 2000

// Number は n を 3 桁区切りで整形する。例: 1,234,567
func Number[T ~int | ~uint64](n T) string {
	s := fmt.Sprintf("%d", n)

	var sb strings.Builder

	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			sb.WriteRune(',')
		}

		sb.WriteRune(r)
	}

	return sb.String()
}

// Truncate は s を最大 limit 文字（バイトではない）に切り詰め、コードブロックは閉じたままにする。
func Truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	cut := string(runes[:limit-10])
	if strings.Count(cut, "```")%2 == 1 {
		cut += "\n```"
	}

	return cut + "\n…"
}

// JapaneseCount はきりのいい数を日本語の単位で整形する。例: 1000万、1億
func JapaneseCount[T ~int | ~uint64](n T) string {
	switch {
	case n >= 100_000_000 && n%100_000_000 == 0:
		return fmt.Sprintf("%d億", n/100_000_000)
	case n >= 10_000 && n%10_000 == 0:
		return fmt.Sprintf("%d万", n/10_000)
	default:
		return Number(n)
	}
}
