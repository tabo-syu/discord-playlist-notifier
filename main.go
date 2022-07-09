package main

import (
	"context"
	"discord-playlist-notifier/command"
	"discord-playlist-notifier/registerer"
	"discord-playlist-notifier/repository"
	"discord-playlist-notifier/router"
	"discord-playlist-notifier/server"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB_HOST     = os.Getenv("DB_HOST")
	DB_PORT     = os.Getenv("DB_PORT")
	DB_NAME     = os.Getenv("DB_NAME")
	DB_USER     = os.Getenv("DB_USER")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	DB_TIMEZONE = os.Getenv("DB_TIMEZONE")

	GUILD_ID      = os.Getenv("GUILD_ID")
	DISCORD_TOKEN = os.Getenv("DISCORD_ACCESS_TOKEN")
	YOUTUBE_TOKEN = os.Getenv("YOUTUBE_APIKEY")
)

var (
	ctx context.Context
	db  *gorm.DB
	dc  *discordgo.Session
	yt  *youtube.Service
)

func init() {
	var err error

	dsn := strings.Join([]string{
		"host=" + DB_HOST,
		"user=" + DB_USER,
		"password=" + DB_PASSWORD,
		"dbname=" + DB_NAME,
		"port=" + DB_PORT,
		"sslmode=disable",
		"TimeZone=" + DB_TIMEZONE,
	}, " ")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect the db: %v", err)
	}

	dc, err = discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid discord token: %v", err)
	}

	ctx = context.Background()
	yt, err = youtube.NewService(ctx, option.WithAPIKey(YOUTUBE_TOKEN))
	if err != nil {
		log.Fatalf("Invalid youtube token: %v", err)
	}
}

func main() {
	commands := []*command.Command{&command.PlaylistNotifier}

	router := router.NewRouter(commands)
	repository := repository.NewYouTubeRepository(yt)
	server := server.NewServer(dc, router, repository)
	if err := server.Serve(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	registerer := registerer.NewRegisterer(dc, GUILD_ID, commands)
	if err := registerer.Register(); err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	defer registerer.Unregister()
	defer server.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
