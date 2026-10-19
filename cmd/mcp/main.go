package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/env"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/mcpserver"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// main は読み取り専用の MCP サーバーだけを起動する。スキーマは cmd/migrate が
// 更新するので、ここでは変更しない。
func main() {
	location, err := time.LoadLocation(env.DB_TIMEZONE)
	if err != nil {
		log.Fatalf("Could not load time location: %v", err)
	}

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

	sbr := persistence.NewSubscriptionRepository(db)
	pr := persistence.NewPlaylistRepository(db)
	vr := persistence.NewVideoRepository(db)
	cr := persistence.NewChannelRepository(db)
	car := persistence.NewCategoryRepository(db)
	vsr := persistence.NewVideoSnapshotRepository(db)

	// QueryService は保存済みの中身しか読まず YouTube を呼ばないので、YouTube のクライアントは作らない
	qs := application.NewQueryService(sbr, library.NewCatalog(pr, vr), vsr, car, cr, location, time.Now)

	server := mcpserver.NewServer(env.MCP_ADDR, qs, location, time.Now)
	if err = server.Serve(); err != nil {
		log.Fatalf("Could not start the MCP server: %v", err)
	}

	defer func() {
		if err := server.Stop(); err != nil {
			log.Println("Could not stop the MCP server cause:", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Gracefully shutting down.")
}
