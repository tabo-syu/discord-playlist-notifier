package notifier_test

import (
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

func TestNotifyMilestone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		notice *notification.Milestone
		live   map[library.VideoID]*application.LiveVideo
		setup  func(discord *mocks.MockSender)
	}{
		{
			name: "達した節目と、送る直前の再生数を出す",
			notice: &notification.Milestone{
				ChannelID:  "ch",
				PlaylistID: "PL",
				Video:      video("a", 1_000_500),
				Milestone:  1_000_000,
			},
			live: map[library.VideoID]*application.LiveVideo{
				"a": {
					ChannelName: "YOASOBI",
					ChannelIcon: "https://example.com/icon.png",
					Views:       1_002_000,
				},
			},
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:  notifier.GOLD,
					Author: &discordgo.MessageEmbedAuthor{Name: "🎉 100万回再生を突破しました！"},
					Title:  "a",
					URL:    "https://www.youtube.com/watch?v=a&list=PL",
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "YOASOBI",
						IconURL: "https://example.com/icon.png",
					},
					Fields: []*discordgo.MessageEmbedField{
						{Name: "再生回数", Value: "1,002,000", Inline: true},
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
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

			err := n.NotifyMilestone(tt.notice, tt.live)

			assert.NoError(t, err)
		})
	}
}
