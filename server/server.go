package server

import (
	"discord-playlist-notifier/router"

	"github.com/bwmarrin/discordgo"
)

type Server interface {
	Serve() error
	Stop() error
}

type server struct {
	session *discordgo.Session
	router  router.Router
}

func NewServer(session *discordgo.Session, router router.Router) *server {
	return &server{session, router}
}

func (s *server) Serve() error {
	s.session.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		request := i.ApplicationCommandData()
		command := s.router.Route(request.Name)
		if command == nil {
			return
		}

		response := command(&request)

		d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	})

	return s.session.Open()
}

func (s *server) Stop() error {
	return s.session.Close()
}
