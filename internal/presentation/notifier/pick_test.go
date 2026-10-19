package notifier_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

func TestPostPick(t *testing.T) {
	t.Parallel()

	// UTC では 2026/03/31、JST では 2026/04/01
	picked := added(video("a", 100), time.Date(2026, time.March, 31, 15, 0, 0, 0, time.UTC))

	tests := []struct {
		name  string
		sub   *subscription.Subscription
		setup func(discord *mocks.MockSender)
	}{
		{
			name: "毎日なら今日の一本として投稿する",
			sub:  sub("お気に入り", subscription.PickDaily),
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:  notifier.BLUE,
					Author: &discordgo.MessageEmbedAuthor{Name: "🎬 今日の一本 from お気に入り"},
					Title:  "a",
					URL:    "https://www.youtube.com/watch?v=a&list=PL",
					Fields: []*discordgo.MessageEmbedField{
						{Name: "追加日時", Value: "2026/04/01", Inline: true},
						{Name: "再生回数", Value: "100", Inline: true},
					},
					Image: &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
				}).Return(&discordgo.Message{}, nil)
			},
		},
		{
			name: "毎週なら今週の一本として投稿する",
			sub:  sub("お気に入り", subscription.PickWeekly),
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:  notifier.BLUE,
					Author: &discordgo.MessageEmbedAuthor{Name: "🎬 今週の一本 from お気に入り"},
					Title:  "a",
					URL:    "https://www.youtube.com/watch?v=a&list=PL",
					Fields: []*discordgo.MessageEmbedField{
						{Name: "追加日時", Value: "2026/04/01", Inline: true},
						{Name: "再生回数", Value: "100", Inline: true},
					},
					Image: &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
				}).Return(&discordgo.Message{}, nil)
			},
		},
		{
			name: "毎月なら今月の一本として投稿する",
			sub:  sub("お気に入り", subscription.PickMonthly),
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:  notifier.BLUE,
					Author: &discordgo.MessageEmbedAuthor{Name: "🎬 今月の一本 from お気に入り"},
					Title:  "a",
					URL:    "https://www.youtube.com/watch?v=a&list=PL",
					Fields: []*discordgo.MessageEmbedField{
						{Name: "追加日時", Value: "2026/04/01", Inline: true},
						{Name: "再生回数", Value: "100", Inline: true},
					},
					Image: &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
				}).Return(&discordgo.Message{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			discord := mocks.NewMockSender(t)
			tt.setup(discord)
			n := notifier.New(discord, jst)

			err := n.PostPick(tt.sub, picked, nil)

			assert.NoError(t, err)
		})
	}
}
