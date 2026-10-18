package notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
)

func TestNotifyHiddenVideo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		notice *notification.HiddenVideo
		setup  func(discord *mocks.MockSender)
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
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:     notifier.GRAY,
					Author:    &discordgo.MessageEmbedAuthor{Name: "お気に入り の動画が非公開になりました"},
					Title:     "a",
					URL:       "https://www.youtube.com/watch?v=a&list=PL",
					Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://i.ytimg.com/vi/a/hqdefault.jpg"},
				}).Return(&discordgo.Message{}, nil)
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
			setup: func(discord *mocks.MockSender) {
				discord.EXPECT().ChannelMessageSendEmbed("ch", &discordgo.MessageEmbed{
					Color:     notifier.GRAY,
					Author:    &discordgo.MessageEmbedAuthor{Name: "お気に入り の動画が削除されました"},
					Title:     "a",
					URL:       "https://www.youtube.com/watch?v=a&list=PL",
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

			err := n.NotifyHiddenVideo(tt.notice)

			assert.NoError(t, err)
		})
	}
}
