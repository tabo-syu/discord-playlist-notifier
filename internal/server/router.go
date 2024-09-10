package server

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/server/command"
)

type Router interface {
	Route(string) command.HandleType
}

type router struct {
	routes map[string]command.HandleType
}

func NewRouter(cs []command.Command) Router {
	routes := map[string]command.HandleType{}
	for _, command := range cs {
		routes[command.GetCommand().Name] = command.Handle
	}

	return &router{routes}
}

func (r *router) Route(commandName string) command.HandleType {
	if h, ok := r.routes[commandName]; ok {
		return h
	}

	return nil
}