package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeTrending(t *testing.T) {
	tests := []struct {
		name     string
		videos   map[library.VideoID]*library.Video
		earliest map[library.VideoID]*library.VideoSnapshot
		latest   map[library.VideoID]*library.VideoSnapshot
		n        int
		want     *library.Trending
	}{
		{
			name: "記録がない",
			n:    3,
			want: &library.Trending{},
		},
		{
			name: "伸びた再生数の多い順に並べる",
			videos: map[library.VideoID]*library.Video{
				"a": video("a", library.PrivacyPublic, 0),
				"b": video("b", library.PrivacyPublic, 0),
			},
			earliest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(0, 100),
				"b": recorded(0, 100),
			},
			latest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(6, 200),
				"b": recorded(6, 400),
			},
			n: 3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
				Videos: []*library.TrendingVideo{
					{Video: video("b", library.PrivacyPublic, 0), Increase: 300},
					{Video: video("a", library.PrivacyPublic, 0), Increase: 100},
				},
			},
		},
		{
			name: "伸びが同じなら ID の順に並べる",
			videos: map[library.VideoID]*library.Video{
				"a": video("a", library.PrivacyPublic, 0),
				"b": video("b", library.PrivacyPublic, 0),
			},
			earliest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(0, 0),
				"b": recorded(0, 0),
			},
			latest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(6, 100),
				"b": recorded(6, 100),
			},
			n: 3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
				Videos: []*library.TrendingVideo{
					{Video: video("a", library.PrivacyPublic, 0), Increase: 100},
					{Video: video("b", library.PrivacyPublic, 0), Increase: 100},
				},
			},
		},
		{
			name: "上位 n 件に絞る",
			videos: map[library.VideoID]*library.Video{
				"a": video("a", library.PrivacyPublic, 0),
				"b": video("b", library.PrivacyPublic, 0),
			},
			earliest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(0, 0),
				"b": recorded(0, 0),
			},
			latest: map[library.VideoID]*library.VideoSnapshot{
				"a": recorded(6, 100),
				"b": recorded(6, 200),
			},
			n: 1,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
				Videos: []*library.TrendingVideo{
					{Video: video("b", library.PrivacyPublic, 0), Increase: 200},
				},
			},
		},
		{
			name:     "見られない動画は除くが、期間には含める",
			videos:   map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPrivate, 0)},
			earliest: map[library.VideoID]*library.VideoSnapshot{"a": recorded(0, 0)},
			latest:   map[library.VideoID]*library.VideoSnapshot{"a": recorded(6, 100)},
			n:        3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
			},
		},
		{
			name:     "プレイリストにない動画は除くが、期間には含める",
			videos:   map[library.VideoID]*library.Video{},
			earliest: map[library.VideoID]*library.VideoSnapshot{"a": recorded(0, 0)},
			latest:   map[library.VideoID]*library.VideoSnapshot{"a": recorded(6, 100)},
			n:        3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
			},
		},
		{
			name:     "1 回しか記録されていない動画は除く",
			videos:   map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
			earliest: map[library.VideoID]*library.VideoSnapshot{"a": recorded(0, 100)},
			latest:   map[library.VideoID]*library.VideoSnapshot{"a": recorded(0, 100)},
			n:        3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(0),
			},
		},
		{
			name:     "再生数が減った動画は除く",
			videos:   map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
			earliest: map[library.VideoID]*library.VideoSnapshot{"a": recorded(0, 100)},
			latest:   map[library.VideoID]*library.VideoSnapshot{"a": recorded(6, 90)},
			n:        3,
			want: &library.Trending{
				From: hoursLater(0),
				To:   hoursLater(6),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := library.ComputeTrending(tt.videos, tt.earliest, tt.latest, tt.n)

			assert.Equal(t, tt.want, got)
		})
	}
}
