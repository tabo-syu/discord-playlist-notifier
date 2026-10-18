package subscription_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestIsNew(t *testing.T) {
	notifiedAt := time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)
	sub := &subscription.Subscription{}
	sub.MarkNotified(notifiedAt)

	tests := map[string]struct {
		addedAt time.Time
		want    bool
	}{
		"基準時刻より前":      {notifiedAt.Add(-time.Nanosecond), false},
		"基準時刻ちょうど":     {notifiedAt, true},
		"基準時刻より後":      {notifiedAt.Add(time.Nanosecond), true},
		"別のタイムゾーンで同時刻": {notifiedAt.In(time.FixedZone("JST", 9*60*60)), true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := sub.IsNew(&library.PlaylistItem{AddedAt: tt.addedAt}); got != tt.want {
				t.Errorf("IsNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPickIntervalDueOn(t *testing.T) {
	sunday := time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)
	monday := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	firstSunday := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		interval subscription.PickInterval
		on       time.Time
		want     bool
	}{
		"毎日":          {subscription.PickDaily, sunday, true},
		"毎週は月曜日に投稿する": {subscription.PickWeekly, monday, true},
		"毎週は月曜日以外に投稿しない":  {subscription.PickWeekly, sunday, false},
		"毎月は 1 日に投稿する":    {subscription.PickMonthly, firstSunday, true},
		"毎月は 1 日以外に投稿しない": {subscription.PickMonthly, monday, false},
		"設定なし": {subscription.PickNone, firstSunday, false},
		"不明な値": {subscription.PickInterval("hourly"), firstSunday, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.interval.DueOn(tt.on); got != tt.want {
				t.Errorf("%q.DueOn(%v) = %v, want %v", tt.interval, tt.on, got, tt.want)
			}
		})
	}
}
