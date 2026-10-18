package notification_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindMilestones(t *testing.T) {
	a := &library.Video{YoutubeID: "a", Views: 1_234_567}

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		videos        []*library.Video
		containing    map[library.VideoID][]library.PlaylistID
		want          []*notification.Milestone
	}{
		{
			name:          "動画を含むプレイリストのチャンネルに、達した節目を通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			videos:        []*library.Video{a},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1", "PL2"}},
			want: []*notification.Milestone{
				{ChannelID: "ch1", PlaylistID: "PL1", Video: a, Milestone: 1_000_000},
				{ChannelID: "ch2", PlaylistID: "PL2", Video: a, Milestone: 1_000_000},
			},
		},
		{
			name:          "同じチャンネルには 1 回だけ通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch1")},
			videos:        []*library.Video{a},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1", "PL2"}},
			want: []*notification.Milestone{
				{ChannelID: "ch1", PlaylistID: "PL1", Video: a, Milestone: 1_000_000},
			},
		},
		{
			name:          "動画を含まないプレイリストには通知しない",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			videos:        []*library.Video{a},
			containing:    map[library.VideoID][]library.PlaylistID{"a": {"PL1"}},
			want: []*notification.Milestone{
				{ChannelID: "ch1", PlaylistID: "PL1", Video: a, Milestone: 1_000_000},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := notification.FindMilestones(tt.subscriptions, tt.videos, tt.containing)

			assert.Equal(t, tt.want, got)
		})
	}
}
