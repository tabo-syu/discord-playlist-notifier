package library_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeStats(t *testing.T) {
	jstDate := func(month time.Month, day, hour int) time.Time {
		return time.Date(2026, month, day, hour, 0, 0, 0, jst)
	}
	now := jstDate(time.October, 18, 12)
	first := available("first", 900)
	thisMonth := available("this-month", 500)
	tie := available("tie", 500)
	low := available("low", 50)
	contents := []*library.PlaylistVideo{
		// 集計の期間（直近 6 か月）より前
		added(hidden("deleted", library.PrivacyDeleted), jstDate(time.April, 30, 23)),
		added(first, jstDate(time.May, 1, 0)),
		added(low, jstDate(time.June, 15, 0)),
		// UTC では 10 月だが、JST では 9 月の最後
		added(hidden("private", library.PrivacyPrivate), time.Date(2026, time.September, 30, 14, 0, 0, 0, time.UTC)),
		// UTC では 9 月だが、JST では 10 月の最初
		added(thisMonth, time.Date(2026, time.September, 30, 15, 30, 0, 0, time.UTC)),
		added(tie, jstDate(time.October, 5, 0)),
		// 2 回追加された動画
		added(thisMonth, jstDate(time.October, 10, 0)),
	}

	got := library.ComputeStats(contents, now, jst)

	counts := []struct {
		name      string
		got, want int
	}{
		// 2 回追加された動画も、アイテムの数だけ数える
		{"Total", got.Total, 7},
		{"Available", got.Available, 5},
		{"Private", got.Private, 1},
		{"Deleted", got.Deleted, 1},
		{"AddedLast30Days", got.AddedLast30Days, 4},
		{"AddedThisMonth", got.AddedThisMonth, 3},
	}
	for _, c := range counts {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}

	if want := jstDate(time.April, 30, 23); !got.FirstAddedAt.Equal(want) || got.FirstAddedAt.Location() != jst {
		t.Errorf("FirstAddedAt = %v, want %v", got.FirstAddedAt, want)
	}
	if want := jstDate(time.October, 10, 0); !got.LastAddedAt.Equal(want) {
		t.Errorf("LastAddedAt = %v, want %v", got.LastAddedAt, want)
	}

	// 再生数が同じなら ID の順。2 回追加された動画は 1 回だけ
	assertIDs(t, "TopViewed", videoIDs(got.TopViewed), []library.VideoID{"first", "this-month", "tie"})

	want := []library.MonthCount{
		{Month: jstDate(time.May, 1, 0), Count: 1},
		{Month: jstDate(time.June, 1, 0), Count: 1},
		{Month: jstDate(time.July, 1, 0), Count: 0},
		{Month: jstDate(time.August, 1, 0), Count: 0},
		{Month: jstDate(time.September, 1, 0), Count: 1},
		{Month: jstDate(time.October, 1, 0), Count: 3},
	}
	if len(got.Monthly) != len(want) {
		t.Fatalf("Monthly = %v, want %v", got.Monthly, want)
	}
	for i := range want {
		if !got.Monthly[i].Month.Equal(want[i].Month) || got.Monthly[i].Count != want[i].Count {
			t.Errorf("Monthly[%d] = %v, want %v", i, got.Monthly[i], want[i])
		}
	}
}

func TestComputeStatsEmpty(t *testing.T) {
	got := library.ComputeStats(nil, time.Date(2026, time.October, 18, 0, 0, 0, 0, jst), jst)

	if got.Total != 0 || !got.FirstAddedAt.IsZero() || !got.LastAddedAt.IsZero() || len(got.TopViewed) != 0 {
		t.Errorf("ComputeStats(nil) = %+v, want zero values", got)
	}
	// 中身がなくても、月ごとの数は 0 で埋める
	if len(got.Monthly) != library.STATS_MONTHS {
		t.Errorf("len(Monthly) = %d, want %d", len(got.Monthly), library.STATS_MONTHS)
	}
}
