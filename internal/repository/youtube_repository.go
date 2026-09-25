package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"google.golang.org/api/youtube/v3"
)

const (
	YOUTUBE_TIMEFORMAT   = "2006-01-02T15:04:05Z"
	MAX_RESULTS_PER_PAGE = 50 // Maximum allowed by YouTube API
	MAX_BATCH_SIZE       = 50 // Maximum number of IDs per API call
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
	var response = []*domain.Playlist{}

	// Process playlists in batches to respect YouTube API limits
	for i := 0; i < len(ids); i += MAX_BATCH_SIZE {
		end := i + MAX_BATCH_SIZE
		if end > len(ids) {
			end = len(ids)
		}

		batchIds := ids[i:end]
		lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).
			MaxResults(int64(len(batchIds))).
			Id(batchIds...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		if len(lists.Items) == 0 && len(ids) == 1 {
			return nil, domain.ErrYouTubePlaylistNotFound
		}

		for _, playlist := range lists.Items {
			response = append(response, &domain.Playlist{
				YoutubeID: playlist.Id,
				Title:     playlist.Snippet.Title,
			})
		}
	}

	if len(response) == 0 {
		return nil, domain.ErrYouTubePlaylistNotFound
	}

	return response, nil
}

func (r *youTubeRepository) FindPlaylistsWithVideos(ids ...string) ([]*domain.Playlist, error) {
	// Log if there are too many IDs
	if len(ids) > MAX_BATCH_SIZE {
		log.Printf("Warning: %d playlist IDs provided, processing in batches of %d", len(ids), MAX_BATCH_SIZE)
	}

	var response = []*domain.Playlist{}

	// Process playlists in batches to respect YouTube API limits
	for i := 0; i < len(ids); i += MAX_BATCH_SIZE {
		end := i + MAX_BATCH_SIZE
		if end > len(ids) {
			end = len(ids)
		}

		batchIds := ids[i:end]
		lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).
			MaxResults(int64(len(batchIds))).
			Id(batchIds...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		if len(lists.Items) == 0 {
			if len(ids) == 1 {
				return nil, domain.ErrYouTubePlaylistNotFound
			}
			continue
		}

		// Process each playlist
		for _, playlist := range lists.Items {
			// Fetch all playlist items with pagination
			var allPlaylistItems []*youtube.PlaylistItem
			nextPageToken := ""

			for {
				call := r.youtube.PlaylistItems.List([]string{"snippet"}).
					MaxResults(MAX_RESULTS_PER_PAGE).
					PlaylistId(playlist.Id)

				if nextPageToken != "" {
					call = call.PageToken(nextPageToken)
				}

				playlistItems, err := call.Do()
				if err != nil {
					return nil, fmt.Errorf("failed to fetch playlist items: %w", err)
				}

				for _, item := range playlistItems.Items {
					if item.Snippet == nil || item.Snippet.ResourceId == nil {
						continue
					}
					allPlaylistItems = append(allPlaylistItems, item)
				}

				nextPageToken = playlistItems.NextPageToken
				if nextPageToken == "" {
					break
				}
			}

			// Skip playlists with no videos
			if len(allPlaylistItems) == 0 {
				log.Printf("Playlist %s has no videos, skipping", playlist.Id)
				continue
			}

			// Collect video IDs
			var vids []string
			for _, item := range allPlaylistItems {
				vids = append(vids, item.Snippet.ResourceId.VideoId)
			}

			// Process video IDs in batches
			var allVideos []*youtube.Video
			for j := 0; j < len(vids); j += MAX_BATCH_SIZE {
				endJ := j + MAX_BATCH_SIZE
				if endJ > len(vids) {
					endJ = len(vids)
				}

				batchVids := vids[j:endJ]
				videos, err := r.youtube.Videos.List([]string{"id", "snippet", "statistics"}).
					MaxResults(int64(len(batchVids))).
					Id(batchVids...).Do()
				if err != nil {
					return nil, fmt.Errorf("failed to fetch videos: %w", err)
				}

				allVideos = append(allVideos, videos.Items...)
			}

			// Skip if all videos were deleted
			if len(allVideos) == 0 {
				log.Printf("All videos in playlist %s have been deleted, skipping", playlist.Id)
				continue
			}

			// Collect channel IDs
			var cids []string
			for _, item := range allVideos {
				if item.Snippet == nil {
					continue
				}
				cids = append(cids, item.Snippet.ChannelId)
			}

			// Remove duplicate channel IDs
			uniqueCids := make(map[string]bool)
			var uniqueCidsList []string
			for _, cid := range cids {
				if !uniqueCids[cid] {
					uniqueCids[cid] = true
					uniqueCidsList = append(uniqueCidsList, cid)
				}
			}

			// Process channel IDs in batches
			var allChannels []*youtube.Channel
			for j := 0; j < len(uniqueCidsList); j += MAX_BATCH_SIZE {
				endJ := j + MAX_BATCH_SIZE
				if endJ > len(uniqueCidsList) {
					endJ = len(uniqueCidsList)
				}

				batchCids := uniqueCidsList[j:endJ]
				channels, err := r.youtube.Channels.List([]string{"id", "snippet"}).
					MaxResults(int64(len(batchCids))).
					Id(batchCids...).Do()
				if err != nil {
					return nil, fmt.Errorf("failed to fetch channels: %w", err)
				}

				allChannels = append(allChannels, channels.Items...)
			}

			// Skip if all channels were deleted
			if len(allChannels) == 0 {
				log.Printf("All channels for playlist %s have been deleted, skipping", playlist.Id)
				continue
			}

			// Create a map for faster lookups
			videoMap := make(map[string]*youtube.Video)
			for _, v := range allVideos {
				videoMap[v.Id] = v
			}

			channelMap := make(map[string]*youtube.Channel)
			for _, c := range allChannels {
				channelMap[c.Id] = c
			}

			var listVideos = []domain.Video{}
			for _, listVideo := range allPlaylistItems {
				videoId := listVideo.Snippet.ResourceId.VideoId
				video, videoExists := videoMap[videoId]
				if !videoExists || video.Snippet == nil {
					// Video was deleted, skip
					continue
				}

				channelId := video.Snippet.ChannelId
				channel, channelExists := channelMap[channelId]
				if !channelExists || channel.Snippet == nil {
					// Channel was deleted, skip
					continue
				}

				publishedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, listVideo.Snippet.PublishedAt)
				if err != nil {
					return nil, fmt.Errorf("failed to parse video publish time: %w", err)
				}

				ownerPublishedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, video.Snippet.PublishedAt)
				if err != nil {
					return nil, fmt.Errorf("failed to parse video owner publish time: %w", err)
				}

				listVideos = append(listVideos, domain.Video{
					YoutubeID:        videoId,
					Title:            listVideo.Snippet.Title,
					Views:            videoViews(video),
					Thumbnail:        videoThumbnail(video),
					ChannelName:      channel.Snippet.Title,
					ChannelIcon:      channelIcon(channel),
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
	}

	if len(response) == 0 {
		return nil, domain.ErrYouTubePlaylistNotFound
	}

	return response, nil
}

// Some fields of YouTube API responses may be missing, so read them nil-safely
func videoViews(v *youtube.Video) uint64 {
	if v.Statistics == nil {
		return 0
	}
	return v.Statistics.ViewCount
}

func videoThumbnail(v *youtube.Video) string {
	if v.Snippet == nil || v.Snippet.Thumbnails == nil {
		return ""
	}
	for _, t := range []*youtube.Thumbnail{v.Snippet.Thumbnails.High, v.Snippet.Thumbnails.Medium, v.Snippet.Thumbnails.Default} {
		if t != nil {
			return t.Url
		}
	}
	return ""
}

func channelIcon(c *youtube.Channel) string {
	if c.Snippet == nil || c.Snippet.Thumbnails == nil || c.Snippet.Thumbnails.Default == nil {
		return ""
	}
	return c.Snippet.Thumbnails.Default.Url
}
