package subscription_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestIsNew(t *testing.T) {
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
			s := &subscription.Subscription{}
			s.MarkNotified(notifiedAt)

			got := s.IsNew(&library.PlaylistItem{AddedAt: tt.addedAt})

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPickIntervalDueOn(t *testing.T) {
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
			got := tt.interval.DueOn(tt.date)

			assert.Equal(t, tt.want, got)
		})
	}
}
