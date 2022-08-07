package server

import (
	"discord-playlist-notifier/handler/event"
	"discord-playlist-notifier/registerer"
	"discord-playlist-notifier/router"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Server interface {
	Serve() error
	Stop() error
}

type server struct {
	session    *discordgo.Session
	registerer registerer.Registerer
	event      event.Event
	router     router.Router
}

func NewServer(s *discordgo.Session, rg registerer.Registerer, e event.Event, rt router.Router) Server {
	return &server{s, rg, e, rt}
}

func (s *server) Serve() error {
	// Bot がサーバーに参加したとき
	s.session.AddHandler(func(d *discordgo.Session, g *discordgo.GuildCreate) {
		if err := s.registerer.Register(g.Guild.ID); err != nil {
			log.Println("Commands could not register:", g.Guild.ID, "cause:", err)
		} else {
			log.Println("Commands registered:", g.Guild.ID)
		}

		if err := s.event.GuildCreate(g.Guild.ID); err != nil {
			log.Println("Guild record could not create:", g.Guild.ID, "cause:", err)
		} else {
			log.Println("Guild record created:", g.Guild.ID)
		}
	})

	// Bot がサーバーから削除されたとき
	s.session.AddHandler(func(d *discordgo.Session, g *discordgo.GuildDelete) {
		if err := s.event.GuildDelete(g.Guild.ID); err != nil {
			log.Println("Guild record could not delete:", g.Guild.ID, "cause:", err)
		} else {
			log.Println("Guild record deleted:", g.Guild.ID)
		}
	})

	// コマンドを受け付けたとき
	s.session.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		request := i.ApplicationCommandData()
		command := s.router.Route(request.Name)
		if command == nil {
			return
		}

		response := command(&request, i.GuildID, i.ChannelID)

		d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	})

	if err := s.session.Open(); err != nil {
		return err
	}
	log.Println("Started server")

	return nil
}

func (s *server) Stop() error {
	s.registerer.Unregister()
	log.Println("Commands unregistered")

	if err := s.session.Close(); err != nil {
		return err
	}
	log.Println("Stopped server")

	return nil
}
