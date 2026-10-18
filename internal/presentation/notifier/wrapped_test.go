package notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

// まとめの本文の整形は view のゴールデンテストで確かめるので、ここでは曲のないまとめで
// Embed の形だけを確かめる。
func TestPostWrapped(t *testing.T) {
	tests := []struct {
		name    string
		wrapped *library.Wrapped
		setup   func(discord *mocks.MockSender)
	}{
		{
			name: "まとめを Embed の説明に入れて投稿する",
			wrapped: &library.Wrapped{
				Year: 2026,
			},
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:       notifier.PURPLE,
					Title:       "🎁 お気に入り の 2026 年まとめ",
					URL:         "https://www.youtube.com/playlist?list=PL",
					Description: "2026 年に追加された曲はありません。\n",
				}).Return(&discordgo.Message{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discord := mocks.NewMockSender(t)
			tt.setup(discord)
			n := notifier.New(discord, jst)

			err := n.PostWrapped(sub("お気に入り", ""), tt.wrapped)

			assert.NoError(t, err)
		})
	}
}
