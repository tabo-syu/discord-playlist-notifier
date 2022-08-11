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
	FindPlaylistsWithVideos(...string) ([]*domain.Playlist, error)
}

type youTubeRepository struct {
	youtube *youtube.Service
}

func NewYouTubeRepository(yt *youtube.Service) YouTubeRepository {
	return &youTubeRepository{yt}
}

func (r *youTubeRepository) FindPlaylists(ids ...string) ([]*domain.Playlist, error) {
	lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).MaxResults(MAX_RESULTS).
		Id(ids...).Do()
	if err != nil {
		return nil, err
	}
	if len(lists.Items) == 0 {
		return nil, errs.ErrYouTubePlaylistCouldNotFound
	}

	var response = []*domain.Playlist{}
	for _, playlist := range lists.Items {
		response = append(response, &domain.Playlist{
			YoutubeID: playlist.Id,
			Title:     playlist.Snippet.Title,
		})
	}

	return response, nil
}

func (r *youTubeRepository) FindPlaylistsWithVideos(ids ...string) ([]*domain.Playlist, error) {
	// TODO: if len(ids) > MAX_RESULTS {} の時のロギング
	lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).MaxResults(MAX_RESULTS).
		Id(ids...).Do()
	if err != nil {
		return nil, err
	}
	if len(lists.Items) == 0 {
		return nil, errs.ErrYouTubePlaylistCouldNotFound
	}

	var response = []*domain.Playlist{}
	for _, playlist := range lists.Items {
		listChan, err := r.youtube.PlaylistItems.List([]string{"snippet"}).MaxResults(MAX_RESULTS).
			PlaylistId(playlist.Id).Do()
		if err != nil {
			return nil, err
		}

		var listVideos = []domain.Video{}
		for _, listVideo := range listChan.Items {
			video, err := r.youtube.Videos.List([]string{"snippet", "statistics"}).MaxResults(MAX_RESULTS).
				Id(listVideo.Snippet.ResourceId.VideoId).Do()
			if err != nil {
				return nil, err
			}
			channels, err := r.youtube.Channels.List([]string{"snippet"}).MaxResults(MAX_RESULTS).
				Id(video.Items[0].Snippet.ChannelId).Do()
			if err != nil {
				return nil, err
			}

			publishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, listVideo.Snippet.PublishedAt)
			ownerPublishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, video.Items[0].Snippet.PublishedAt)
			listVideos = append(listVideos, domain.Video{
				YoutubeID:        listVideo.Snippet.ResourceId.VideoId,
				Title:            listVideo.Snippet.Title,
				Views:            video.Items[0].Statistics.ViewCount,
				Thumbnail:        listVideo.Snippet.Thumbnails.High.Url,
				ChannelName:      channels.Items[0].Snippet.Title,
				ChannelIcon:      channels.Items[0].Snippet.Thumbnails.Default.Url,
				PublishedAt:      publishedAt,
				OwnerPublishedAt: ownerPublishedAt,
			})
		}

		response = append(response, &domain.Playlist{
			YoutubeID: playlist.Id,
			Title:     playlist.Snippet.Title,
			Videos:    listVideos,
		})
	}

	return response, nil
}
