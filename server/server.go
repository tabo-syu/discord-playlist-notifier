package server

import (
	"discord-playlist-notifier/handler/event"
	"discord-playlist-notifier/router"
	"discord-playlist-notifier/service"

	"github.com/bwmarrin/discordgo"
)

type Server interface {
	Serve() error
	Stop() error
}

type server struct {
	session  *discordgo.Session
	playlist service.PlaylistService
	router   router.Router
}

func NewServer(session *discordgo.Session, playlist service.PlaylistService, router router.Router) Server {
	return &server{session, playlist, router}
}

func (s *server) Serve() error {
	// Bot がサーバーに参加したとき
	s.session.AddHandler(func(d *discordgo.Session, g *discordgo.GuildCreate) {
		event.GuildCreate(g.Guild.ID)
	})

	// Bot がサーバーから削除されたとき
	s.session.AddHandler(func(d *discordgo.Session, g *discordgo.GuildDelete) {
		event.GuildDelete(g.Guild.ID)
	})

	// コマンドを受け付けたとき
	s.session.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		request := i.ApplicationCommandData()
		command := s.router.Route(request.Name)
		if command == nil {
			return
		}

		response := command(i.GuildID, &request, s.playlist)

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
