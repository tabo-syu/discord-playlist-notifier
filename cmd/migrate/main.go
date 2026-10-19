package main

import (
	"fmt"
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/env"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// main はスキーマを更新して終了する。ボット（cmd/bot）と MCP サーバー（cmd/mcp）は
// スキーマを変更しないので、これを先に実行しておく。
func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=%s",
		env.DB_HOST,
		env.DB_USER,
		env.DB_PASSWORD,
		env.DB_NAME,
		env.DB_PORT,
		env.DB_TIMEZONE,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect the db: %v", err)
	}

	if err = persistence.Migrate(db); err != nil {
		log.Fatalf("Could not migrate tables: %v", err)
	}

	log.Println("Migrated tables")
}
