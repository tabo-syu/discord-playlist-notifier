package notification_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindNewVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		snapshots     map[library.PlaylistID]*library.Snapshot
		want          []*notification.NewVideos
	}{
		{
			name:          "前回の通知の後に追加された動画を通知する",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL", added("PL", video("a", library.PrivacyPublic), 1)),
			},
			want: []*notification.NewVideos{{
				Subscription:  sub("PL", "ch"),
				PlaylistTitle: "PL",
				Videos:        []*library.PlaylistVideo{added("PL", video("a", library.PrivacyPublic), 1)},
			}},
		},
		{
			name:          "前回の通知とちょうど同時刻に追加された動画を通知する",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL", added("PL", video("a", library.PrivacyPublic), 0)),
			},
			want: []*notification.NewVideos{{
				Subscription:  sub("PL", "ch"),
				PlaylistTitle: "PL",
				Videos:        []*library.PlaylistVideo{added("PL", video("a", library.PrivacyPublic), 0)},
			}},
		},
		{
			name:          "追加された順に並べる",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL",
					added("PL", video("b", library.PrivacyPublic), 2),
					added("PL", video("a", library.PrivacyPublic), 1),
				),
			},
			want: []*notification.NewVideos{{
				Subscription:  sub("PL", "ch"),
				PlaylistTitle: "PL",
				Videos: []*library.PlaylistVideo{
					added("PL", video("a", library.PrivacyPublic), 1),
					added("PL", video("b", library.PrivacyPublic), 2),
				},
			}},
		},
		{
			name:          "前回の通知の前に追加された動画は通知しない",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL", added("PL", video("a", library.PrivacyPublic), -1)),
			},
			want: nil,
		},
		{
			name:          "見られない動画は通知しない",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL", added("PL", video("a", library.PrivacyPrivate), 1)),
			},
			want: nil,
		},
		{
			name:          "詳細を取得していない動画は通知しない",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": snapshot("PL", added("PL", &library.Video{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}, 1)),
			},
			want: nil,
		},
		{
			name:          "削除されたプレイリストは通知しない",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots: map[library.PlaylistID]*library.Snapshot{
				"PL": deleted(snapshot("PL", added("PL", video("a", library.PrivacyPublic), 1))),
			},
			want: nil,
		},
		{
			name:          "同期に失敗したプレイリストは通知しない",
			subscriptions: []*subscription.Subscription{sub("PL", "ch")},
			snapshots:     map[library.PlaylistID]*library.Snapshot{},
			want:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := notification.FindNewVideos(tt.subscriptions, tt.snapshots)

			assert.Equal(t, tt.want, got)
		})
	}
}
