package repository

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"
	"time"

	"google.golang.org/api/youtube/v3"
)

const (
	YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"
	MAX_RESULTS        = 20
)

type YouTubeRepository interface {
	FindPlaylists(...string) ([]*domain.Playlist, error)
}

type youTubeRepository struct {
	youtube *youtube.Service
}

func NewYouTubeRepository(yt *youtube.Service) YouTubeRepository {
	return &youTubeRepository{yt}
}

func (r *youTubeRepository) FindPlaylists(ids ...string) ([]*domain.Playlist, error) {
	// TODO: if len(ids) > MAX_RESULTS {} の時のロギング
	lists, err := r.youtube.Playlists.List([]string{"id"}).MaxResults(MAX_RESULTS).
		Id(ids...).Do()
	if err != nil {
		return nil, err
	}
	if len(lists.Items) == 0 {
		return nil, errs.ErrYouTubePlaylistCouldNotFound
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
			publishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, video.Snippet.PublishedAt)
			videos = append(videos, domain.Video{
				YoutubeID:   video.Snippet.ResourceId.VideoId,
				Title:       video.Snippet.Title,
				PublishedAt: publishedAt,
			})
		}

		response = append(response, &domain.Playlist{
			YoutubeID: playlist.Id,
			Videos:    videos,
		})
	}

	return response, nil
}
