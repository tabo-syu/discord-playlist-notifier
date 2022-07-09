package repository

import (
	"google.golang.org/api/youtube/v3"
)

const MAX_RESULTS = 20

type YouTubeRepository interface {
	GetPlaylist(string) (string, error)
}

type youTubeRepository struct {
	youtube *youtube.Service
}

func NewYouTubeRepository(yt *youtube.Service) *youTubeRepository {
	return &youTubeRepository{yt}
}

func (s *youTubeRepository) GetPlaylist(id string) (string, error) {
	call := s.youtube.PlaylistItems.List([]string{"snippet"}).MaxResults(MAX_RESULTS)
	res, err := call.PlaylistId(id).Do()
	if err != nil {
		return "", err
	}

	return res.Items[0].Snippet.Title, nil
}
