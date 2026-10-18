package persistence_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

// now はテストの現在時刻。GORM が CreatedAt・UpdatedAt に書く時刻をこれに固定する。
var now = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

// hoursAgo は now の hours 時間前の時刻を返す。
func hoursAgo(hours int) time.Time {
	return now.Add(-time.Duration(hours) * time.Hour)
}

// dsn はテスト用の PostgreSQL に接続する DSN。TestMain が設定する。
var dsn string

// TestMain は Testcontainers で PostgreSQL を 1 つ立て、スキーマを作ってから全てのテストを
// 走らせる。Docker が必要。go test -short では DB のテストを飛ばす。
func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	flag.Parse()
	// テストの出力を読みやすくするため、マイグレーションのログを捨てる
	log.SetOutput(io.Discard)

	// Testcontainers のログを捨てる
	tclog.SetDefault(tclog.NewNoopLogger())

	if testing.Short() {
		fmt.Println("skipping database tests in short mode")

		return 0
	}

	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:14.4-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Println("could not start PostgreSQL (Docker is required):", err)

		return 1
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Println("could not terminate PostgreSQL:", err)
		}
	}()

	dsn, err = container.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
	if err != nil {
		fmt.Println("could not get the connection string:", err)

		return 1
	}

	db, err := open()
	if err != nil {
		fmt.Println("could not connect to PostgreSQL:", err)

		return 1
	}

	if err := persistence.Migrate(db); err != nil {
		fmt.Println("could not migrate:", err)

		return 1
	}

	return m.Run()
}

func open() (*gorm.DB, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	// 読み出した時刻を UTC にし、time.Date(..., time.UTC) で書いた期待値と比べられるようにする
	conn := stdlib.OpenDB(*config, stdlib.OptionAfterConnect(func(_ context.Context, c *pgx.Conn) error {
		c.TypeMap().RegisterType(&pgtype.Type{Name: "timestamptz", OID: pgtype.TimestamptzOID, Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC}})

		return nil
	}))

	return gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		NowFunc: func() time.Time { return now },
		Logger:  logger.Discard,
	})
}

// newDB は全てのテーブルを空にした DB を返す。ID は 1 から振り直す。
func newDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := open()
	require.NoError(t, err)
	require.NoError(t, db.Exec("TRUNCATE guilds, playlists, playlist_items, videos, channels, categories, video_snapshots RESTART IDENTITY CASCADE").Error)

	return db
}

// seed はテストの前提となる行を DB に保存する。
func seed(t *testing.T, db *gorm.DB, rows ...any) {
	t.Helper()

	for _, row := range rows {
		require.NoError(t, db.Create(row).Error)
	}
}

// execute はテストの前提となる SQL を実行する。
func execute(t *testing.T, db *gorm.DB, sqls ...string) {
	t.Helper()

	for _, sql := range sqls {
		require.NoError(t, db.Exec(sql).Error)
	}
}

// rows は T のテーブルの全ての行を、論理削除したものも含めて ID の順に返す。
func rows[T any](t *testing.T, db *gorm.DB) []*T {
	t.Helper()

	var found []*T
	require.NoError(t, db.Unscoped().Order("id").Find(&found).Error)

	return found
}

// model は ID が id で、now に作られた行の gorm.Model を返す。
func model(id uint) gorm.Model {
	return gorm.Model{ID: id, CreatedAt: now, UpdatedAt: now}
}

// deleted は ID が id で、now に論理削除された行の gorm.Model を返す。
func deleted(id uint) gorm.Model {
	return gorm.Model{ID: id, CreatedAt: now, UpdatedAt: now, DeletedAt: gorm.DeletedAt{Time: now, Valid: true}}
}

// newSubscription は保存する前の、ギルド guildID の通知設定を返す。新着動画の
// 通知の基準時刻（UpdatedAt）は 6 時間前にする。
func newSubscription(guildID uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	return &subscription.Subscription{
		UpdatedAt:     hoursAgo(6),
		YoutubeID:     playlistID,
		SendChannelID: "ch",
		Title:         title,
		GuildID:       guildID,
	}
}

// storedSubscription は newSubscription で作って ID id で保存した通知設定を返す。
func storedSubscription(id, guildID uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	s := newSubscription(guildID, playlistID, title)
	s.ID = id
	s.CreatedAt = now

	return s
}

// deletedSubscription は保存する前の、論理削除した通知設定を返す。
func deletedSubscription(guildID uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	s := newSubscription(guildID, playlistID, title)
	s.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}

	return s
}

// newItem は保存する前のプレイリストのアイテムを返す。
func newItem(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// storedItem は newItem で作って ID id で保存したアイテムを返す。
func storedItem(id uint, playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	i := newItem(playlistID, itemID, videoID, addedAt)
	i.ID, i.CreatedAt, i.UpdatedAt = id, now, now

	return i
}

// newVideo は保存する前の公開の動画を返す。タイトルは ID と同じにする。
func newVideo(youtubeID library.VideoID, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: youtubeID, Title: string(youtubeID), PrivacyStatus: library.PrivacyPublic, Views: views}
}

// storedVideo は newVideo で作って ID id で保存した動画を返す。
func storedVideo(id uint, youtubeID library.VideoID, views library.ViewCount) *library.Video {
	v := newVideo(youtubeID, views)
	v.ID, v.CreatedAt, v.UpdatedAt = id, now, now

	return v
}

// newSnapshot は保存する前の、hoursAgo(hours) に記録した動画の状態を返す。
func newSnapshot(videoID library.VideoID, hours int, views library.ViewCount) *library.VideoSnapshot {
	return &library.VideoSnapshot{VideoYoutubeID: videoID, RecordedAt: hoursAgo(hours), Views: views}
}

// storedSnapshot は newSnapshot で作って ID id で保存した状態を返す。
func storedSnapshot(id uint, videoID library.VideoID, hours int, views library.ViewCount) *library.VideoSnapshot {
	s := newSnapshot(videoID, hours, views)
	s.ID = id

	return s
}
