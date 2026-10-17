// Package mcpserver は読み取り専用のユースケースを Streamable HTTP の
// MCP サーバーとして提供し、他のボット（Claude Code）が通知登録された
// プレイリストを参照できるようにする。
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

// MCP のエンドポイントを提供するパス
const ENDPOINT_PATH = "/mcp"

type server struct {
	http *http.Server
}

func NewServer(addr string, q *application.QueryService, loc *time.Location) *server {
	s := mcp.NewServer(&mcp.Implementation{Name: "playlist-notifier", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: INSTRUCTIONS,
	})
	addTools(s, &tools{q, loc})

	// ステートレス: ツールがクライアントを呼び返すことはないので、セッションは保持しない
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
	mux := http.NewServeMux()
	mux.Handle(ENDPOINT_PATH, handler)

	return &server{&http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}}
}

// Serve は待ち受けを始め、バックグラウンドで提供する。
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
