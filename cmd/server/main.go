package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/env"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/youtube"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command/playlist_notifier"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/mcpserver"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/scheduler"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"google.golang.org/api/option"
	ytapi "google.golang.org/api/youtube/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sr *gocron.Scheduler
	db *gorm.DB
	dc *discordgo.Session
	yt *ytapi.Service
)

func init() {
	var err error

	location, err := time.LoadLocation(env.DB_TIMEZONE)
	if err != nil {
		log.Fatalf("Could not load time location: %v", err)
	}
	sr = gocron.NewScheduler(location)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=%s",
		env.DB_HOST,
		env.DB_USER,
		env.DB_PASSWORD,
		env.DB_NAME,
		env.DB_PORT,
		env.DB_TIMEZONE,
	)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect the db: %v", err)
	}
	if err := persistence.Migrate(db); err != nil {
		log.Fatalf("Could not migrate tables: %v", err)
	}

	dc, err = discordgo.New("Bot " + env.DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid discord token: %v", err)
	}

	yt, err = ytapi.NewService(context.Background(), option.WithAPIKey(env.YOUTUBE_TOKEN))
	if err != nil {
		log.Fatalf("Invalid youtube token: %v", err)
	}
}

func main() {
	location := sr.Location()

	yc := youtube.NewClient(yt)
	gr := persistence.NewGuildRepository(db)
	sbr := persistence.NewSubscriptionRepository(db)
	lr := persistence.NewLibraryRepository(db)

	ss := application.NewSubscriptionService(yc, sbr, gr)
	gs := application.NewGuildService(gr, sbr)
	ls := application.NewLibraryService(yc, lr, location)
	ns := application.NewNotificationService(sbr, ls, notifier.New(dc, location), location)

	commands := []command.Command{playlist_notifier.NewPlaylistNotifier(ss, ls, location)}
	server := discord.NewServer(
		dc,
		discord.NewRegisterer(dc, commands),
		discord.NewEvent(gs),
		discord.NewRouter(commands),
	)
	if err := server.Serve(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer server.Stop()

	scheduler := scheduler.NewScheduler(sr, ns)
	scheduler.Start()
	defer scheduler.Stop()

	// The MCP server only serves lookups, so the bot keeps running without it
	if env.MCP_ADDR != "" {
		mcp := mcpserver.NewServer(env.MCP_ADDR, application.NewQueryService(sbr, ls), location)
		if err := mcp.Serve(); err != nil {
			log.Println("Could not start the MCP server cause:", err)
		} else {
			defer mcp.Stop()
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
