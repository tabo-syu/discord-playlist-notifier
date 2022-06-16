package main

import (
	"discord-playlist-notifier/command"
	"discord-playlist-notifier/registerer"
	"discord-playlist-notifier/router"
	"discord-playlist-notifier/server"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID = os.Getenv("GUILD_ID")
	Token   = os.Getenv("DISCORD_ACCESS_TOKEN")
)

var d *discordgo.Session

func init() {
	var err error
	d, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	commands := []*command.Command{&command.PlaylistNotifier}

	rt := router.NewRouter(commands)
	sv := server.NewServer(d, rt)
	if err := sv.Serve(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	rr := registerer.NewRegisterer(d, GuildID, commands)
	if err := rr.Register(); err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if err := rr.Unregister(); err != nil {
		log.Fatalf("Cannot unregister commands: %v", err)
	}
	if err := sv.Stop(); err != nil {
		log.Fatalf("Cannot close the session: %v", err)
	}
	log.Println("Gracefully shutting down.")
}
