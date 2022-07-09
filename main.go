package main

import (
	"context"
	"discord-playlist-notifier/command"
	"discord-playlist-notifier/registerer"
	"discord-playlist-notifier/router"
	"discord-playlist-notifier/server"
	"discord-playlist-notifier/service"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Bot parameters
var (
	GUILD_ID      = os.Getenv("GUILD_ID")
	DISCORD_TOKEN = os.Getenv("DISCORD_ACCESS_TOKEN")
	YOUTUBE_TOKEN = os.Getenv("YOUTUBE_APIKEY")
)

var ctx context.Context
var dc *discordgo.Session
var yt *youtube.Service

func init() {
	var err error

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
	service := service.NewYouTubeService(yt)
	server := server.NewServer(dc, router, service)
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
