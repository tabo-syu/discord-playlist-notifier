package persistence

import (
	"fmt"
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"gorm.io/gorm"
)

// Migrate brings the schema up to date. AutoMigrate only adds tables and
// columns; renaming or dropping them has to be done by hand.
func Migrate(db *gorm.DB) error {
	// The videos table used to be a history of notified videos. It now holds
	// the latest state of each video, so the old table is dropped once.
	if db.Migrator().HasColumn("videos", "playlist_id") {
		if err := db.Migrator().DropTable("videos"); err != nil {
			return fmt.Errorf("could not drop the old videos table: %w", err)
		}
		log.Println("Dropped the old videos table")
	}

	if err := db.AutoMigrate(&guild.Guild{}, &subscription.Subscription{}, &library.PlaylistItem{}, &library.Video{}); err != nil {
		return err
	}

	// Aggregates reference each other by ID only, so GORM no longer derives the
	// foreign key it used to create from the Guild.Playlists association.
	if !db.Migrator().HasConstraint(&subscription.Subscription{}, "fk_guilds_playlists") {
		err := db.Exec("ALTER TABLE playlists ADD CONSTRAINT fk_guilds_playlists FOREIGN KEY (guild_id) REFERENCES guilds(id)").Error
		if err != nil {
			return fmt.Errorf("could not add the foreign key of playlists: %w", err)
		}
	}

	return nil
}
