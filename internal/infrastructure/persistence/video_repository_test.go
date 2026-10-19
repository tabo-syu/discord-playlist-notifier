package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestVideoRepositoryFind(t *testing.T) {
	tests := []struct {
		name string
		ids  []library.VideoID
		want map[library.VideoID]*library.Video
	}{
		{
			name: "指定した動画を返す",
			ids:  []library.VideoID{"a", "c", "unknown"},
			want: map[library.VideoID]*library.Video{
				"a": video("a", 100),
				"c": video("c", 300),
			},
		},
		{
			name: "行の全てのカラムを動画に写す",
			ids:  []library.VideoID{"full"},
			want: map[library.VideoID]*library.Video{
				"full": fullVideo(),
			},
		},
		{
			name: "動画を指定しなければ空のマップを返す",
			ids:  nil,
			want: map[library.VideoID]*library.Video{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, videoRow("a", 100), videoRow("b", 200), videoRow("c", 300), fullVideoRow())

			got, err := persistence.NewVideoRepository(db).Find(tt.ids)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVideoRepositoryFindInPlaylists(t *testing.T) {
	tests := []struct {
		name        string
		playlistIDs []library.PlaylistID
		want        map[library.VideoID]*library.Video
	}{
		{
			name:        "指定したプレイリストのアイテムの動画を返す",
			playlistIDs: []library.PlaylistID{"PL1", "PL2"},
			want: map[library.VideoID]*library.Video{
				"a": video("a", 100),
				"b": video("b", 200),
			},
		},
		{
			name:        "アイテムのないプレイリストは空のマップを返す",
			playlistIDs: []library.PlaylistID{"EMPTY"},
			want:        map[library.VideoID]*library.Video{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				videoRow("a", 100), videoRow("b", 200), videoRow("c", 300), videoRow("orphan", 400),
				itemRow("PL1", "1", "a", hoursAgo(1)), itemRow("PL2", "2", "b", hoursAgo(1)), itemRow("OTHER", "3", "c", hoursAgo(1)),
			)

			got, err := persistence.NewVideoRepository(db).FindInPlaylists(tt.playlistIDs...)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVideoRepositoryFindListed(t *testing.T) {
	tests := []struct {
		name string
		want map[library.VideoID]*library.Video
	}{
		{
			name: "どれかのプレイリストにある動画だけを返す",
			want: map[library.VideoID]*library.Video{
				"a": video("a", 100),
				"b": video("b", 200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				videoRow("a", 100), videoRow("b", 200), videoRow("orphan", 300),
				itemRow("PL1", "1", "a", hoursAgo(1)), itemRow("PL2", "2", "b", hoursAgo(1)), itemRow("PL2", "3", "a", hoursAgo(1)),
			)

			got, err := persistence.NewVideoRepository(db).FindListed()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVideoRepositorySave(t *testing.T) {
	tests := []struct {
		name   string
		videos []*library.Video
		want   []*persistence.VideoRecord
	}{
		{
			// id と created_at は保存済みの行のまま、updated_at は保存した時刻にする。
			// 保存済みの行は詳細を取得していないので、details_fetched も上書きされることを確かめる
			name: "保存済みの動画は全てのカラムを上書きし、新しい動画は追加する",
			videos: []*library.Video{
				fullVideo(),
				video("b", 200),
			},
			want: []*persistence.VideoRecord{
				storedVideoRow(1, fullVideoRow()),
				storedVideoRow(3, videoRow("b", 200)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.VideoRecord{YoutubeID: "full", PrivacyStatus: "public"})

			err := persistence.NewVideoRepository(db).Save(tt.videos)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[persistence.VideoRecord](t, db))
		})
	}
}
