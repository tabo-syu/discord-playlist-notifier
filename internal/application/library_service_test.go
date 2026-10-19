package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestLibraryServiceSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  map[library.PlaylistID]*library.Snapshot
		// 動画を保存できたときだけ返す
		wantHidden []*library.VideoHidden
	}{
		{
			name: "新しいアイテムを保存し、詳細を持っていない動画だけ詳細を取得する",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 2},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(2), library.PrivacyPublic),
					fetched("2", "b", hoursAgo(1), library.PrivacyPublic),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"b"}).Return(map[library.VideoID]*library.Video{}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"b"}).Return([]*library.Video{
					{YoutubeID: "b", Title: "b"},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				saved := m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
					item("PL", "2", "b", hoursAgo(1)),
				)).Return(nil).Call
				// プレイリストを先に保存する（LibraryService.syncPlaylist のコメント）
				m.videos.EXPECT().Save([]*library.Video{
					video("b", library.PrivacyPublic, 0, 0),
				}).Return(nil).NotBefore(saved)
			},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					Playlist: playlist("PL",
						item("PL", "1", "a", hoursAgo(2)),
						item("PL", "2", "b", hoursAgo(1)),
					),
					Title: "タイトル",
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
						"b": video("b", library.PrivacyPublic, 0, 0),
					},
				},
			},
		},
		{
			name: "別のプレイリストの分として保存済みの動画は、詳細を取得しない",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL"), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(1)),
				)).Return(nil)
				m.videos.EXPECT().Save([]*library.Video(nil)).Return(nil)
			},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					Playlist: playlist("PL",
						item("PL", "1", "a", hoursAgo(1)),
					),
					Title: "タイトル",
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
					},
				},
			},
		},
		{
			name: "YouTube から消えたアイテムを除く",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
					item("PL", "2", "b", hoursAgo(1)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(2), library.PrivacyPublic),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
					"b": video("b", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
				m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				)).Return(nil)
				m.videos.EXPECT().Save([]*library.Video(nil)).Return(nil)
			},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					Playlist: playlist("PL",
						item("PL", "1", "a", hoursAgo(2)),
					),
					Title: "タイトル",
					// 除いたアイテムの動画も、同期の前に読み込んだので残る
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPublic, 0, 0),
						"b": video("b", library.PrivacyPublic, 0, 0),
					},
				},
			},
		},
		{
			name: "公開状態を保存し、見られなくなった動画のイベントを返す",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(2), library.PrivacyPrivate),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
				saved := m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				)).Return(nil).Call
				// 公開状態の変化はプレイリストの後で保存する。逆だと、プレイリストの保存に
				// 失敗したときに、次の同期で非公開になったことを検知できない
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPrivate, 0, 0),
				}).Return(nil).NotBefore(saved)
			},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					Playlist: playlist("PL",
						item("PL", "1", "a", hoursAgo(2)),
					),
					Title: "タイトル",
					Videos: map[library.VideoID]*library.Video{
						"a": video("a", library.PrivacyPrivate, 0, 0),
					},
				},
			},
			wantHidden: []*library.VideoHidden{
				{Video: video("a", library.PrivacyPrivate, 0, 0)},
			},
		},
		{
			name: "プレイリストの保存に失敗したら、動画を保存せず結果から除く",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(2), library.PrivacyPrivate),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
				m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				)).Return(errMock)
			},
			want: map[library.PlaylistID]*library.Snapshot{},
		},
		{
			name: "動画の保存に失敗したら、結果から除き、イベントも返さない",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(2), library.PrivacyPrivate),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
				m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(2)),
				)).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPrivate, 0, 0),
				}).Return(errMock)
			},
			want: map[library.PlaylistID]*library.Snapshot{},
		},
		{
			name: "YouTube から消えたプレイリストは、削除されたものとして返す",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{}, nil)
			},
			want: map[library.PlaylistID]*library.Snapshot{
				"PL": {
					Playlist: playlist("PL"),
					Deleted:  true,
				},
			},
		},
		{
			name: "アイテムの取得に失敗したプレイリストは結果から除く",
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL"), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return(nil, errMock)
			},
			want: map[library.PlaylistID]*library.Snapshot{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			got, hidden, err := s.Sync([]library.PlaylistID{"PL"})

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantHidden, hidden)
		})
	}
}

