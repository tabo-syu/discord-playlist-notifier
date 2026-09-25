package repository

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
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

// PlaylistMeta is the playlist-level information used to decide whether the
// items of a playlist need to be fetched again.
type PlaylistMeta struct {
	YoutubeID string
	Title     string
	ItemCount int64
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
	// seen (private, deleted) are missing from the result. Channel icons are
	// only fetched when withChannelIcons is true, because it costs extra quota.
	FetchVideos(ids []string, withChannelIcons bool) ([]*domain.YouTubeVideo, error)
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
		lists, err := r.youtube.Playlists.List([]string{"id", "snippet", "contentDetails"}).
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
			if p.ContentDetails != nil {
				meta.ItemCount = p.ContentDetails.ItemCount
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

// FetchVideos costs 1 unit per 50 videos, plus 1 unit per 50 channels when
// withChannelIcons is true.
func (r *youTubeRepository) FetchVideos(ids []string, withChannelIcons bool) ([]*domain.YouTubeVideo, error) {
	var videos []*domain.YouTubeVideo
	for _, batch := range batches(ids) {
		res, err := r.youtube.Videos.List([]string{"id", "snippet", "statistics", "contentDetails", "status"}).
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

			video := &domain.YouTubeVideo{
				YoutubeID:     v.Id,
				Title:         v.Snippet.Title,
				ChannelID:     v.Snippet.ChannelId,
				ChannelName:   v.Snippet.ChannelTitle,
				Thumbnail:     videoThumbnail(v),
				Views:         videoViews(v),
				PrivacyStatus: domain.PrivacyPublic,
				PublishedAt:   publishedAt,
			}
			if v.ContentDetails != nil {
				video.Duration = ParseDuration(v.ContentDetails.Duration)
			}
			if v.Status != nil && v.Status.PrivacyStatus != "" {
				video.PrivacyStatus = v.Status.PrivacyStatus
			}
			videos = append(videos, video)
		}
	}

	if !withChannelIcons {
		return videos, nil
	}

	var channelIds []string
	seen := map[string]bool{}
	for _, v := range videos {
		if !seen[v.ChannelID] {
			seen[v.ChannelID] = true
			channelIds = append(channelIds, v.ChannelID)
		}
	}

	icons := map[string]string{}
	for _, batch := range batches(channelIds) {
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

	for _, v := range videos {
		v.ChannelIcon = icons[v.ChannelID]
	}

	return videos, nil
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

var durationPattern = regexp.MustCompile(`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

// ParseDuration parses the ISO 8601 durations used by YouTube (e.g. "PT4M13S").
// Unknown formats are treated as zero.
func ParseDuration(s string) time.Duration {
	m := durationPattern.FindStringSubmatch(s)
	if m == nil {
		return 0
	}

	var d time.Duration
	units := []time.Duration{24 * time.Hour, time.Hour, time.Minute, time.Second}
	for i, unit := range units {
		if m[i+1] == "" {
			continue
		}
		n, err := strconv.ParseInt(m[i+1], 10, 64)
		if err != nil {
			return 0
		}
		d += time.Duration(n) * unit
	}

	return d
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
