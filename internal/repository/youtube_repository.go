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

	// Title YouTube returns for playlist items whose video has been deleted
	DELETED_VIDEO_TITLE = "Deleted video"
)

// PlaylistMeta is the playlist-level information fetched on every sync.
type PlaylistMeta struct {
	YoutubeID string
	Title     string
}

// PlaylistItemInfo is one playlist item as returned by the API.
type PlaylistItemInfo struct {
	ItemID        string
	VideoID       string
	AddedAt       time.Time
	PrivacyStatus string
}

type YouTubeRepository interface {
	FindPlaylists(...string) ([]*domain.Playlist, error)
	FetchPlaylistMetas(...string) (map[string]*PlaylistMeta, error)
	FetchPlaylistItems(playlistId string) ([]*PlaylistItemInfo, error)
	// FetchVideos returns details of the given videos. Videos that cannot be
	// seen (private, deleted) are missing from the result.
	FetchVideos(ids []string) ([]*domain.Video, error)
	// FetchLiveVideos returns the channel and current views of the given
	// videos, keyed by video ID. Videos that cannot be seen are missing.
	FetchLiveVideos(ids []string) (map[string]*domain.LiveVideo, error)
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

// FetchPlaylistMetas costs 1 unit per 50 playlists. Playlists that no longer
// exist are missing from the result.
func (r *youTubeRepository) FetchPlaylistMetas(ids ...string) (map[string]*PlaylistMeta, error) {
	metas := map[string]*PlaylistMeta{}
	for _, batch := range batches(ids) {
		lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		for _, p := range lists.Items {
			meta := &PlaylistMeta{YoutubeID: p.Id}
			if p.Snippet != nil {
				meta.Title = p.Snippet.Title
			}
			metas[p.Id] = meta
		}
	}

	return metas, nil
}

// FetchPlaylistItems costs 1 unit per 50 items.
func (r *youTubeRepository) FetchPlaylistItems(playlistId string) ([]*PlaylistItemInfo, error) {
	var items []*PlaylistItemInfo
	nextPageToken := ""
	for {
		call := r.youtube.PlaylistItems.List([]string{"snippet", "status"}).
			MaxResults(MAX_RESULTS_PER_PAGE).
			PlaylistId(playlistId)
		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}

		page, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist items: %w", err)
		}

		for _, item := range page.Items {
			if item.Snippet == nil || item.Snippet.ResourceId == nil {
				continue
			}

			addedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, item.Snippet.PublishedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse playlist item publish time: %w", err)
			}

			items = append(items, &PlaylistItemInfo{
				ItemID:        item.Id,
				VideoID:       item.Snippet.ResourceId.VideoId,
				AddedAt:       addedAt,
				PrivacyStatus: itemPrivacy(item),
			})
		}

		nextPageToken = page.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return items, nil
}

// FetchVideos costs 1 unit per 50 videos.
func (r *youTubeRepository) FetchVideos(ids []string) ([]*domain.Video, error) {
	var videos []*domain.Video
	for _, batch := range batches(ids) {
		res, err := r.youtube.Videos.List([]string{"id", "snippet", "statistics", "status"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos: %w", err)
		}

		for _, v := range res.Items {
			if v.Snippet == nil {
				continue
			}

			publishedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, v.Snippet.PublishedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse video publish time: %w", err)
			}

			video := &domain.Video{
				YoutubeID:     v.Id,
				Title:         v.Snippet.Title,
				Views:         videoViews(v),
				PrivacyStatus: domain.PrivacyPublic,
				PublishedAt:   publishedAt,
			}
			if v.Status != nil && v.Status.PrivacyStatus != "" {
				video.PrivacyStatus = v.Status.PrivacyStatus
			}
			videos = append(videos, video)
		}
	}

	return videos, nil
}

// FetchLiveVideos costs 1 unit per 50 videos plus 1 unit per 50 channels.
func (r *youTubeRepository) FetchLiveVideos(ids []string) (map[string]*domain.LiveVideo, error) {
	live := map[string]*domain.LiveVideo{}
	channelOf := map[string]string{}
	var channelIds []string
	for _, batch := range batches(ids) {
		res, err := r.youtube.Videos.List([]string{"id", "snippet", "statistics"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos: %w", err)
		}

		for _, v := range res.Items {
			if v.Snippet == nil {
				continue
			}
			live[v.Id] = &domain.LiveVideo{ChannelName: v.Snippet.ChannelTitle, Views: videoViews(v)}
			channelOf[v.Id] = v.Snippet.ChannelId
			channelIds = append(channelIds, v.Snippet.ChannelId)
		}
	}

	icons := map[string]string{}
	for _, batch := range batches(unique(channelIds)) {
		res, err := r.youtube.Channels.List([]string{"id", "snippet"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channels: %w", err)
		}

		for _, c := range res.Items {
			icons[c.Id] = channelIcon(c)
		}
	}
	for id, v := range live {
		v.ChannelIcon = icons[channelOf[id]]
	}

	return live, nil
}

func batches(ids []string) [][]string {
	var result [][]string
	for i := 0; i < len(ids); i += MAX_BATCH_SIZE {
		end := i + MAX_BATCH_SIZE
		if end > len(ids) {
			end = len(ids)
		}
		result = append(result, ids[i:end])
	}

	return result
}

// Private items report "private", deleted items report
// "privacyStatusUnspecified" with the title "Deleted video".
func itemPrivacy(item *youtube.PlaylistItem) string {
	if item.Snippet.Title == DELETED_VIDEO_TITLE {
		return domain.PrivacyDeleted
	}
	if item.Status == nil {
		return domain.PrivacyPublic
	}

	switch item.Status.PrivacyStatus {
	case domain.PrivacyPublic, domain.PrivacyUnlisted, domain.PrivacyPrivate:
		return item.Status.PrivacyStatus
	case "privacyStatusUnspecified":
		return domain.PrivacyDeleted
	default:
		log.Println("Unknown playlist item privacy status:", item.Status.PrivacyStatus, "item:", item.Id)
		return domain.PrivacyPublic
	}
}

// Some fields of YouTube API responses may be missing, so read them nil-safely
func videoViews(v *youtube.Video) uint64 {
	if v.Statistics == nil {
		return 0
	}
	return v.Statistics.ViewCount
}

func channelIcon(c *youtube.Channel) string {
	if c.Snippet == nil || c.Snippet.Thumbnails == nil || c.Snippet.Thumbnails.Default == nil {
		return ""
	}
	return c.Snippet.Thumbnails.Default.Url
}

func unique(ids []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}
