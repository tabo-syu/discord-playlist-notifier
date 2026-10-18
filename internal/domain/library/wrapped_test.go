package library_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeWrapped(t *testing.T) {
	jstDate := func(year int, month time.Month, day, hour int) time.Time {
		return time.Date(year, month, day, hour, 0, 0, 0, jst)
	}
	first := available("first", 10)
	big := available("big", 1_000)
	last := available("last", 100)
	// Contents が返すとおり、追加された順に並べる
	contents := []*library.PlaylistVideo{
		added(available("2024", 9_999), jstDate(2024, time.June, 1, 0)),
		// UTC では 2026 年だが、JST では 2025 年の最後
		added(available("previous", 9_999), time.Date(2025, time.December, 31, 14, 0, 0, 0, time.UTC)),
		// UTC では 2025 年だが、JST では 2026 年の最初
		added(first, time.Date(2025, time.December, 31, 15, 30, 0, 0, time.UTC)),
		added(hidden("private", library.PrivacyPrivate), jstDate(2026, time.March, 1, 0)),
		added(big, jstDate(2026, time.March, 10, 0)),
		added(hidden("deleted", library.PrivacyDeleted), jstDate(2026, time.March, 20, 0)),
		added(last, jstDate(2026, time.December, 31, 23)),
		// 見られなくなった動画は最後の 1 曲にしない
		added(hidden("hidden-last", library.PrivacyPrivate), jstDate(2026, time.December, 31, 23)),
		added(available("next", 9_999), jstDate(2027, time.January, 1, 0)),
	}

	got := library.ComputeWrapped(contents, 2026, jst)

	if got.Year != 2026 {
		t.Errorf("Year = %d, want 2026", got.Year)
	}
	// 見られなくなった動画も、追加した数には含める
	if got.Added != 6 {
		t.Errorf("Added = %d, want 6", got.Added)
	}
	if got.PreviousYearAdded != 1 {
		t.Errorf("PreviousYearAdded = %d, want 1", got.PreviousYearAdded)
	}
	if want := jstDate(2026, time.March, 1, 0); !got.BusiestMonth.Month.Equal(want) || got.BusiestMonth.Count != 3 {
		t.Errorf("BusiestMonth = %v, want {%v 3}", got.BusiestMonth, want)
	}
	if got.First == nil || got.First.Video != first {
		t.Errorf("First = %+v, want %q", got.First, first.YoutubeID)
	}
	if got.Last == nil || got.Last.Video != last {
		t.Errorf("Last = %+v, want %q", got.Last, last.YoutubeID)
	}
	assertIDs(t, "TopViewed", videoIDs(got.TopViewed), []library.VideoID{"big", "last", "first"})
}

func TestComputeWrappedBusiestMonthTie(t *testing.T) {
	contents := []*library.PlaylistVideo{
		added(available("a", 1), time.Date(2026, time.February, 1, 0, 0, 0, 0, jst)),
		added(available("b", 1), time.Date(2026, time.May, 1, 0, 0, 0, 0, jst)),
	}

	got := library.ComputeWrapped(contents, 2026, jst)

	// 同じ数なら早い月
	if got.BusiestMonth.Month.Month() != time.February {
		t.Errorf("BusiestMonth = %v, want February", got.BusiestMonth)
	}
}

func TestComputeWrappedEmptyYear(t *testing.T) {
	contents := []*library.PlaylistVideo{
		added(available("a", 1), time.Date(2025, time.June, 1, 0, 0, 0, 0, jst)),
	}

	got := library.ComputeWrapped(contents, 2026, jst)

	if got.Added != 0 || got.PreviousYearAdded != 1 || got.First != nil || got.Last != nil || got.BusiestMonth.Count != 0 || len(got.TopViewed) != 0 {
		t.Errorf("ComputeWrapped() = %+v, want nothing added in 2026", got)
	}
}
