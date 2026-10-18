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

	// 通知の順はプレイリストのマップの順で変わるので、順序は比べない
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		snapshots     map[library.PlaylistID]*library.Snapshot
		want          []*notification.HiddenVideo
	}{
		{
			name:          "見られなくなった動画を通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": withHidden(snapshot("PL1", added("PL1", a, 0)), a),
			},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "同じ動画を含む他のプレイリストにも通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": withHidden(snapshot("PL1", added("PL1", a, 0)), a),
				"PL2": snapshot("PL2", added("PL2", a, 0)),
			},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
				{ChannelID: "ch2", PlaylistID: "PL2", PlaylistTitle: "PL2", Video: a},
			},
		},
		{
			name:          "同じチャンネルには 1 回だけ通知する",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch1")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": withHidden(snapshot("PL1", added("PL1", a, 0)), a),
				"PL2": snapshot("PL2", added("PL2", a, 0)),
			},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "動画を含まないプレイリストには通知しない",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": withHidden(snapshot("PL1", added("PL1", a, 0)), a),
				"PL2": snapshot("PL2"),
			},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "同期に失敗したプレイリストには通知しない",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1"), sub("PL2", "ch2")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": withHidden(snapshot("PL1", added("PL1", a, 0)), a),
			},
			want: []*notification.HiddenVideo{
				{ChannelID: "ch1", PlaylistID: "PL1", PlaylistTitle: "PL1", Video: a},
			},
		},
		{
			name:          "見られなくなった動画がなければ通知しない",
			subscriptions: []*subscription.Subscription{sub("PL1", "ch1")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL1": snapshot("PL1", added("PL1", video("b", library.PrivacyPublic), 0)),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := notification.FindHiddenVideos(tt.subscriptions, tt.snapshots)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
