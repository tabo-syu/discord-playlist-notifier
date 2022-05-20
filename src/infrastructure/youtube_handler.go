package infrastructure

import (
	"context"
	"discord-playlist-notifier/src/interfaces"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubeHandler struct {
	YouTube *youtube.Service
	ctx     context.Context
}

func NewYouTubeHandler(ctx context.Context) (interfaces.YouTubeHandler, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_APIKEY")))
	if err != nil {
		return nil, err
	}

	return &YouTubeHandler{service, ctx}, nil
}

func (s *YouTubeHandler) PlaylistItems() interfaces.PlaylistItemsService {
	return s.YouTube.PlaylistItems
}

func (s *YouTubeHandler) Playlists() interfaces.PlaylistsService {
	return s.YouTube.Playlists
}
