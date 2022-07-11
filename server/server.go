package server

import (
	"discord-playlist-notifier/repository"
	"discord-playlist-notifier/router"

	"github.com/bwmarrin/discordgo"
)

type Server interface {
	Serve() error
	Stop() error
}

type server struct {
	session *discordgo.Session
	db      repository.DBRepository
	youtube repository.YouTubeRepository
	router  router.Router
}

func NewServer(session *discordgo.Session, db repository.DBRepository, youtube repository.YouTubeRepository, router router.Router) *server {
	return &server{session, db, youtube, router}
}

func (s *server) Serve() error {
	s.session.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		request := i.ApplicationCommandData()
		command := s.router.Route(request.Name)
		if command == nil {
			return
		}

		response := command(&request, s.db, s.youtube)

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
