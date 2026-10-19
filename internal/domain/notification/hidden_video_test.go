package notification_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindHiddenVideos(t *testing.T) {
	t.Parallel()

	a := video("a", library.PrivacyPrivate)
	b := video("b", library.PrivacyDeleted)

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		events        []*library.VideoHidden
		containing    map[library.VideoID][]library.PlaylistID
		want          []*notification.HiddenVideo
	}{
		{
			name:          "動画を含むプレイリストのチャンネルに、通知設定のタイトルで通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			events:        []*library.VideoHidden{{Video: a}},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1", "PL2"}},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
				{ChannelID: "ch2", PlaylistID: "PL2", PlaylistTitle: "PL2", Video: a},
			},
		},
		{
			name:          "同じチャンネルには 1 回だけ通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch1")},
			events:        []*library.VideoHidden{{Video: a}},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1", "PL2"}},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "同じ動画のイベントが 2 回あっても、1 回だけ通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1")},
			events:        []*library.VideoHidden{{Video: a}, {Video: a}},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1"}},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "同じチャンネルでも、別の動画はそれぞれ通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1")},
			events:        []*library.VideoHidden{{Video: a}, {Video: b}},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1"}, "b": {"PL1"}},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: b},
			},
		},
		{
			name:          "動画を含まないプレイリストには通知しない",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			events:        []*library.VideoHidden{{Video: a}},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1"}},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := notification.FindHiddenVideos(tt.subscriptions, tt.events, tt.containing)

			assert.Equal(t, tt.want, got)
		})
	}
}
