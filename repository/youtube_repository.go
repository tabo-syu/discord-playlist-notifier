package repository

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"

	"google.golang.org/api/youtube/v3"
)

const MAX_RESULTS = 20

type YouTubeRepository interface {
	GetPlaylists(...string) ([]*domain.Playlist, error)
}

type youTubeRepository struct {
	youtube *youtube.Service
}

func NewYouTubeRepository(yt *youtube.Service) *youTubeRepository {
	return &youTubeRepository{yt}
}

func (r *youTubeRepository) GetPlaylists(ids ...string) ([]*domain.Playlist, error) {
	// TODO: if len(ids) > MAX_RESULTS {} の時のロギング
	lists, err := r.youtube.Playlists.List([]string{"id"}).MaxResults(MAX_RESULTS).
		Id(ids...).Do()
	if err != nil {
		return nil, err
	}
	if len(lists.Items) == 0 {
		return nil, errs.ErrPlaylistCouldNotFound
	}

	var response = []*domain.Playlist{}
	for _, playlist := range lists.Items {
		item, err := r.youtube.PlaylistItems.List([]string{"snippet"}).MaxResults(MAX_RESULTS).
			PlaylistId(playlist.Id).Do()
		if err != nil {
			return nil, err
		}

		var videos = []domain.Video{}
		for _, video := range item.Items {
			videos = append(videos, domain.Video{
				YoutubeID: video.Snippet.ResourceId.VideoId,
				Title:     video.Snippet.Title,
			})
		}

		response = append(response, &domain.Playlist{
			YoutubeID: playlist.Id,
			Mention:   false,
			Videos:    videos,
		})
	}

	return response, nil
}
