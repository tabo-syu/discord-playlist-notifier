package library_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestSyncVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		fetched         []*library.FetchedItem
		videos          map[library.VideoID]*library.Video
		wantHidden      []*library.VideoHidden
		wantNeedDetails []library.VideoID
		wantChanged     []*library.Video
	}{
		{
			name:            "初めて見る動画を作り、詳細を取る",
			fetched:         []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			videos:          map[library.VideoID]*library.Video{},
			wantNeedDetails: []library.VideoID{"a"},
			wantChanged:     []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}},
		},
		{
			name:    "見られなくなった動画を報告する",
			fetched: []*library.FetchedItem{fetched("1", "a", library.PrivacyPrivate)},
			videos:  map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
			wantHidden: []*library.VideoHidden{
				{Video: video("a", library.PrivacyPrivate, 0)},
			},
			wantChanged: []*library.Video{video("a", library.PrivacyPrivate, 0)},
		},
		{
			name:        "初めて見る非公開の動画は報告せず、詳細も取らない",
			fetched:     []*library.FetchedItem{fetched("1", "a", library.PrivacyPrivate)},
			videos:      map[library.VideoID]*library.Video{},
			wantChanged: []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPrivate}},
		},
		{
			name:            "同じ動画が 2 回追加されても、詳細は 1 回だけ取る",
			fetched:         []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic), fetched("2", "a", library.PrivacyPublic)},
			videos:          map[library.VideoID]*library.Video{},
			wantNeedDetails: []library.VideoID{"a"},
			wantChanged:     []*library.Video{{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic}},
		},
		{
			name:    "変化がなければ何もしない",
			fetched: []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			videos:  map[library.VideoID]*library.Video{"a": video("a", library.PrivacyPublic, 0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.SyncVideos(tt.fetched, tt.videos)

			assert.Equal(t, tt.wantHidden, got.Hidden)
			assert.Equal(t, tt.wantNeedDetails, got.NeedDetails)
			assert.ElementsMatch(t, tt.wantChanged, got.ChangedVideos())
		})
	}
}

func TestVideoSyncApplyDetails(t *testing.T) {
	t.Parallel()

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
				"a": {
					YoutubeID:      "a",
					DetailsFetched: true,
					Title:          "曲",
					Views:          250_000,
					PrivacyStatus:  library.PrivacyPublic,
					ViewMilestone:  new(library.ViewCount(100_000)),
				},
			},
			wantChanged: []*library.Video{
				{YoutubeID: "a", DetailsFetched: true, Title: "曲", Views: 250_000, PrivacyStatus: library.PrivacyPublic, ViewMilestone: new(library.ViewCount(100_000))},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.SyncVideos(tt.fetched, map[library.VideoID]*library.Video{})

			got.ApplyDetails(tt.details)

			assert.Equal(t, tt.want, got.Videos)
			assert.ElementsMatch(t, tt.wantChanged, got.ChangedVideos())
		})
	}
}

func TestFullSyncDue(t *testing.T) {
	t.Parallel()

	// 前回の全件取得は syncedAt に、アイテムが 1 件のときにした
	last := library.NewFullSync(&library.PlaylistMeta{YoutubeID: "PL", ItemCount: 1}, syncedAt)

	tests := []struct {
		name     string
		fullSync library.FullSync
		meta     *library.PlaylistMeta
		now      time.Time
		want     bool
	}{
		{
			name:     "まだ全件取得していなければ取得する",
			fullSync: library.FullSync{},
			meta:     &library.PlaylistMeta{YoutubeID: "PL", ItemCount: 1},
			now:      syncedAt,
			want:     true,
		},
		{
			name:     "アイテム数が同じで FULL_SYNC_INTERVAL が経っていなければ取得しない",
			fullSync: last,
			meta:     &library.PlaylistMeta{YoutubeID: "PL", ItemCount: 1},
			now:      syncedAt.Add(library.FULL_SYNC_INTERVAL - time.Nanosecond),
			want:     false,
		},
		{
			name:     "FULL_SYNC_INTERVAL がちょうど経ったら取得する",
			fullSync: last,
			meta:     &library.PlaylistMeta{YoutubeID: "PL", ItemCount: 1},
			now:      syncedAt.Add(library.FULL_SYNC_INTERVAL),
			want:     true,
		},
		{
			name:     "アイテム数が変わったら、間隔が経っていなくても取得する",
			fullSync: last,
			meta:     &library.PlaylistMeta{YoutubeID: "PL", ItemCount: 2},
			now:      syncedAt.Add(time.Minute),
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fullSync.Due(tt.meta, tt.now)

			assert.Equal(t, tt.want, got)
		})
	}
}
