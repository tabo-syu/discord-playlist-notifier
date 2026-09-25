// Package mcpserver serves the read-only use cases as an MCP server over
// Streamable HTTP, so that other bots (Claude Code) can look up the
// subscribed playlists.
package mcpserver

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Path the MCP endpoint is served at
const ENDPOINT_PATH = "/mcp"

type server struct {
	http *http.Server
}

func NewServer(addr string, q *application.QueryService, loc *time.Location) *server {
	s := mcp.NewServer(&mcp.Implementation{Name: "playlist-notifier", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: INSTRUCTIONS,
	})
	addTools(s, &tools{q, loc})

	// Stateless: the tools never call back the client, so no session is kept
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
	mux := http.NewServeMux()
	mux.Handle(ENDPOINT_PATH, handler)

	return &server{&http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}}
}

// Serve starts listening and serves in the background.
func (s *server) Serve() error {
	listener, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.http.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("MCP server stopped cause:", err)
		}
	}()
	log.Println("Started MCP server:", s.http.Addr+ENDPOINT_PATH)

	return nil
}

func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}
	log.Println("Stopped MCP server")

	return nil
}