// クォータを節約するため、前回の同期からアイテム数が変わっておらず、
// FULL_SYNC_INTERVAL が経っていなければアイテムを取得し直さない。
// 2 回同期し、2 回目にアイテムを取得したかを、YouTube を呼んだ回数で確かめる。
func TestLibraryServiceSyncSkipsFetching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		elapsed time.Duration
		setup   func(m *deps)
	}{
		{
			name:    "アイテム数が同じで 1 時間経っていなければ、2 回目は取得しない",
			elapsed: 59 * time.Minute,
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil).Times(2)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(1)),
				), nil).Times(2)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
				}, nil).Once()
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil).Times(2)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil).Once()
				m.playlists.EXPECT().Save(playlist("PL", item("PL", "1", "a", hoursAgo(1)))).Return(nil).Once()
				m.videos.EXPECT().Save([]*library.Video(nil)).Return(nil).Once()
			},
		},
		{
			name:    "1 時間経ったら、2 回目も取得する",
			elapsed: time.Hour,
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil).Times(2)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(1)),
				), nil).Times(2)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
				}, nil).Times(2)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil).Times(2)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil).Times(2)
				m.playlists.EXPECT().Save(playlist("PL", item("PL", "1", "a", hoursAgo(1)))).Return(nil).Times(2)
				m.videos.EXPECT().Save([]*library.Video(nil)).Return(nil).Times(2)
			},
		},
		{
			name:    "アイテム数が変わったら、2 回目も取得する",
			elapsed: time.Minute,
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 1},
				}, nil).Once()
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": {YoutubeID: "PL", Title: "タイトル", ItemCount: 2},
				}, nil).Once()
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(1)),
				), nil).Times(2)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
				}, nil).Times(2)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil).Times(2)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil).Times(2)
				m.playlists.EXPECT().Save(playlist("PL", item("PL", "1", "a", hoursAgo(1)))).Return(nil).Times(2)
				m.videos.EXPECT().Save([]*library.Video(nil)).Return(nil).Times(2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)

			current := now
			s := newLibraryService(m, func() time.Time { return current })
			// 前回の同期
			_, _, err := s.Sync([]library.PlaylistID{"PL"})
			assert.NoError(t, err)

			current = now.Add(tt.elapsed)

			_, _, err = s.Sync([]library.PlaylistID{"PL"})

			assert.NoError(t, err)
		})
	}
}

func TestLibraryServiceRefreshVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  []*library.MilestoneReached
	}{
		{
			name: "再生数を更新して履歴に記録し、新しい節目に達した動画のイベントを返す",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 90_000, 0),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 120_000},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPublic, 120_000, 100_000),
				}).Return(nil)
				m.snapshots.EXPECT().Add([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 120_000},
				}).Return(nil)
			},
			want: []*library.MilestoneReached{
				{Video: video("a", library.PrivacyPublic, 120_000, 100_000), Milestone: 100_000},
			},
		},
		{
			name: "同じ節目のままなら返さない",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 110_000, 100_000),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 120_000},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPublic, 120_000, 100_000),
				}).Return(nil)
				m.snapshots.EXPECT().Add([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 120_000},
				}).Return(nil)
			},
			want: nil,
		},
		{
			name: "見られない動画は取得しない",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 100, 0),
					"b": video("b", library.PrivacyPrivate, 100, 0),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 200},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPublic, 200, 0),
				}).Return(nil)
				m.snapshots.EXPECT().Add([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 200},
				}).Return(nil)
			},
			want: nil,
		},
		{
			name: "履歴の記録に失敗しても、節目に達した動画のイベントは返す",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 90_000, 0),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 120_000},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPublic, 120_000, 100_000),
				}).Return(nil)
				m.snapshots.EXPECT().Add([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: now, Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 120_000},
				}).Return(errMock)
			},
			want: []*library.MilestoneReached{
				{Video: video("a", library.PrivacyPublic, 120_000, 100_000), Milestone: 100_000},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			got, err := s.RefreshVideos()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServiceContents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  []*library.PlaylistVideo
	}{
		{
			name: "アイテムを動画と一緒に返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlist("PL1", item("PL1", "1", "a", hoursAgo(2))), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlist("PL2", item("PL2", "2", "b", hoursAgo(1))), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
					"b": video("b", library.PrivacyPrivate, 0, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(2)), Video: video("a", library.PrivacyPublic, 0, 0)},
				{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPrivate, 0, 0)},
			},
		},
		{
			name: "プレイリストをまたいで、追加された時刻の順に並べる",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlist("PL1", item("PL1", "1", "a", hoursAgo(1))), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlist("PL2", item("PL2", "2", "b", hoursAgo(2))), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
					"b": video("b", library.PrivacyPublic, 0, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(2)), Video: video("b", library.PrivacyPublic, 0, 0)},
				{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
			},
		},
		{
			name: "動画が保存されていないアイテムは除く",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlist("PL1",
					item("PL1", "1", "a", hoursAgo(2)),
					item("PL1", "2", "missing", hoursAgo(1)),
				), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlist("PL2"), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(2)), Video: video("a", library.PrivacyPublic, 0, 0)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			// 同じプレイリストを 2 回指定しても、1 回だけ読む
			got, err := s.Contents([]library.PlaylistID{"PL1", "PL2", "PL1"})

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServicePlaylistsContaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  map[library.VideoID][]library.PlaylistID
	}{
		{
			name: "動画ごとに、それを含むプレイリストを返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a", "b", "a"}).Return([]*library.Playlist{
					playlist("PL1", item("PL1", "1", "a", hoursAgo(3))),
					playlist("PL2", item("PL2", "2", "a", hoursAgo(2)), item("PL2", "3", "b", hoursAgo(1))),
				}, nil)
			},
			want: map[library.VideoID][]library.PlaylistID{
				"a": {"PL1", "PL2"},
				"b": {"PL2"},
			},
		},
		{
			name: "1 つのプレイリストに 2 回入っている動画でも、プレイリストは 1 回だけ返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a", "b", "a"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "a", hoursAgo(2)), item("PL", "2", "a", hoursAgo(1))),
				}, nil)
			},
			want: map[library.VideoID][]library.PlaylistID{
				"a": {"PL"},
			},
		},
		{
			name: "どのプレイリストにもなければ空のマップを返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a", "b", "a"}).Return([]*library.Playlist{}, nil)
			},
			want: map[library.VideoID][]library.PlaylistID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			// 同じ動画を 2 回指定しても、プレイリストは 1 回だけ返す
			got, err := s.PlaylistsContaining([]library.VideoID{"a", "b", "a"})

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServiceHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *deps)
		want    *library.VideoHistory
		wantErr error
	}{
		{
			name: "指定したプレイリストにある動画の、指定した時刻以降の状態を返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "a", hoursAgo(100))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 300, 0),
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("a"), hoursAgo(48)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 200},
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 300},
				}, nil)
			},
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
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("OTHER", item("OTHER", "1", "a", hoursAgo(100))),
				}, nil)
			},
			want:    nil,
			wantErr: library.ErrVideoNotFound,
		},
		{
			name: "詳細を取得していない動画は返さない",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "a", hoursAgo(100))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": {YoutubeID: "a", PrivacyStatus: library.PrivacyPublic},
				}, nil)
			},
			want:    nil,
			wantErr: library.ErrVideoNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			got, err := s.History([]library.PlaylistID{"PL"}, "a", hoursAgo(48))

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryServiceTrending(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  *library.Trending
	}{
		{
			name: "指定した時刻以降に伸びた動画を、その間の状態と一緒に返す",
			setup: func(m *deps) {
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
					"b": video("b", library.PrivacyPublic, 0, 0),
				}, nil)
				m.snapshots.EXPECT().FindEarliest(sameElements([]library.VideoID{"a", "b"}), hoursAgo(48)).Return(map[library.VideoID]*library.VideoSnapshot{
					"a": {VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
					"b": {VideoYoutubeID: "b", RecordedAt: hoursAgo(36), Views: 100},
				}, nil)
				m.snapshots.EXPECT().FindLatest(sameElements([]library.VideoID{"a", "b"}), hoursAgo(48)).Return(map[library.VideoID]*library.VideoSnapshot{
					"a": {VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
					"b": {VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 400},
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("b"), hoursAgo(48)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "b", RecordedAt: hoursAgo(36), Views: 100},
					{VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 400},
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("a"), hoursAgo(48)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
				}, nil)
			},
			want: &library.Trending{
				From: hoursAgo(36),
				To:   hoursAgo(12),
				Videos: []*library.TrendingVideo{
					{
						Video:    video("b", library.PrivacyPublic, 0, 0),
						Increase: 300,
						Snapshots: []*library.VideoSnapshot{
							{VideoYoutubeID: "b", RecordedAt: hoursAgo(36), Views: 100},
							{VideoYoutubeID: "b", RecordedAt: hoursAgo(12), Views: 400},
						},
					},
					{
						Video:    video("a", library.PrivacyPublic, 0, 0),
						Increase: 100,
						Snapshots: []*library.VideoSnapshot{
							{VideoYoutubeID: "a", RecordedAt: hoursAgo(36), Views: 100},
							{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newLibraryService(m, clock)

			got, err := s.Trending("PL", hoursAgo(48))

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
