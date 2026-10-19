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
		if err = container.Terminate(ctx); err != nil {
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

// guildRow は保存する前のギルドの行を返す。
func guildRow(discordID string) *persistence.GuildRecord {
	return &persistence.GuildRecord{DiscordID: discordID}
}

// subscriptionRow は保存する前の、ギルド guildID の通知設定の行を返す。新着動画の
// 通知の基準時刻（notified_until）は 6 時間前、ピックは毎週にする。
func subscriptionRow(guildID uint, playlistID, title string) *persistence.SubscriptionRecord {
	return &persistence.SubscriptionRecord{
		YoutubeID:     playlistID,
		SendChannelID: "ch",
		Title:         title,
		PickInterval:  "weekly",
		GuildID:       guildID,
		NotifiedUntil: hoursAgo(6),
	}
}

// storedSubscriptionRow は subscriptionRow で作って ID id で保存した行を返す。
func storedSubscriptionRow(id, guildID uint, playlistID, title string) *persistence.SubscriptionRecord {
	r := subscriptionRow(guildID, playlistID, title)
	r.Model = model(id)

	return r
}

// deletedSubscriptionRow は保存する前の、論理削除した通知設定の行を返す。
func deletedSubscriptionRow(guildID uint, playlistID, title string) *persistence.SubscriptionRecord {
	r := subscriptionRow(guildID, playlistID, title)
	r.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}

	return r
}

// storedSubscription は storedSubscriptionRow の行をリポジトリが読んだときの通知設定を返す。
func storedSubscription(id, guildID uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	return &subscription.Subscription{
		ID:            id,
		YoutubeID:     playlistID,
		SendChannelID: "ch",
		Title:         title,
		PickInterval:  subscription.PickWeekly,
		GuildID:       guildID,
		NotifiedUntil: hoursAgo(6),
	}
}

// playlist は items を持つプレイリストを返す。
func playlist(playlistID library.PlaylistID, items ...*library.PlaylistItem) *library.Playlist {
	return &library.Playlist{YoutubeID: playlistID, Items: items}
}

// item はプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// itemRow は保存する前の、item と同じアイテムの行を返す。
func itemRow(playlistID, itemID, videoID string, addedAt time.Time) *persistence.PlaylistItemRecord {
	return &persistence.PlaylistItemRecord{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// storedItemRow は行 r を ID id で now に保存した行を返す。
func storedItemRow(id uint, r *persistence.PlaylistItemRecord) *persistence.PlaylistItemRecord {
	r.ID, r.CreatedAt, r.UpdatedAt = id, now, now

	return r
}

// video は公開の動画を返す。タイトルは ID と同じにする。
func video(youtubeID library.VideoID, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: youtubeID, DetailsFetched: true, Title: string(youtubeID), PrivacyStatus: library.PrivacyPublic, Views: views}
}

// videoRow は保存する前の、video と同じ動画の行を返す。
func videoRow(youtubeID string, views uint64) *persistence.VideoRecord {
	return &persistence.VideoRecord{YoutubeID: youtubeID, DetailsFetched: true, Title: youtubeID, PrivacyStatus: "public", Views: views}
}

// storedVideoRow は行 r を ID id で now に保存した行を返す。
func storedVideoRow(id uint, r *persistence.VideoRecord) *persistence.VideoRecord {
	r.ID, r.CreatedAt, r.UpdatedAt = id, now, now

	return r
}

// fullVideo は全てのフィールドに値のある動画を返す。変換で写し忘れたフィールドを見つけるのに使う。
func fullVideo() *library.Video {
	milestone := library.ViewCount(100_000)

	return &library.Video{
		YoutubeID:        "full",
		DetailsFetched:   true,
		Title:            "全部入り",
		Views:            120_000,
		PrivacyStatus:    library.PrivacyUnlisted,
		PublishedAt:      hoursAgo(24),
		ViewMilestone:    &milestone,
		Likes:            300,
		Comments:         40,
		ChannelYoutubeID: "UC1",
		Duration:         3 * time.Minute,
		Tags:             []string{"tag"},
		CategoryID:       "10",
	}
}

// fullVideoRow は保存する前の、fullVideo と同じ動画の行を返す。
func fullVideoRow() *persistence.VideoRecord {
	milestone := uint64(100_000)

	return &persistence.VideoRecord{
		YoutubeID:        "full",
		DetailsFetched:   true,
		Title:            "全部入り",
		Views:            120_000,
		PrivacyStatus:    "unlisted",
		PublishedAt:      hoursAgo(24),
		ViewMilestone:    &milestone,
		Likes:            300,
		Comments:         40,
		ChannelYoutubeID: "UC1",
		Duration:         3 * time.Minute,
		Tags:             []string{"tag"},
		CategoryID:       "10",
	}
}

// snapshot は hoursAgo(hours) に記録した、公開の動画の状態を返す。タイトルは動画の ID と同じにする。
func snapshot(videoID library.VideoID, hours int, views library.ViewCount) *library.VideoSnapshot {
	return &library.VideoSnapshot{VideoYoutubeID: videoID, RecordedAt: hoursAgo(hours), Title: string(videoID), PrivacyStatus: library.PrivacyPublic, Views: views, Likes: 2, Comments: 1}
}

// snapshotRow は保存する前の、snapshot と同じ状態の行を返す。
func snapshotRow(videoID string, hours int, views uint64) *persistence.VideoSnapshotRecord {
	return &persistence.VideoSnapshotRecord{VideoYoutubeID: videoID, RecordedAt: hoursAgo(hours), Title: videoID, PrivacyStatus: "public", Views: views, Likes: 2, Comments: 1}
}
