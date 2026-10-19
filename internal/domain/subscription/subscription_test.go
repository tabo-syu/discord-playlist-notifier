package subscription_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestIsNew(t *testing.T) {
	t.Parallel()

	notifiedAt := time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		addedAt time.Time
		want    bool
	}{
		{
			name:    "前回の通知より前に追加された",
			addedAt: notifiedAt.Add(-time.Nanosecond),
			want:    false,
		},
		{
			name:    "前回の通知とちょうど同時刻に追加された",
			addedAt: notifiedAt,
			want:    true,
		},
		{
			name:    "前回の通知より後に追加された",
			addedAt: notifiedAt.Add(time.Nanosecond),
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &subscription.Subscription{}
			s.MarkNotified(notifiedAt)

			got := s.IsNew(&library.PlaylistItem{AddedAt: tt.addedAt})

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubscriptionTakeNewVideos(t *testing.T) {
	t.Parallel()

	// 基準時刻を進める時刻
	at := hoursLater(10)

	tests := []struct {
		name      string
		sub       *subscription.Subscription
		latest    *library.Snapshot
		want      *subscription.Subscription
		wantAdded *subscription.VideosAdded
	}{
		{
			name:   "前回の通知の後に追加された動画を返し、基準時刻とタイトルを進める",
			sub:    sub("古いタイトル", hoursLater(0)),
			latest: snapshot(added(video("a", library.PrivacyPublic), 1)),
			want:   sub("新しいタイトル", at),
			wantAdded: &subscription.VideosAdded{
				ChannelID:     "ch",
				PlaylistID:    "PL",
				PlaylistTitle: "新しいタイトル",
				Videos: []*library.PlaylistVideo{
					added(video("a", library.PrivacyPublic), 1),
				},
			},
		},
		{
			name:   "前回の通知とちょうど同時刻に追加された動画を返す",
			sub:    sub("PL", hoursLater(0)),
			latest: snapshot(added(video("a", library.PrivacyPublic), 0)),
			want:   sub("新しいタイトル", at),
			wantAdded: &subscription.VideosAdded{
				ChannelID:     "ch",
				PlaylistID:    "PL",
				PlaylistTitle: "新しいタイトル",
				Videos: []*library.PlaylistVideo{
					added(video("a", library.PrivacyPublic), 0),
				},
			},
		},
		{
			name: "追加された順に並べる",
			sub:  sub("PL", hoursLater(0)),
			latest: snapshot(
				added(video("b", library.PrivacyPublic), 2),
				added(video("a", library.PrivacyPublic), 1),
			),
			want: sub("新しいタイトル", at),
			wantAdded: &subscription.VideosAdded{
				ChannelID:     "ch",
				PlaylistID:    "PL",
				PlaylistTitle: "新しいタイトル",
				Videos: []*library.PlaylistVideo{
					added(video("a", library.PrivacyPublic), 1),
					added(video("b", library.PrivacyPublic), 2),
				},
			},
		},
		{
			name:      "前回の通知の前に追加された動画は返さず、何も変えない",
			sub:       sub("PL", hoursLater(0)),
			latest:    snapshot(added(video("a", library.PrivacyPublic), -1)),
			want:      sub("PL", hoursLater(0)),
			wantAdded: nil,
		},
		{
			name:      "見られない動画は返さない",
			sub:       sub("PL", hoursLater(0)),
			latest:    snapshot(added(video("a", library.PrivacyPrivate), 1)),
			want:      sub("PL", hoursLater(0)),
			wantAdded: nil,
		},
		{
			name:      "詳細を取得していない動画は返さない",
			sub:       sub("PL", hoursLater(0)),
			latest:    snapshot(added(&library.Video{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}, 1)),
			want:      sub("PL", hoursLater(0)),
			wantAdded: nil,
		},
		{
			name:      "削除されたプレイリストの動画は返さない",
			sub:       sub("PL", hoursLater(0)),
			latest:    deleted(snapshot(added(video("a", library.PrivacyPublic), 1))),
			want:      sub("PL", hoursLater(0)),
			wantAdded: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.sub.TakeNewVideos(tt.latest, at)

			assert.Equal(t, tt.want, tt.sub)
			assert.Equal(t, tt.wantAdded, got)
		})
	}
}

func TestPickIntervalDueOn(t *testing.T) {
	t.Parallel()

	sunday := time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)
	monday := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	firstDay := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		interval subscription.PickInterval
		date     time.Time
		want     bool
	}{
		{
			name:     "毎日なら投稿する",
			interval: subscription.PickDaily,
			date:     sunday,
			want:     true,
		},
		{
			name:     "毎週なら月曜日に投稿する",
			interval: subscription.PickWeekly,
			date:     monday,
			want:     true,
		},
		{
			name:     "毎週なら月曜日以外は投稿しない",
			interval: subscription.PickWeekly,
			date:     sunday,
			want:     false,
		},
		{
			name:     "毎月なら 1 日に投稿する",
			interval: subscription.PickMonthly,
			date:     firstDay,
			want:     true,
		},
		{
			name:     "毎月なら 1 日以外は投稿しない",
			interval: subscription.PickMonthly,
			date:     monday,
			want:     false,
		},
		{
			name:     "設定がなければ投稿しない",
			interval: subscription.PickNone,
			date:     firstDay,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.interval.DueOn(tt.date)

			assert.Equal(t, tt.want, got)
		})
	}
}
