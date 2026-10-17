package view

import (
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Target は文言を表示する場所で、リンクの書き方が変わる。
type Target int

const (
	// 通常のメッセージ。URL を <> で囲まない限り Discord はリンクごとに
	// プレビューを付けるので、リンクはそのように書く。
	InMessage Target = iota
	// Embed の説明。Discord はプレビューを付けない。
	InEmbed
)

// タイトル中の角括弧は Markdown のリンクを壊すので、
// 全角のものに置き換える。
var linkTitleReplacer = strings.NewReplacer("[", "［", "]", "］")

// VideoLink は動画のタイトルを、動画への Markdown のリンクとして整形する。
func VideoLink(v *library.Video, target Target) string {
	url := "https://www.youtube.com/watch?v=" + string(v.YoutubeID)
	if target == InMessage {
		url = "<" + url + ">"
	}

	return "[" + linkTitleReplacer.Replace(v.Title) + "](" + url + ")"
}
