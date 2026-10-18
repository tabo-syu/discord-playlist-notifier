package notifier_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

func TestNotifyNewVideos(t *testing.T) {
	t.Parallel()

	published := &library.Video{YoutubeID: "a", Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 1000, PublishedAt: time.Date(2019, time.December, 15, 0, 0, 0, 0, time.UTC)}

	tests := []struct {
		name       string
		news       *notification.NewVideos
		live       map[library.VideoID]*library.LiveVideo
		failTitles []string
		want       []*fake.SentEmbed
		wantErr    error
	}{
		{
			name: "送る直前に取得した再生数とチャンネルを出し、投稿日時を Discord に渡す",
			news: &notification.NewVideos{
				Subscription: sub("お気に入り", ""),
				Videos: []*library.PlaylistVideo{
					// UTC では 10/18 の 0 時、JST では 9 時
					added(published, time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)),
				},
			},
			live: map[library.VideoID]*library.LiveVideo{
				"a": {
					ChannelName: "YOASOBI",
					ChannelIcon: "https://example.com/icon.png",
					Views:       1234,
				},
			},
			want: []*fake.SentEmbed{
				{
					ChannelID: "ch",
					Embed: &discordgo.MessageEmbed{
						Color:  notifier.RED,
						Author: &discordgo.MessageEmbedAuthor{Name: "お気に入り に追加されました！"},
						Title:  "a",
						URL:    "https://www.youtube.com/watch?v=a&list=PL",
						Footer: &discordgo.MessageEmbedFooter{
							Text:    "YOASOBI",
							IconURL: "https://example.com/icon.png",
						},
						Fields: []*discordgo.MessageEmbedField{
							{Name: "追加日時", Value: "2026/10/18 09:00:00", Inline: true},
							{Name: "再生回数", Value: "1,234", Inline: true},
						},
						Image:     &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
						Timestamp: "2019-12-15T00:00:00Z",
					},
				},
			},
		},
		{
			name: "取得できなかった動画は、保存した再生数でチャンネルなしで送る",
			news: &notification.NewVideos{
				Subscription: sub("お気に入り", ""),
				Videos: []*library.PlaylistVideo{
					added(video("b", 50), time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)),
				},
			},
			live: nil,
			want: []*fake.SentEmbed{
				{
					ChannelID: "ch",
					Embed: &discordgo.MessageEmbed{
						Color:  notifier.RED,
						Author: &discordgo.MessageEmbedAuthor{Name: "お気に入り に追加されました！"},
						Title:  "b",
						URL:    "https://www.youtube.com/watch?v=b&list=PL",
						Fields: []*discordgo.MessageEmbedField{
							{Name: "追加日時", Value: "2026/10/18 09:00:00", Inline: true},
							{Name: "再生回数", Value: "50", Inline: true},
						},
						Image: &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/b/hqdefault.jpg"},
					},
				},
			},
		},
		{
			name: "1 通失敗しても残りを送り、失敗を返す",
			news: &notification.NewVideos{
				Subscription: sub("お気に入り", ""),
				Videos: []*library.PlaylistVideo{
					added(video("b", 50), time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)),
					added(video("c", 60), time.Date(2026, time.October, 18, 1, 0, 0, 0, time.UTC)),
				},
			},
			failTitles: []string{"b"},
			want: []*fake.SentEmbed{
				{
					ChannelID: "ch",
					Embed: &discordgo.MessageEmbed{
						Color:  notifier.RED,
						Author: &discordgo.MessageEmbedAuthor{Name: "お気に入り に追加されました！"},
						Title:  "c",
						URL:    "https://www.youtube.com/watch?v=c&list=PL",
						Fields: []*discordgo.MessageEmbedField{
							{Name: "追加日時", Value: "2026/10/18 10:00:00", Inline: true},
							{Name: "再生回数", Value: "60", Inline: true},
						},
						Image: &discordgo.MessageEmbedImage{URL: "https://i.ytimg.com/vi/c/hqdefault.jpg"},
					},
				},
			},
			wantErr: fake.Err,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			discord := &fake.Discord{FailTitles: tt.failTitles}
			n := notifier.New(discord, jst)

			err := n.NotifyNewVideos(tt.news, tt.live)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, discord.Sent)
		})
	}
}
