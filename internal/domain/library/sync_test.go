package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestReconcile(t *testing.T) {
	tests := []struct {
		name            string
		stored          []*library.PlaylistItem
		fetched         []*library.FetchedItem
		videos          map[library.VideoID]*library.Video
		wantAdded       []*library.PlaylistItem
		wantRemoved     []*library.PlaylistItem
		wantHidden      []*library.Video
		wantNeedDetails []library.VideoID
		wantChanged     []*library.Video
	}{
		{
			name:            "新しいアイテムを追加し、動画の詳細を取る",
			fetched:         []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			videos:          map[library.VideoID]*library.Video{},
			wantAdded:       []*library.PlaylistItem{stored("1", "a")},
			wantNeedDetails: []library.VideoID{"a"},
			wantChanged:     []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}},
		},
		{
			name:        "消えたアイテムを除く",
			stored:      []*library.PlaylistItem{stored("1", "a")},
			videos:      map[library.VideoID]*library.Video{},
			wantRemoved: []*library.PlaylistItem{stored("1", "a")},
		},
		{
			name:        "見られなくなった動画を報告する",
			stored:      []*library.PlaylistItem{stored("1", "a")},
			fetched:     []*library.FetchedItem{fetched("1", "a", library.PrivacyPrivate)},
			videos:      map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
			wantHidden:  []*library.Video{video("a", library.PrivacyPrivate, 0)},
			wantChanged: []*library.Video{video("a", library.PrivacyPrivate, 0)},
		},
		{
			name:        "初めて見る非公開の動画は報告せず、詳細も取らない",
			fetched:     []*library.FetchedItem{fetched("1", "a", library.PrivacyPrivate)},
			videos:      map[library.VideoID]*library.Video{},
			wantAdded:   []*library.PlaylistItem{stored("1", "a")},
			wantChanged: []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPrivate}},
		},
		{
			name:            "同じ動画が 2 回追加されても、詳細は 1 回だけ取る",
			fetched:         []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic), fetched("2", "a", library.PrivacyPublic)},
			videos:          map[library.VideoID]*library.Video{},
			wantAdded:       []*library.PlaylistItem{stored("1", "a"), stored("2", "a")},
			wantNeedDetails: []library.VideoID{"a"},
			wantChanged:     []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}},
		},
		{
			name:    "変化がなければ何もしない",
			stored:  []*library.PlaylistItem{stored("1", "a")},
			fetched: []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			videos:  map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := library.Reconcile("PL", tt.stored, tt.fetched, tt.videos)

			assert.Equal(t, tt.wantAdded, got.Added)
			assert.Equal(t, tt.wantRemoved, got.Removed)
			assert.Equal(t, tt.wantHidden, got.Hidden)
			assert.Equal(t, tt.wantNeedDetails, got.NeedDetails)
			assert.ElementsMatch(t, tt.wantChanged, got.ChangedVideos())
		})
	}
}

func TestSyncApplyDetails(t *testing.T) {
	tests := []struct {
		name        string
		fetched     []*library.FetchedItem
		details     []*library.Video
		want        map[library.VideoID]*library.Video
		wantChanged []*library.Video
	}{
		{
			name:    "詳細をコピーし、今の再生数の節目を記録する",
			fetched: []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			details: []*library.Video{{YoutubeID: "a", Title: "曲", Views: 250_000}},
			want: map[library.VideoID]*library.Video{
				"a": {YoutubeID: "a", Title: "曲", Views: 250_000, PrivacyStatus: library.PrivacyPublic, ViewMilestone: viewCount(100_000)},
			},
			wantChanged: []*library.Video{
				{YoutubeID: "a", Title: "曲", Views: 250_000, PrivacyStatus: library.PrivacyPublic, ViewMilestone: viewCount(100_000)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := library.Reconcile("PL", nil, tt.fetched, map[library.VideoID]*library.Video{})

			got.ApplyDetails(tt.details)

			assert.Equal(t, tt.want, got.Videos)
			assert.ElementsMatch(t, tt.wantChanged, got.ChangedVideos())
		})
	}
}
