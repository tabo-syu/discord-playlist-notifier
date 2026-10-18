package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestLibraryServiceSync(t *testing.T) {
	tests := []struct {
		name              string
		storedItems       []*library.PlaylistItem
		storedVideos      []*library.Video
		playlists         []*youtubePlaylist
		itemFails         []library.PlaylistID
		playlistIDs       []library.PlaylistID
		want              map[library.PlaylistID]*library.Snapshot
		wantItems         []*library.PlaylistItem
		wantVideoRequests [][]library.VideoID
	}{
		{
			name: "新しいアイテムを保存し、詳細を持っていない動画だけ詳細を取得する",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(2)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "タイトル",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(2), library.PrivacyPublic),
						fetched("2", "b", hoursAgo(1), library.PrivacyPublic),
					},
				},
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					YoutubeID: "PL",
					Title:     "タイトル",
					Items: []*library.PlaylistItem{
						item("PL", "1", "a", hoursAgo(2)),
						item("PL", "2", "b", hoursAgo(1)),
					},
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
						"b": video("b", library.PrivacyPublic, 0, 0),
					},
				},
			},
			wantItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(2)),
				item("PL", "2", "b", hoursAgo(1)),
			},
			wantVideoRequests: [][]library.VideoID{
				{"b"},
			},
		},
		{
			name: "別のプレイリストの分として保存済みの動画は、詳細を取得しない",
			storedItems: []*library.PlaylistItem{
				item("OTHER", "9", "a", hoursAgo(2)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "タイトル",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
					},
				},
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					YoutubeID: "PL",
					Title:     "タイトル",
					Items: []*library.PlaylistItem{
						item("PL", "1", "a", hoursAgo(1)),
					},
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
					},
				},
			},
			wantItems: []*library.PlaylistItem{
				item("OTHER", "9", "a", hoursAgo(2)),
				item("PL", "1", "a", hoursAgo(1)),
			},
		},
		{
			name: "YouTube から消えたアイテムを除く",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(2)),
				item("PL", "2", "b", hoursAgo(1)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
				video("b", library.PrivacyPublic, 0, 0),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "タイトル",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(2), library.PrivacyPublic),
					},
				},
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					YoutubeID: "PL",
					Title:     "タイトル",
					Items: []*library.PlaylistItem{
						item("PL", "1", "a", hoursAgo(2)),
					},
					// 除いたアイテムの動画も、同期の前に読み込んだので残る
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
						"b": video("b", library.PrivacyPublic, 0, 0),
					},
				},
			},
			wantItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(2)),
			},
		},
		{
			name:        "YouTube から消えたプレイリストは、削除されたものとして返す",
			playlistIDs: []library.PlaylistID{"PL"},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					YoutubeID: "PL",
					Deleted:   true,
				},
			},
		},
		{
			name: "アイテムの取得に失敗したプレイリストは結果から除く",
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "タイトル",
				},
			},
			itemFails:   []library.PlaylistID{"PL"},
			playlistIDs: []library.PlaylistID{"PL"},
			want:        map[library.PlaylistID]*library.Snapshot{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &libraryRepository{items: tt.storedItems, videos: tt.storedVideos}
			yt := &youtube{playlists: tt.playlists, itemFails: tt.itemFails}
			s := application.NewLibraryService(yt, lib, jst, clock)

			got, err := s.Sync(tt.playlistIDs)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantItems, lib.items)
			assert.Equal(t, tt.wantVideoRequests, yt.videoRequests)
		})
	}
}

// クォータを節約するため、前回の同期からアイテム数が変わっておらず、
// FULL_SYNC_INTERVAL が経っていなければアイテムを取得し直さない。
func TestLibraryServiceSyncSkipsFetching(t *testing.T) {
	synced := []*library.FetchedItem{
		fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
	}

	tests := []struct {
		name            string
		elapsed         time.Duration
		items           []*library.FetchedItem
		wantItemFetches int
	}{
		{
			name:            "アイテム数が同じで 1 時間経っていなければ取得しない",
			elapsed:         59 * time.Minute,
			items:           synced,
			wantItemFetches: 0,
		},
		{
			name:            "1 時間経ったら取得する",
			elapsed:         time.Hour,
			items:           synced,
			wantItemFetches: 1,
		},
		{
			name:    "アイテム数が変わったら取得する",
			elapsed: time.Minute,
			items: []*library.FetchedItem{
				fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
				fetched("2", "b", hoursAgo(0), library.PrivacyPublic),
			},
			wantItemFetches: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := now
			playlist := &youtubePlaylist{id: "PL", title: "タイトル", items: synced}
			yt := &youtube{playlists: []*youtubePlaylist{playlist}}
			s := application.NewLibraryService(yt, &libraryRepository{}, jst, func() time.Time { return current })
			// 前回の同期
			_, _ = s.Sync([]library.PlaylistID{"PL"})
			current = now.Add(tt.elapsed)
			playlist.items = tt.items
			yt.itemFetches = 0

			_, err := s.Sync([]library.PlaylistID{"PL"})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantItemFetches, yt.itemFetches)
		})
	}
}

func TestLibraryServiceRefreshVideos(t *testing.T) {
	tests := []struct {
		name              string
		storedItems       []*library.PlaylistItem
		storedVideos      []*library.Video
		views             map[library.VideoID]library.ViewCount
		want              []*library.Video
		wantVideos        []*library.Video
		wantSnapshots     []*library.VideoSnapshot
		wantVideoRequests [][]library.VideoID
	}{
		{
			name: "再生数を更新して履歴に記録し、新しい節目に達した動画を返す",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 90_000, 0),
			},
			views: map[library.VideoID]library.ViewCount{
				"a": 120_000,
			},
			want: []*library.Video{
				video("a", library.PrivacyPublic, 120_000, 100_000),
			},
			wantVideos: []*library.Video{
				video("a", library.PrivacyPublic, 120_000, 100_000),
			},
			wantSnapshots: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 120_000},
			},
			wantVideoRequests: [][]library.VideoID{
				{"a"},
			},
		},
		{
			name: "同じ節目のままなら返さない",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 110_000, 100_000),
			},
			views: map[library.VideoID]library.ViewCount{
				"a": 120_000,
			},
			want: nil,
			wantVideos: []*library.Video{
				video("a", library.PrivacyPublic, 120_000, 100_000),
			},
			wantSnapshots: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 120_000},
			},
			wantVideoRequests: [][]library.VideoID{
				{"a"},
			},
		},
		{
			name: "見られない動画とプレイリストにない動画は取得しない",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
				item("PL", "2", "b", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 100, 0),
				video("b", library.PrivacyPrivate, 100, 0),
				video("removed", library.PrivacyPublic, 100, 0),
			},
			views: map[library.VideoID]library.ViewCount{
				"a": 200,
			},
			want: nil,
			wantVideos: []*library.Video{
				video("a", library.PrivacyPublic, 200, 0),
				video("b", library.PrivacyPrivate, 100, 0),
				video("removed", library.PrivacyPublic, 100, 0),
			},
			wantSnapshots: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 200},
			},
			wantVideoRequests: [][]library.VideoID{
				{"a"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &libraryRepository{items: tt.storedItems, videos: tt.storedVideos}
			yt := &youtube{views: tt.views}
			s := application.NewLibraryService(yt, lib, jst, clock)

			got, err := s.RefreshVideos()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantVideos, lib.videos)
			assert.Equal(t, tt.wantSnapshots, lib.snapshots)
			assert.Equal(t, tt.wantVideoRequests, yt.videoRequests)
		})
	}
}
