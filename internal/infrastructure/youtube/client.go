// Package youtube は YouTube Data API v3 で library.YouTube を実装する。
package youtube

import (
	"fmt"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	ytapi "google.golang.org/api/youtube/v3"
)

const (
	YOUTUBE_TIMEFORMAT   = "2006-01-02T15:04:05Z"
	MAX_RESULTS_PER_PAGE = 50 // YouTube API が許す最大値
	MAX_BATCH_SIZE       = 50 // API 呼び出し 1 回あたりの ID の最大数

	// 動画が削除されたプレイリストのアイテムに YouTube が返すタイトル
	DELETED_VIDEO_TITLE = "Deleted video"
)

type client struct {
	youtube *ytapi.Service
}

func NewClient(yt *ytapi.Service) *client {
	return &client{yt}
}

// FindPlaylist は 1 ユニットを消費する。
func (c *client) FindPlaylist(id library.PlaylistID) (*library.PlaylistMeta, error) {
	lists, err := c.youtube.Playlists.List([]string{"id", "snippet"}).
		MaxResults(1).
		Id(string(id)).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlists: %w", err)
	}

	if len(lists.Items) == 0 {
		return nil, domain.ErrYouTubePlaylistNotFound
	}

	playlist := lists.Items[0]
	return &library.PlaylistMeta{
		YoutubeID: library.PlaylistID(playlist.Id),
		Title:     playlist.Snippet.Title,
	}, nil
}

// FetchPlaylistMetas は 50 プレイリストごとに 1 ユニットを消費する。
func (c *client) FetchPlaylistMetas(ids ...library.PlaylistID) (map[library.PlaylistID]*library.PlaylistMeta, error) {
	metas := map[library.PlaylistID]*library.PlaylistMeta{}
	for _, batch := range batches(ids) {
		lists, err := c.youtube.Playlists.List([]string{"id", "snippet", "contentDetails"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		for _, p := range lists.Items {
			id := library.PlaylistID(p.Id)
			meta := &library.PlaylistMeta{YoutubeID: id}
			if p.Snippet != nil {
				meta.Title = p.Snippet.Title
			}
			if p.ContentDetails != nil {
				meta.ItemCount = p.ContentDetails.ItemCount
			}
			metas[id] = meta
		}
	}

	return metas, nil
}

// FetchPlaylistItems は 50 アイテムごとに 1 ユニットを消費する。
func (c *client) FetchPlaylistItems(id library.PlaylistID) ([]*library.FetchedItem, error) {
	var items []*library.FetchedItem
	nextPageToken := ""
	for {
		call := c.youtube.PlaylistItems.List([]string{"snippet", "status"}).
			MaxResults(MAX_RESULTS_PER_PAGE).
			PlaylistId(string(id))
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

			items = append(items, &library.FetchedItem{
				ItemID:        library.ItemID(item.Id),
				VideoID:       library.VideoID(item.Snippet.ResourceId.VideoId),
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

// FetchVideos は 50 動画ごとに 1 ユニットを消費する。
func (c *client) FetchVideos(ids []library.VideoID) ([]*library.Video, error) {
	var videos []*library.Video
	for _, batch := range batches(ids) {
		res, err := c.youtube.Videos.List([]string{"id", "snippet", "statistics", "status"}).
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

			video := &library.Video{
				YoutubeID:     library.VideoID(v.Id),
				Title:         v.Snippet.Title,
				Views:         videoViews(v),
				PrivacyStatus: library.PrivacyPublic,
				PublishedAt:   publishedAt,
			}
			if v.Statistics != nil {
				video.Likes = v.Statistics.LikeCount
				video.Comments = v.Statistics.CommentCount
			}
			if v.Status != nil && v.Status.PrivacyStatus != "" {
				video.PrivacyStatus = library.PrivacyStatus(v.Status.PrivacyStatus)
			}
			videos = append(videos, video)
		}
	}

	return videos, nil
}

// FetchLiveVideos は 50 動画ごとに 1 ユニット、加えて 50 チャンネルごとに 1 ユニットを消費する。
func (c *client) FetchLiveVideos(ids []library.VideoID) (map[library.VideoID]*library.LiveVideo, error) {
	live := map[library.VideoID]*library.LiveVideo{}
	channelOf := map[library.VideoID]string{}
	var channelIds []string
	for _, batch := range batches(ids) {
		res, err := c.youtube.Videos.List([]string{"id", "snippet", "statistics"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos: %w", err)
		}

		for _, v := range res.Items {
			if v.Snippet == nil {
				continue
			}
			id := library.VideoID(v.Id)
			live[id] = &library.LiveVideo{ChannelName: v.Snippet.ChannelTitle, Views: videoViews(v)}
			channelOf[id] = v.Snippet.ChannelId
			channelIds = append(channelIds, v.Snippet.ChannelId)
		}
	}

	icons := map[string]string{}
	for _, batch := range batches(unique(channelIds)) {
		res, err := c.youtube.Channels.List([]string{"id", "snippet"}).
			MaxResults(int64(len(batch))).
			Id(batch...).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channels: %w", err)
		}

		for _, ch := range res.Items {
			icons[ch.Id] = channelIcon(ch)
		}
	}
	for id, v := range live {
		v.ChannelIcon = icons[channelOf[id]]
	}

	return live, nil
}

// batches は ID を API が一度に受け付ける塊に分ける。
func batches[T ~string](ids []T) [][]string {
	var result [][]string
	for i := 0; i < len(ids); i += MAX_BATCH_SIZE {
		end := min(i+MAX_BATCH_SIZE, len(ids))
		batch := make([]string, 0, end-i)
		for _, id := range ids[i:end] {
			batch = append(batch, string(id))
		}
		result = append(result, batch)
	}

	return result
}

// 非公開のアイテムは "private" を返し、削除済みのアイテムは
// "privacyStatusUnspecified" とタイトル "Deleted video" を返す。
func itemPrivacy(item *ytapi.PlaylistItem) library.PrivacyStatus {
	if item.Snippet.Title == DELETED_VIDEO_TITLE {
		return library.PrivacyDeleted
	}
	if item.Status == nil {
		return library.PrivacyPublic
	}

	switch status := library.PrivacyStatus(item.Status.PrivacyStatus); status {
	case library.PrivacyPublic, library.PrivacyUnlisted, library.PrivacyPrivate:
		return status
	case "privacyStatusUnspecified":
		return library.PrivacyDeleted
	default:
		log.Println("Unknown playlist item privacy status:", item.Status.PrivacyStatus, "item:", item.Id)
		return library.PrivacyPublic
	}
}

// YouTube API のレスポンスは一部のフィールドが欠けることがあるので、nil を考慮して読む
func videoViews(v *ytapi.Video) library.ViewCount {
	if v.Statistics == nil {
		return 0
	}
	return library.ViewCount(v.Statistics.ViewCount)
}

func channelIcon(c *ytapi.Channel) string {
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
