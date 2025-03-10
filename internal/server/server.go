package server

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Server struct {
	session    *discordgo.Session
	registerer *registrar
	event      *event
	router     *router
}

func NewServer(s *discordgo.Session, rg *registrar, e *event, rt *router) *Server {
	return &Server{s, rg, e, rt}
}

func (s *Server) Serve() error {
	// When the bot joins a server
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

	// When the bot is removed from a server
	s.session.AddHandler(func(d *discordgo.Session, g *discordgo.GuildDelete) {
		if err := s.event.GuildDelete(g.Guild.ID); err != nil {
			log.Println("Guild record could not delete:", g.Guild.ID, "cause:", err)
		} else {
			log.Println("Guild record deleted:", g.Guild.ID)
		}
	})

	// When a command is received
	s.session.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		request := i.ApplicationCommandData()
		command := s.router.Route(request.Name)
		if command == nil {
			log.Println("Unknown command received:", request.Name)
			// Respond with an error message
			err := d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command. Please use one of the available commands.",
				},
			})
			if err != nil {
				log.Println("Error responding to interaction:", err)
			}
			return
		}

		response := command(&request, i.GuildID, i.ChannelID)

		err := d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
		if err != nil {
			log.Println("Error responding to interaction:", err, "for command:", request.Name, "in guild:", i.GuildID)
		}
	})

	if err := s.session.Open(); err != nil {
		return err
	}
	log.Println("Started server")

	return nil
}

func (s *Server) Stop() error {
	s.registerer.Unregister()
	log.Println("Commands unregistered")

	if err := s.session.Close(); err != nil {
		return err
	}
	log.Println("Stopped server")

	return nil
}
