package service

import (
	"google.golang.org/api/youtube/v3"
)

const MAX_RESULTS = 20

type YouTubeService interface {
	GetPlaylist(string) (string, error)
}

type youTubeService struct {
	youtube *youtube.Service
}

func NewYouTubeService(yt *youtube.Service) *youTubeService {
	return &youTubeService{yt}
}

func (s *youTubeService) GetPlaylist(id string) (string, error) {
	call := s.youtube.PlaylistItems.List([]string{"snippet"}).MaxResults(MAX_RESULTS)
	res, err := call.PlaylistId(id).Do()
	if err != nil {
		return "", err
	}

	return res.Items[0].Snippet.Title, nil
}
