package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
		// Migrate の後の playlists の全ての行
		wantPlaylists []*persistence.SubscriptionRecord
		// Migrate の後の videos の全ての行
		wantVideos []*persistence.VideoRecord
	}{
		{
			name: "最新のスキーマに対して実行しても失敗しない",
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
			wantPlaylists: []*persistence.SubscriptionRecord{},
			wantVideos:    []*persistence.VideoRecord{},
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
			wantPlaylists: []*persistence.SubscriptionRecord{},
			wantVideos:    []*persistence.VideoRecord{},
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
			wantPlaylists: []*persistence.SubscriptionRecord{},
			wantVideos:    []*persistence.VideoRecord{},
		},
		{
			name: "notified_until がなかった行は、以前の基準時刻の updated_at で埋める",
			before: []string{
				"ALTER TABLE playlists DROP COLUMN notified_until",
				"INSERT INTO guilds (discord_id) VALUES ('guild')",
				"INSERT INTO playlists (created_at, updated_at, youtube_id, send_channel_id, title, pick_interval, guild_id) VALUES ('2026-10-18 12:00:00+00', '2026-10-18 06:00:00+00', 'PL', 'ch', 'PL', '', 1)",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
			wantPlaylists: []*persistence.SubscriptionRecord{
				{ID: 1, CreatedAt: now, UpdatedAt: hoursAgo(6), YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1, NotifiedUntil: hoursAgo(6)},
			},
			wantVideos: []*persistence.VideoRecord{},
		},
		{
			name: "notified_until のある行は updated_at で上書きしない",
			before: []string{
				"INSERT INTO guilds (discord_id) VALUES ('guild')",
				"INSERT INTO playlists (created_at, updated_at, youtube_id, send_channel_id, title, pick_interval, guild_id, notified_until) VALUES ('2026-10-18 12:00:00+00', '2026-10-18 12:00:00+00', 'PL', 'ch', 'PL', 'weekly', 1, '2026-10-18 06:00:00+00')",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
			wantPlaylists: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 1, "PL", "PL"),
			},
			wantVideos: []*persistence.VideoRecord{},
		},
		{
			name: "details_fetched がなかった行は、以前の判断のとおりタイトルの有無で埋める",
			before: []string{
				"ALTER TABLE videos DROP COLUMN details_fetched",
				"INSERT INTO videos (created_at, updated_at, youtube_id, title, views, privacy_status, published_at, likes, comments, channel_youtube_id, duration, category_id) VALUES ('2026-10-18 12:00:00+00', '2026-10-18 12:00:00+00', 'fetched', '曲', 0, 'public', '0001-01-01 00:00:00+00', 0, 0, '', 0, '')",
				"INSERT INTO videos (created_at, updated_at, youtube_id, title, views, privacy_status, published_at, likes, comments, channel_youtube_id, duration, category_id) VALUES ('2026-10-18 12:00:00+00', '2026-10-18 12:00:00+00', 'unfetched', '', 0, 'public', '0001-01-01 00:00:00+00', 0, 0, '', 0, '')",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
			wantPlaylists: []*persistence.SubscriptionRecord{},
			wantVideos: []*persistence.VideoRecord{
				{ID: 1, CreatedAt: now, UpdatedAt: now, YoutubeID: "fetched", DetailsFetched: true, Title: "曲", PrivacyStatus: "public"},
				{ID: 2, CreatedAt: now, UpdatedAt: now, YoutubeID: "unfetched", DetailsFetched: false, Title: "", PrivacyStatus: "public"},
			},
		},
		{
			name: "details_fetched のある行はタイトルで上書きしない",
			before: []string{
				"INSERT INTO videos (created_at, updated_at, youtube_id, title, views, privacy_status, published_at, likes, comments, channel_youtube_id, duration, category_id) VALUES ('2026-10-18 12:00:00+00', '2026-10-18 12:00:00+00', 'fetched', '曲', 0, 'public', '0001-01-01 00:00:00+00', 0, 0, '', 0, '')",
				"UPDATE videos SET details_fetched = false",
			},
			want: schema{
				VideosHasYoutubeID:          true,
				VideosHasPlaylistID:         false,
				PlaylistsHasGuildForeignKey: true,
			},
			wantPlaylists: []*persistence.SubscriptionRecord{},
			wantVideos: []*persistence.VideoRecord{
				{ID: 1, CreatedAt: now, UpdatedAt: now, YoutubeID: "fetched", DetailsFetched: false, Title: "曲", PrivacyStatus: "public"},
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
				PlaylistsHasGuildForeignKey: db.Migrator().HasConstraint(&persistence.SubscriptionRecord{}, "fk_guilds_playlists"),
			})
			assert.Equal(t, tt.wantPlaylists, rows[persistence.SubscriptionRecord](t, db))
			assert.Equal(t, tt.wantVideos, rows[persistence.VideoRecord](t, db))
		})
	}
}
