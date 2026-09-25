// Package view formats data into the Japanese text shown on Discord. It is
// shared by the slash commands and the scheduled posts.
package view

import (
	"fmt"
	"strings"
)

// Discord rejects messages longer than this
const MAX_MESSAGE_LENGTH = 2000

// Number formats n with thousands separators, e.g. 1,234,567.
func Number[T int | uint64](n T) string {
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

// Truncate cuts s to at most max characters (not bytes), keeping code blocks closed.
func Truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	cut := string(runes[:max-10])
	if strings.Count(cut, "```")%2 == 1 {
		cut += "\n```"
	}

	return cut + "\n…"
}
