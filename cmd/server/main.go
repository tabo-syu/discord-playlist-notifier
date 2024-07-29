package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/env"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
	"github.com/tabo-syu/discord-playlist-notifier/internal/scheduler"
	"github.com/tabo-syu/discord-playlist-notifier/internal/server"
	"github.com/tabo-syu/discord-playlist-notifier/internal/server/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/server/command/playlist_notifier"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sr *gocron.Scheduler
	db *gorm.DB
	dc *discordgo.Session
	yt *youtube.Service
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
	err = db.AutoMigrate(&domain.Guild{}, &domain.Playlist{}, &domain.Video{})
	if err != nil {
		log.Fatalf("Could not migrate tables: %v", err)
	}

	dc, err = discordgo.New("Bot " + env.DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid discord token: %v", err)
	}

	yt, err = youtube.NewService(context.Background(), option.WithAPIKey(env.YOUTUBE_TOKEN))
	if err != nil {
		log.Fatalf("Invalid youtube token: %v", err)
	}
}

func main() {
	yr := repository.NewYouTubeRepository(yt)
	gr := repository.NewGuildRepository(db)
	pr := repository.NewPlaylistRepository(db)

	ps := service.NewPlaylistService(yr, pr, gr)
	gs := service.NewGuildService(gr, pr)
	rr := scheduler.NewRenderer(dc)

	commands := []command.Command{playlist_notifier.NewPlaylistNotifier(ps)}
	server := server.NewServer(
		dc,
		server.NewRegisterer(dc, commands),
		server.NewEvent(gs),
		server.NewRouter(commands),
	)
	if err := server.Serve(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer server.Stop()

	scheduler := scheduler.NewScheduler(sr, scheduler.NewSchedule(ps, rr))
	scheduler.Start()
	defer scheduler.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
