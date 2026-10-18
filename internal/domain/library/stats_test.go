package library_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeStats(t *testing.T) {
	t.Parallel()

	now := date(2026, time.October, 18, 12)

	tests := []struct {
		name     string
		contents []*library.PlaylistVideo
		want     *library.Stats
	}{
		{
			name:     "中身がなくても、直近 6 か月の数は 0 で埋める",
			contents: nil,
			want: &library.Stats{
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 0},
					{Month: date(2026, time.October, 1, 0), Count: 0},
				},
			},
		},
		{
			name: "公開状態ごとに数え、再生数の多い順に 3 件並べる",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 100), date(2026, time.October, 1, 0)),
				added(video("b", library.PrivacyUnlisted, 400), date(2026, time.October, 1, 0)),
				added(video("c", library.PrivacyPublic, 300), date(2026, time.October, 1, 0)),
				added(video("d", library.PrivacyPublic, 200), date(2026, time.October, 1, 0)),
				added(video("e", library.PrivacyPrivate, 900), date(2026, time.October, 1, 0)),
				added(video("f", library.PrivacyDeleted, 900), date(2026, time.October, 1, 0)),
			},
			want: &library.Stats{
				Total:           6,
				Available:       4,
				Private:         1,
				Deleted:         1,
				AddedLast30Days: 6,
				AddedThisMonth:  6,
				FirstAddedAt:    date(2026, time.October, 1, 0),
				LastAddedAt:     date(2026, time.October, 1, 0),
				TopViewed: []*library.Video{
					video("b", library.PrivacyUnlisted, 400),
					video("c", library.PrivacyPublic, 300),
					video("d", library.PrivacyPublic, 200),
				},
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 0},
					{Month: date(2026, time.October, 1, 0), Count: 6},
				},
			},
		},
		{
			name: "2 回追加された動画は 2 回数えるが、ランキングには 1 回だけ入れる",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 100), date(2026, time.October, 1, 0)),
				added(video("a", library.PrivacyPublic, 100), date(2026, time.October, 2, 0)),
			},
			want: &library.Stats{
				Total:           2,
				Available:       2,
				AddedLast30Days: 2,
				AddedThisMonth:  2,
				FirstAddedAt:    date(2026, time.October, 1, 0),
				LastAddedAt:     date(2026, time.October, 2, 0),
				TopViewed:       []*library.Video{video("a", library.PrivacyPublic, 100)},
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 0},
					{Month: date(2026, time.October, 1, 0), Count: 2},
				},
			},
		},
		{
			name: "再生数が同じなら ID の順に並べる",
			contents: []*library.PlaylistVideo{
				added(video("b", library.PrivacyPublic, 100), date(2026, time.October, 1, 0)),
				added(video("a", library.PrivacyPublic, 100), date(2026, time.October, 1, 0)),
			},
			want: &library.Stats{
				Total:           2,
				Available:       2,
				AddedLast30Days: 2,
				AddedThisMonth:  2,
				FirstAddedAt:    date(2026, time.October, 1, 0),
				LastAddedAt:     date(2026, time.October, 1, 0),
				TopViewed: []*library.Video{
					video("a", library.PrivacyPublic, 100),
					video("b", library.PrivacyPublic, 100),
				},
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 0},
					{Month: date(2026, time.October, 1, 0), Count: 2},
				},
			},
		},
		{
			name: "月の区切りは JST で決める",
			contents: []*library.PlaylistVideo{
				// UTC では 10 月、JST では 9 月 30 日 23 時
				added(video("a", library.PrivacyPublic, 0), time.Date(2026, time.September, 30, 14, 0, 0, 0, time.UTC)),
				// UTC では 9 月、JST では 10 月 1 日 0 時
				added(video("b", library.PrivacyPublic, 0), time.Date(2026, time.September, 30, 15, 0, 0, 0, time.UTC)),
			},
			want: &library.Stats{
				Total:           2,
				Available:       2,
				AddedLast30Days: 2,
				AddedThisMonth:  1,
				FirstAddedAt:    date(2026, time.September, 30, 23),
				LastAddedAt:     date(2026, time.October, 1, 0),
				TopViewed: []*library.Video{
					video("a", library.PrivacyPublic, 0),
					video("b", library.PrivacyPublic, 0),
				},
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 1},
					{Month: date(2026, time.October, 1, 0), Count: 1},
				},
			},
		},
		{
			name: "直近 30 日と 6 か月より前のものは期間の数に入れない",
			contents: []*library.PlaylistVideo{
				added(video("a", library.PrivacyPublic, 0), date(2026, time.April, 30, 23)),
				added(video("b", library.PrivacyPublic, 0), date(2026, time.September, 18, 11)),
				added(video("c", library.PrivacyPublic, 0), date(2026, time.September, 18, 12)),
			},
			want: &library.Stats{
				Total:           3,
				Available:       3,
				AddedLast30Days: 1,
				AddedThisMonth:  0,
				FirstAddedAt:    date(2026, time.April, 30, 23),
				LastAddedAt:     date(2026, time.September, 18, 12),
				TopViewed: []*library.Video{
					video("a", library.PrivacyPublic, 0),
					video("b", library.PrivacyPublic, 0),
					video("c", library.PrivacyPublic, 0),
				},
				Monthly: []library.MonthCount{
					{Month: date(2026, time.May, 1, 0), Count: 0},
					{Month: date(2026, time.June, 1, 0), Count: 0},
					{Month: date(2026, time.July, 1, 0), Count: 0},
					{Month: date(2026, time.August, 1, 0), Count: 0},
					{Month: date(2026, time.September, 1, 0), Count: 2},
					{Month: date(2026, time.October, 1, 0), Count: 0},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.ComputeStats(tt.contents, now, jst)

			assert.Equal(t, tt.want, got)
		})
	}
}
