package persistence

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrate はスキーマを最新にする。AutoMigrate はテーブルとカラムの追加しか
// しないので、リネームや削除は手作業で行う必要がある。
func Migrate(db *gorm.DB) error {
	// videos テーブルは以前、通知した動画の履歴だった。今は各動画の最新の状態を
	// 持つので、古いテーブルは一度だけ削除する。
	if db.Migrator().HasColumn("videos", "playlist_id") {
		if err := db.Migrator().DropTable("videos"); err != nil {
			return fmt.Errorf("could not drop the old videos table: %w", err)
		}

		log.Println("Dropped the old videos table")
	}

	if err := db.AutoMigrate(&guildRecord{}, &subscriptionRecord{}, &playlistItemRecord{}, &videoRecord{}, &channelRecord{}, &categoryRecord{}, &videoSnapshotRecord{}); err != nil {
		return err
	}

	// 新着動画の通知の基準時刻は、以前は updated_at だった。AutoMigrate が足した
	// notified_until を空のままにすると全ての動画を新着とみなすので、updated_at で埋める。
	// 途中で落ちても次の実行で埋まるように、カラムの有無にかかわらず毎回実行する。
	result := db.Exec("UPDATE playlists SET notified_until = updated_at WHERE notified_until IS NULL")
	if err := result.Error; err != nil {
		return fmt.Errorf("could not fill notified_until of playlists: %w", err)
	}

	if result.RowsAffected > 0 {
		log.Println("Filled notified_until of playlists:", result.RowsAffected)
	}

	// 詳細を取得済みかは、以前はタイトルが空でないかで判断していた。AutoMigrate が足した
	// details_fetched は既存の行で NULL なので、その判断で埋める。notified_until と同じく
	// 毎回実行する。
	result = db.Exec("UPDATE videos SET details_fetched = (COALESCE(title, '') <> '') WHERE details_fetched IS NULL")
	if err := result.Error; err != nil {
		return fmt.Errorf("could not fill details_fetched of videos: %w", err)
	}

	if result.RowsAffected > 0 {
		log.Println("Filled details_fetched of videos:", result.RowsAffected)
	}

	// 集約は互いを ID でのみ参照するので、GORM は Guild.Playlists の関連付けから
	// 作っていた外部キーをもう導出しない。
	if !db.Migrator().HasConstraint(&subscriptionRecord{}, "fk_guilds_playlists") {
		err := db.Exec("ALTER TABLE playlists ADD CONSTRAINT fk_guilds_playlists FOREIGN KEY (guild_id) REFERENCES guilds(id)").Error
		if err != nil {
			return fmt.Errorf("could not add the foreign key of playlists: %w", err)
		}
	}

	return nil
}
