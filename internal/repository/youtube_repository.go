package repository

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

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

func NewYouTubeRepository(yt *youtube.Service) *youTubeRepository {
	return &youTubeRepository{yt}
}

func (r *youTubeRepository) FindPlaylists(ids ...string) ([]*domain.Playlist, error) {
	lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).MaxResults(MAX_RESULTS).
		Id(ids...).Do()
	if err != nil {
		return nil, err
	}
	if len(lists.Items) == 0 {
		return nil, domain.ErrYouTubePlaylistCouldNotFound
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
		return nil, domain.ErrYouTubePlaylistCouldNotFound
	}

	var response = []*domain.Playlist{}
	for _, playlist := range lists.Items {
		playlistItems, err := r.youtube.PlaylistItems.List([]string{"snippet"}).MaxResults(MAX_RESULTS).PlaylistId(playlist.Id).Do()
		if err != nil {
			return nil, err
		}
		// 動画の登録されていないプレイリストはスキップ
		if len(playlistItems.Items) == 0 {
			continue
		}

		var vids []string
		for _, item := range playlistItems.Items {
			vids = append(vids, item.Snippet.ResourceId.VideoId)
		}
		videos, err := r.youtube.Videos.List([]string{"id", "snippet", "statistics"}).MaxResults(MAX_RESULTS).Id(vids...).Do()
		if err != nil {
			return nil, err
		}
		// プレイリストに登録されているがすべての動画が削除されている場合はスキップ
		if len(videos.Items) == 0 {
			continue
		}

		var cids []string
		for _, item := range videos.Items {
			cids = append(cids, item.Snippet.ChannelId)
		}
		channels, err := r.youtube.Channels.List([]string{"id", "snippet"}).MaxResults(MAX_RESULTS).Id(cids...).Do()
		if err != nil {
			return nil, err
		}
		// プレイリストに登録されているがすべてのチャンネルが削除されている場合はスキップ
		if len(channels.Items) == 0 {
			continue
		}

		var listVideos = []domain.Video{}
		for _, listVideo := range playlistItems.Items {
			var video *youtube.Video
			for _, v := range videos.Items {
				if v.Id == listVideo.Snippet.ResourceId.VideoId {
					video = v

					break
				}
			}
			// 動画が削除されている場合は次の動画へ
			if video == nil {
				continue
			}

			var channel *youtube.Channel
			for _, c := range channels.Items {
				if c.Id == video.Snippet.ChannelId {
					channel = c

					break
				}
			}

			publishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, listVideo.Snippet.PublishedAt)
			ownerPublishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, video.Snippet.PublishedAt)
			listVideos = append(listVideos, domain.Video{
				YoutubeID:        listVideo.Snippet.ResourceId.VideoId,
				Title:            listVideo.Snippet.Title,
				Views:            video.Statistics.ViewCount,
				Thumbnail:        video.Snippet.Thumbnails.High.Url,
				ChannelName:      channel.Snippet.Title,
				ChannelIcon:      channel.Snippet.Thumbnails.Default.Url,
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
