// Package notifier は Discord のチャンネルに Embed を投稿することで
// application.Notifier を実装する。
package notifier

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"github.com/bwmarrin/discordgo"
)

type notifier struct {
	session *discordgo.Session
	// 日付の表示に使うタイムゾーン
	location *time.Location
}

func New(s *discordgo.Session, loc *time.Location) *notifier {
	return &notifier{s, loc}
}

func (n *notifier) send(channelID subscription.ChannelID, embed *discordgo.MessageEmbed) error {
	_, err := n.session.ChannelMessageSendEmbed(string(channelID), embed)

	return err
}

func color(hex string) int {
	color, err := strconv.ParseInt(hex, 16, 0)
	if err != nil {
		// エラーのときは赤にする
		return 0xff0000
	}

	return int(color)
}

func separator(integer library.ViewCount) string {
	// strings.Builder を使って効率よく組み立てる
	var sb strings.Builder
	str := fmt.Sprintf("%d", integer)

	// 必要なカンマの数を計算する
	commas := (len(str) - 1) / 3

	// 再割り当てを避けるためにバッファを事前に確保する
	sb.Grow(len(str) + commas)

	// カンマを入れながら数字を追加する
	for i, char := range str {
		// 先頭を除き、右から 3 桁ごとにカンマを入れる
		if i > 0 && (len(str)-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(char)
	}

	return sb.String()
}
