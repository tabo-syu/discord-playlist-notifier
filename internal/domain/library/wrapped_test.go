package library_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeWrapped(t *testing.T) {
	// contents は Contents が返すとおり、追加された順に並べる
	tests := []struct {
		name     string
		contents []*library.PlaylistVideo
		want     *library.Wrapped
	}{
		{
			name:     "その年に何も追加されていない",
			contents: []*library.PlaylistVideo{added(video("a", library.PrivacyPublic, 0), date(2025, time.June, 1, 0))},
			want:     &library.Wrapped{Year: 2026, PreviousYearAdded: 1},
		},
		{
			name: "その年に追加されたものを集計する",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 100), date(2026, time.March, 1, 0)),
				added(video("b", library.PrivacyPublic, 300), date(2026, time.March, 2, 0)),
				added(video("c", library.PrivacyPublic, 200), date(2026, time.July, 1, 0)),
			},
			want: &library.Wrapped{
				Year: 2026, Added: 3,
				BusiestMonth: library.MonthCount{Month: date(2026, time.March, 1, 0), Count: 2},
				TopViewed: []*library.Video{
					video("b", library.PrivacyPublic, 300),
					video("c", library.PrivacyPublic, 200),
					video("a", library.PrivacyPublic, 100),
				},
				First: added(video("a", library.PrivacyPublic, 100), date(2026, time.March, 1, 0)),
				Last:  added(video("c", library.PrivacyPublic, 200), date(2026, time.July, 1, 0)),
			},
		},
		{
			name: "見られない動画は追加した数に入れるが、ランキングと最初・最後には入れない",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPrivate, 0), date(2026, time.March, 1, 0)),
				added(video("b", library.PrivacyPublic, 0), date(2026, time.March, 2, 0)),
				added(video("c", library.PrivacyDeleted, 0), date(2026, time.March, 3, 0)),
			},
			want: &library.Wrapped{
				Year: 2026, Added: 3,
				BusiestMonth: library.MonthCount{Month: date(2026, time.March, 1, 0), Count: 3},
				TopViewed:    []*library.Video{video("b", library.PrivacyPublic, 0)},
				First:        added(video("b", library.PrivacyPublic, 0), date(2026, time.March, 2, 0)),
				Last:         added(video("b", library.PrivacyPublic, 0), date(2026, time.March, 2, 0)),
			},
		},
		{
			name: "追加した数が同じ月なら、早い月を一番多い月にする",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 0), date(2026, time.February, 1, 0)),
				added(video("b", library.PrivacyPublic, 0), date(2026, time.May, 1, 0)),
			},
			want: &library.Wrapped{
				Year: 2026, Added: 2,
				BusiestMonth: library.MonthCount{Month: date(2026, time.February, 1, 0), Count: 1},
				TopViewed: []*library.Video{
					video("a", library.PrivacyPublic, 0),
					video("b", library.PrivacyPublic, 0),
				},
				First: added(video("a", library.PrivacyPublic, 0), date(2026, time.February, 1, 0)),
				Last:  added(video("b", library.PrivacyPublic, 0), date(2026, time.May, 1, 0)),
			},
		},
		{
			name: "年の区切りは JST で決める",
			contents: []*library.PlaylistVideo{
				// UTC では 2025 年、JST では 2026 年 1 月 1 日 0 時
				added(video("a", library.PrivacyPublic, 0), time.Date(2025, time.December, 31, 15, 0, 0, 0, time.UTC)),
				// UTC では 2026 年、JST では 2026 年 12 月 31 日 23 時
				added(video("b", library.PrivacyPublic, 0), time.Date(2026, time.December, 31, 14, 0, 0, 0, time.UTC)),
				// UTC では 2026 年、JST では 2027 年 1 月 1 日 0 時
				added(video("c", library.PrivacyPublic, 0), time.Date(2026, time.December, 31, 15, 0, 0, 0, time.UTC)),
			},
			want: &library.Wrapped{
				Year: 2026, Added: 2,
				BusiestMonth: library.MonthCount{Month: date(2026, time.January, 1, 0), Count: 1},
				TopViewed: []*library.Video{
					video("a", library.PrivacyPublic, 0),
					video("b", library.PrivacyPublic, 0),
				},
				First: added(video("a", library.PrivacyPublic, 0), time.Date(2025, time.December, 31, 15, 0, 0, 0, time.UTC)),
				Last:  added(video("b", library.PrivacyPublic, 0), time.Date(2026, time.December, 31, 14, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "前の年に追加した数を数える",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 0), date(2024, time.June, 1, 0)),
				added(video("b", library.PrivacyPublic, 0), date(2025, time.June, 1, 0)),
				added(video("c", library.PrivacyPublic, 0), date(2025, time.July, 1, 0)),
			},
			want: &library.Wrapped{Year: 2026, PreviousYearAdded: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := library.ComputeWrapped(tt.contents, 2026, jst)

			assert.Equal(t, tt.want, got)
		})
	}
}
