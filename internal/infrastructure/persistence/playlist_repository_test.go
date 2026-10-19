package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestPlaylistRepositoryFind(t *testing.T) {
	tests := []struct {
		name string
		id   library.PlaylistID
		want *library.Playlist
	}{
		{
			name: "プレイリストのアイテムを、追加された順に返す",
			id:   "PL",
			want: playlist("PL",
				item("PL", "3", "c", hoursAgo(3)),
				item("PL", "1", "a", hoursAgo(2)),
			),
		},
		{
			name: "アイテムがなければ、アイテムのないプレイリストを返す",
			id:   "EMPTY",
			want: &library.Playlist{YoutubeID: "EMPTY", Items: []*library.PlaylistItem{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				itemRow("PL", "1", "a", hoursAgo(2)),
				itemRow("OTHER", "2", "b", hoursAgo(1)),
				itemRow("PL", "3", "c", hoursAgo(3)),
			)

			got, err := persistence.NewPlaylistRepository(db).Find(tt.id)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlaylistRepositoryFindContaining(t *testing.T) {
	tests := []struct {
		name     string
		videoIDs []library.VideoID
		want     []*library.Playlist
	}{
		{
			name:     "指定した動画を含むプレイリストを、全てのアイテムと一緒に返す",
			videoIDs: []library.VideoID{"a"},
			want: []*library.Playlist{
				playlist("PL1", item("PL1", "2", "b", hoursAgo(3)), item("PL1", "1", "a", hoursAgo(2))),
				playlist("PL2", item("PL2", "3", "a", hoursAgo(1))),
			},
		},
		{
			name:     "どのプレイリストにもない動画は空のスライスを返す",
			videoIDs: []library.VideoID{"unknown"},
			want:     nil,
		},
		{
			name:     "動画を指定しなければ空のスライスを返す",
			videoIDs: nil,
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				itemRow("PL1", "1", "a", hoursAgo(2)),
				itemRow("PL1", "2", "b", hoursAgo(3)),
				itemRow("PL2", "3", "a", hoursAgo(1)),
				itemRow("OTHER", "4", "c", hoursAgo(1)),
			)

			got, err := persistence.NewPlaylistRepository(db).FindContaining(tt.videoIDs)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlaylistRepositorySave(t *testing.T) {
	tests := []struct {
		name     string
		playlist *library.Playlist
		want     []*persistence.PlaylistItemRecord
	}{
		{
			// 残ったアイテムの行は書き換えないので、ID も作った時刻もそのまま
			name: "外れたアイテムを削除し、増えたアイテムを追加する",
			playlist: playlist("PL",
				item("PL", "1", "a", hoursAgo(3)),
				item("PL", "4", "d", hoursAgo(1)),
			),
			want: []*persistence.PlaylistItemRecord{
				storedItemRow(1, itemRow("PL", "1", "a", hoursAgo(3))),
				storedItemRow(3, itemRow("OTHER", "3", "c", hoursAgo(2))),
				storedItemRow(4, itemRow("PL", "4", "d", hoursAgo(1))),
			},
		},
		{
			name:     "アイテムがなくなれば、そのプレイリストの行を全て削除する",
			playlist: playlist("PL"),
			want: []*persistence.PlaylistItemRecord{
				storedItemRow(3, itemRow("OTHER", "3", "c", hoursAgo(2))),
			},
		},
		{
			name: "何も変わらなければ何もしない",
			playlist: playlist("PL",
				item("PL", "1", "a", hoursAgo(3)),
				item("PL", "2", "b", hoursAgo(2)),
			),
			want: []*persistence.PlaylistItemRecord{
				storedItemRow(1, itemRow("PL", "1", "a", hoursAgo(3))),
				storedItemRow(2, itemRow("PL", "2", "b", hoursAgo(2))),
				storedItemRow(3, itemRow("OTHER", "3", "c", hoursAgo(2))),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				itemRow("PL", "1", "a", hoursAgo(3)),
				itemRow("PL", "2", "b", hoursAgo(2)),
				itemRow("OTHER", "3", "c", hoursAgo(2)),
			)

			err := persistence.NewPlaylistRepository(db).Save(tt.playlist)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[persistence.PlaylistItemRecord](t, db))
		})
	}
}
