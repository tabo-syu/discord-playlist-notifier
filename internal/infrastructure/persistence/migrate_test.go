package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

// schema は Migrate の結果のうち、確かめたいスキーマの状態。
type schema struct {
	// videos が今の形（動画ごとに 1 行）か
	VideosHasYoutubeID bool
	// videos が以前の形（通知した動画の履歴）か
	VideosHasPlaylistID bool
	// playlists.guild_id の外部キーがあるか
	PlaylistsHasGuildForeignKey bool
}

func TestMigrate(t *testing.T) {
	tests := []struct {
		name string
		// Migrate の前に実行する SQL
		before []string
		want   schema
	}{
		{
			name: "最新のスキーマに対して実行しても失敗しない",
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
		},
		{
			name: "以前の形の videos テーブルを消して作り直す",
			before: []string{
				"DROP TABLE videos",
				"CREATE TABLE videos (id bigserial PRIMARY KEY, playlist_id bigint, title text)",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
		},
		{
			name: "playlists の外部キーがなければ作る",
			before: []string{
				"ALTER TABLE playlists DROP CONSTRAINT fk_guilds_playlists",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			execute(t, db, tt.before...)

			err := persistence.Migrate(db)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, schema{
				VideosHasYoutubeID:          db.Migrator().HasColumn("videos", "youtube_id"),
				VideosHasPlaylistID:         db.Migrator().HasColumn("videos", "playlist_id"),
				PlaylistsHasGuildForeignKey: db.Migrator().HasConstraint(&subscription.Subscription{}, "fk_guilds_playlists"),
			})
		})
	}
}
