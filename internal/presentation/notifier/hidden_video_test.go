package notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

func TestNotifyHiddenVideo(t *testing.T) {
	tests := []struct {
		name   string
		notice *notification.HiddenVideo
		want   []*fake.SentEmbed
	}{
		{
			name: "非公開になった動画を、小さいサムネイルで知らせる",
			notice: &notification.HiddenVideo{
				ChannelID:     "ch",
				PlaylistID:    "PL",
				PlaylistTitle: "お気に入り",
				Video: &library.Video{
					YoutubeID:     "a",
					Title:         "a",
					PrivacyStatus: library.PrivacyPrivate,
				},
			},
			want: []*fake.SentEmbed{
				{
					ChannelID: "ch",
					Embed: &discordgo.MessageEmbed{
						Color:     notifier.GRAY,
						Author:    &discordgo.MessageEmbedAuthor{Name: "お気に入り の動画が非公開になりました"},
						Title:     "a",
						URL:       "https://www.youtube.com/watch?v=a&list=PL",
						Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
					},
				},
			},
		},
		{
			name: "削除された動画を知らせる",
			notice: &notification.HiddenVideo{
				ChannelID:     "ch",
				PlaylistID:    "PL",
				PlaylistTitle: "お気に入り",
				Video: &library.Video{
					YoutubeID:     "a",
					Title:         "a",
					PrivacyStatus: library.PrivacyDeleted,
				},
			},
			want: []*fake.SentEmbed{
				{
					ChannelID: "ch",
					Embed: &discordgo.MessageEmbed{
						Color:     notifier.GRAY,
						Author:    &discordgo.MessageEmbedAuthor{Name: "お気に入り の動画が削除されました"},
						Title:     "a",
						URL:       "https://www.youtube.com/watch?v=a&list=PL",
						Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discord := &fake.Discord{}
			n := notifier.New(discord, jst)

			err := n.NotifyHiddenVideo(tt.notice)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, discord.Sent)
		})
	}
}
