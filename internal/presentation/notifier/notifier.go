// Package notifier は Discord のチャンネルに Embed を投稿することで
// application.Notifier を実装する。
package notifier

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Embed の左端の線の色
const (
	RED    = 0xff0000
	GOLD   = 0xffd700
	BLUE   = 0x1e90ff
	GRAY   = 0x808080
	PURPLE = 0x9b59b6
)

type notifier struct {
	session *discordgo.Session
	// 日付の表示に使うタイムゾーン
	location *time.Location
}

func New(s *discordgo.Session, loc *time.Location) *notifier {
	return &notifier{s, loc}
}
