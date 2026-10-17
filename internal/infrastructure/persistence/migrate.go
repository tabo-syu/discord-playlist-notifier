package persistence

import (
	"fmt"
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

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

	if err := db.AutoMigrate(&guild.Guild{}, &subscription.Subscription{}, &library.PlaylistItem{}, &library.Video{}, &library.Channel{}, &library.Category{}, &library.VideoSnapshot{}); err != nil {
		return err
	}

	// 集約は互いを ID でのみ参照するので、GORM は Guild.Playlists の関連付けから
	// 作っていた外部キーをもう導出しない。
	if !db.Migrator().HasConstraint(&subscription.Subscription{}, "fk_guilds_playlists") {
		err := db.Exec("ALTER TABLE playlists ADD CONSTRAINT fk_guilds_playlists FOREIGN KEY (guild_id) REFERENCES guilds(id)").Error
		if err != nil {
			return fmt.Errorf("could not add the foreign key of playlists: %w", err)
		}
	}

	return nil
}
