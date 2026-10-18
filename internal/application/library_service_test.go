package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestLibraryServiceSync(t *testing.T) {
	tests := []struct {
		name              string
		storedItems       []*library.PlaylistItem
		storedVideos      []*library.Video
		playlists         []*fake.Playlist
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
				},
			},
			itemFails:   []library.PlaylistID{"PL"},
			playlistIDs: []library.PlaylistID{"PL"},
			want:        map[library.PlaylistID]*library.Snapshot{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			yt := &fake.YouTube{Playlists: tt.playlists, ItemFails: tt.itemFails}
			s := application.NewLibraryService(yt, lib, jst, clock)

			got, err := s.Sync(tt.playlistIDs)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantItems, lib.Items)
			assert.Equal(t, tt.wantVideoRequests, yt.VideoRequests)
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
			playlist := &fake.Playlist{ID: "PL", Title: "タイトル", Items: synced}
			yt := &fake.YouTube{Playlists: []*fake.Playlist{playlist}}
			s := application.NewLibraryService(yt, &fake.LibraryRepository{}, jst, func() time.Time { return current })
			// 前回の同期
			_, _ = s.Sync([]library.PlaylistID{"PL"})
			current = now.Add(tt.elapsed)
			playlist.Items = tt.items
			yt.ItemFetches = 0

			_, err := s.Sync([]library.PlaylistID{"PL"})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantItemFetches, yt.ItemFetches)
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
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			yt := &fake.YouTube{Views: tt.views}
			s := application.NewLibraryService(yt, lib, jst, clock)

			got, err := s.RefreshVideos()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantVideos, lib.Videos)
			assert.Equal(t, tt.wantSnapshots, lib.Snapshots)
			assert.Equal(t, tt.wantVideoRequests, yt.VideoRequests)
		})
	}
}

func TestLibraryServiceContents(t *testing.T) {
	a := video("a", library.PrivacyPublic, 0, 0)
	b := video("b", library.PrivacyPrivate, 0, 0)

	tests := []struct {
		name         string
		storedItems  []*library.PlaylistItem
		storedVideos []*library.Video
		playlistIDs  []library.PlaylistID
		want         []*library.PlaylistVideo
	}{
		{
			name: "指定したプレイリストのアイテムを、追加された順に動画と一緒に返す",
			storedItems: []*library.PlaylistItem{
				item("PL2", "2", "b", hoursAgo(1)),
				item("PL1", "1", "a", hoursAgo(2)),
				item("OTHER", "3", "a", hoursAgo(3)),
			},
			storedVideos: []*library.Video{
				a,
				b,
			},
			playlistIDs: []library.PlaylistID{"PL1", "PL2"},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(2)), Video: a},
				{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: b},
			},
		},
		{
			name: "動画が保存されていないアイテムは除く",
			storedItems: []*library.PlaylistItem{
				item("PL1", "1", "a", hoursAgo(2)),
				item("PL1", "2", "missing", hoursAgo(1)),
			},
			storedVideos: []*library.Video{
				a,
			},
			playlistIDs: []library.PlaylistID{"PL1"},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(2)), Video: a},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			s := application.NewLibraryService(&fake.YouTube{}, lib, jst, clock)

			got, err := s.Contents(tt.playlistIDs)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServicePlaylistsContaining(t *testing.T) {
	tests := []struct {
		name        string
		storedItems []*library.PlaylistItem
		videoIDs    []library.VideoID
		want        map[library.VideoID][]library.PlaylistID
	}{
		{
			name: "動画ごとに、それを含むプレイリストを返す",
			storedItems: []*library.PlaylistItem{
				item("PL1", "1", "a", hoursAgo(3)),
				item("PL2", "2", "a", hoursAgo(2)),
				item("PL2", "3", "b", hoursAgo(1)),
				item("PL3", "4", "c", hoursAgo(1)),
			},
			videoIDs: []library.VideoID{"a", "b"},
			want: map[library.VideoID][]library.PlaylistID{
				"a": {"PL1", "PL2"},
				"b": {"PL2"},
			},
		},
		{
			name:     "どのプレイリストにもなければ空のマップを返す",
			videoIDs: []library.VideoID{"a"},
			want:     map[library.VideoID][]library.PlaylistID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := application.NewLibraryService(&fake.YouTube{}, &fake.LibraryRepository{Items: tt.storedItems}, jst, clock)

			got, err := s.PlaylistsContaining(tt.videoIDs)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServiceHistory(t *testing.T) {
	tests := []struct {
		name            string
		storedItems     []*library.PlaylistItem
		storedVideos    []*library.Video
		storedSnapshots []*library.VideoSnapshot
		playlistIDs     []library.PlaylistID
		want            *library.VideoHistory
		wantErr         error
	}{
		{
			name: "指定した時刻以降に記録された状態を、古い順に返す",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 300, 0),
			},
			storedSnapshots: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 300},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(72), Views: 100},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 200},
				{VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 999},
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want: &library.VideoHistory{
				Video: video("a", library.PrivacyPublic, 300, 0),
				Snapshots: []*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 200},
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 300},
				},
			},
		},
		{
			name: "指定したプレイリストにない動画は返さない",
			storedItems: []*library.PlaylistItem{
				item("OTHER", "1", "a", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 300, 0),
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want:        nil,
			wantErr:     domain.ErrDBRecordNotFound,
		},
		{
			name: "詳細を取得していない動画は返さない",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				{YoutubeID: "a", PrivacyStatus: library.PrivacyPublic},
			},
			playlistIDs: []library.PlaylistID{"PL"},
			want:        nil,
			wantErr:     domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos, Snapshots: tt.storedSnapshots}
			s := application.NewLibraryService(&fake.YouTube{}, lib, jst, clock)

			got, err := s.History(tt.playlistIDs, "a", hoursAgo(48))

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServiceTrending(t *testing.T) {
	a := video("a", library.PrivacyPublic, 0, 0)
	b := video("b", library.PrivacyPublic, 0, 0)

	tests := []struct {
		name            string
		storedItems     []*library.PlaylistItem
		storedSnapshots []*library.VideoSnapshot
		want            *library.Trending
	}{
		{
			name: "指定した時刻以降に伸びた動画を、その間の状態と一緒に返す",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(100)),
				item("PL", "2", "b", hoursAgo(100)),
			},
			storedSnapshots: []*library.VideoSnapshot{
				// 指定した時刻より前の記録は使わない
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(72), Views: 0},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 150},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
				{VideoYoutubeID: "b", RecordedAt: hoursAgo(36), Views: 100},
				{VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 400},
			},
			want: &library.Trending{
				From: hoursAgo(36),
				To:   hoursAgo(12),
				Videos: []*library.TrendingVideo{
					{
						Video:    b,
						Increase: 300,
						Snapshots: []*library.VideoSnapshot{
							{VideoYoutubeID: "b", RecordedAt: hoursAgo(36), Views: 100},
							{VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 400},
						},
					},
					{
						Video:    a,
						Increase: 100,
						Snapshots: []*library.VideoSnapshot{
							{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
							{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 150},
							{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
						},
					},
				},
			},
		},
		{
			name: "他のプレイリストの動画は含めない",
			storedItems: []*library.PlaylistItem{
				item("OTHER", "1", "a", hoursAgo(100)),
			},
			storedSnapshots: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
			},
			want: &library.Trending{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: []*library.Video{a, b}, Snapshots: tt.storedSnapshots}
			s := application.NewLibraryService(&fake.YouTube{}, lib, jst, clock)

			got, err := s.Trending("PL", hoursAgo(48))

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
