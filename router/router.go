package router

import (
	"discord-playlist-notifier/handler/command"
)

type Router interface {
	Route(string) command.Handle
}

type router struct {
	routes map[string]command.Handle
}

func NewRouter(commands []*command.Command) Router {
	routes := map[string]command.Handle{}
	for _, command := range commands {
		routes[command.Info.Name] = command.Handle
	}

	return &router{routes}
}

func (r *router) Route(commandName string) command.Handle {
	if h, ok := r.routes[commandName]; ok {
		return h
	}

	return nil
}
