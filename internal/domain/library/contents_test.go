package library_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestSnapshotContains(t *testing.T) {
	t.Parallel()

	snapshot := &library.Snapshot{Items: []*library.PlaylistItem{{VideoYoutubeID: "a"}}}

	tests := []struct {
		name    string
		videoID library.VideoID
		want    bool
	}{
		{
			name:    "含む",
			videoID: "a",
			want:    true,
		},
		{
			name:    "含まない",
			videoID: "b",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := snapshot.Contains(tt.videoID)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPickRandom(t *testing.T) {
	t.Parallel()

	a := added(video("a", library.PrivacyPublic, 0), date(2026, time.October, 1, 0))
	b := added(video("b", library.PrivacyPublic, 0), date(2026, time.October, 2, 0))
	aAgain := added(video("a", library.PrivacyPublic, 0), date(2026, time.October, 3, 0))
	private := added(video("private", library.PrivacyPrivate, 0), date(2026, time.October, 4, 0))

	// 選ぶ順はランダムなので、全て選ぶ数を指定して中身だけを比べる
	tests := []struct {
		name     string
		contents []*library.PlaylistVideo
		n        int
		want     []*library.PlaylistVideo
	}{
		{
			name:     "視聴できる動画を全て選ぶ",
			contents: []*library.PlaylistVideo{a, b},
			n:        10,
			want:     []*library.PlaylistVideo{a, b},
		},
		{
			name:     "視聴できない動画は選ばない",
			contents: []*library.PlaylistVideo{a, private},
			n:        10,
			want:     []*library.PlaylistVideo{a},
		},
		{
			name:     "2 回追加された動画は 1 回だけ選ぶ",
			contents: []*library.PlaylistVideo{a, aAgain},
			n:        10,
			want:     []*library.PlaylistVideo{a},
		},
		{
			name:     "何も選ばない",
			contents: []*library.PlaylistVideo{a, b},
			n:        0,
			want:     []*library.PlaylistVideo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.PickRandom(tt.contents, tt.n)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestPickRandomLimit(t *testing.T) {
	t.Parallel()

	a := added(video("a", library.PrivacyPublic, 0), date(2026, time.October, 1, 0))
	b := added(video("b", library.PrivacyPublic, 0), date(2026, time.October, 2, 0))
	c := added(video("c", library.PrivacyPublic, 0), date(2026, time.October, 3, 0))

	tests := []struct {
		name     string
		contents []*library.PlaylistVideo
		n        int
		wantLen  int
	}{
		{
			name:     "n 件に絞る",
			contents: []*library.PlaylistVideo{a, b, c},
			n:        2,
			wantLen:  2,
		},
		{
			name:     "候補が n 件より少なければ全て選ぶ",
			contents: []*library.PlaylistVideo{a, b, c},
			n:        5,
			wantLen:  3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.PickRandom(tt.contents, tt.n)

			assert.Len(t, got, tt.wantLen)
			assert.Subset(t, tt.contents, got)
		})
	}
}
