package router

import (
	"discord-playlist-notifier/handler/command"
)

type Router interface {
	Route(string) command.HandleType
}

type router struct {
	routes map[string]command.HandleType
}

func NewRouter(commands []command.Command) Router {
	routes := map[string]command.HandleType{}
	for _, command := range commands {
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
