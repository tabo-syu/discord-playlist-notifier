package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestCatalogContents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  []*library.PlaylistVideo
	}{
		{
			name: "アイテムを動画と一緒に返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlistOf("PL1", listed("PL1", "1", "a", 1)), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlistOf("PL2", listed("PL2", "2", "b", 2)), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0),
					"b": video("b", library.PrivacyPrivate, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: listed("PL1", "1", "a", 1), Video: video("a", library.PrivacyPublic, 0)},
				{PlaylistID: "PL2", Item: listed("PL2", "2", "b", 2), Video: video("b", library.PrivacyPrivate, 0)},
			},
		},
		{
			name: "プレイリストをまたいで、追加された時刻の順に並べる",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlistOf("PL1", listed("PL1", "1", "a", 2)), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlistOf("PL2", listed("PL2", "2", "b", 1)), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0),
					"b": video("b", library.PrivacyPublic, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL2", Item: listed("PL2", "2", "b", 1), Video: video("b", library.PrivacyPublic, 0)},
				{PlaylistID: "PL1", Item: listed("PL1", "1", "a", 2), Video: video("a", library.PrivacyPublic, 0)},
			},
		},
		{
			name: "動画が保存されていないアイテムは除く",
			setup: func(m *deps) {
				m.playlists.EXPECT().Find(library.PlaylistID("PL1")).Return(playlistOf("PL1",
					listed("PL1", "1", "a", 1),
					listed("PL1", "2", "missing", 2),
				), nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL2")).Return(playlistOf("PL2"), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL1", Item: listed("PL1", "1", "a", 1), Video: video("a", library.PrivacyPublic, 0)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			c := library.NewCatalog(m.playlists, m.videos)

			// 同じプレイリストを 2 回指定しても、1 回だけ読む
			got, err := c.Contents([]library.PlaylistID{"PL1", "PL2", "PL1"})

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCatalogPlaylistsContaining(t *testing.T) {
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
					playlistOf("PL1", listed("PL1", "1", "a", 1)),
					playlistOf("PL2", listed("PL2", "2", "a", 2), listed("PL2", "3", "b", 3)),
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
					playlistOf("PL", listed("PL", "1", "a", 1), listed("PL", "2", "a", 2)),
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
			c := library.NewCatalog(m.playlists, m.videos)

			// 同じ動画を 2 回指定しても、プレイリストは 1 回だけ返す
			got, err := c.PlaylistsContaining([]library.VideoID{"a", "b", "a"})

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCatalogFindVideo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *deps)
		want    *library.Video
		wantErr error
	}{
		{
			name: "範囲のプレイリストのどれかにある動画を返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlistOf("OTHER", listed("OTHER", "1", "a", 1)),
					playlistOf("PL2", listed("PL2", "2", "a", 2)),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 300),
				}, nil)
			},
			want:    video("a", library.PrivacyPublic, 300),
			wantErr: nil,
		},
		{
			name: "範囲の外のプレイリストにしかない動画は見せない",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlistOf("OTHER", listed("OTHER", "1", "a", 1)),
				}, nil)
			},
			want:    nil,
			wantErr: library.ErrVideoNotFound,
		},
		{
			name: "詳細を取得していない動画は見せない",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlistOf("PL1", listed("PL1", "1", "a", 1)),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": {YoutubeID: "a", PrivacyStatus: library.PrivacyPublic},
				}, nil)
			},
			want:    nil,
			wantErr: library.ErrVideoNotFound,
		},
		{
			name: "保存されていない動画は見せない",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlistOf("PL1", listed("PL1", "1", "a", 1)),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{}, nil)
			},
			want:    nil,
			wantErr: library.ErrVideoNotFound,
		},
		{
			name: "プレイリストを読めなければ、そのエラーを返す",
			setup: func(m *deps) {
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return(nil, errMock)
			},
			want:    nil,
			wantErr: errMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			c := library.NewCatalog(m.playlists, m.videos)

			got, err := c.FindVideo([]library.PlaylistID{"PL1", "PL2"}, "a")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
