package fake

import (
	"slices"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// YouTube は library.YouTube の偽物。呼ばれた回数と、詳細を取得した動画を記録する。
type YouTube struct {
	Playlists []*Playlist
	// true なら FindPlaylist が失敗する
	FindFails bool
	// FetchPlaylistItems に失敗させるプレイリスト
	ItemFails []library.PlaylistID
	// FetchVideos が返す動画の再生数。ないものは 0
	Views map[library.VideoID]library.ViewCount

	ItemFetches   int
	VideoRequests [][]library.VideoID
}

// Playlist は YouTube 上のプレイリスト。
type Playlist struct {
	ID    library.PlaylistID
	Title string
	Items []*library.FetchedItem
}

func (y *YouTube) find(id library.PlaylistID) *Playlist {
	for _, p := range y.Playlists {
		if p.ID == id {
			return p
		}
	}

	return nil
}

func (y *YouTube) FindPlaylist(id library.PlaylistID) (*library.PlaylistMeta, error) {
	if y.FindFails {
		return nil, Err
	}

	p := y.find(id)
	if p == nil {
		return nil, domain.ErrYouTubePlaylistNotFound
	}

	return &library.PlaylistMeta{YoutubeID: p.ID, Title: p.Title}, nil
}

func (y *YouTube) FetchPlaylistMetas(ids ...library.PlaylistID) (map[library.PlaylistID]*library.PlaylistMeta, error) {
	metas := map[library.PlaylistID]*library.PlaylistMeta{}

	for _, id := range ids {
		if p := y.find(id); p != nil {
			metas[id] = &library.PlaylistMeta{YoutubeID: p.ID, Title: p.Title, ItemCount: int64(len(p.Items))}
		}
	}

	return metas, nil
}

func (y *YouTube) FetchPlaylistItems(id library.PlaylistID) ([]*library.FetchedItem, error) {
	y.ItemFetches++
	if slices.Contains(y.ItemFails, id) {
		return nil, Err
	}

	var items []*library.FetchedItem
	for _, item := range y.find(id).Items {
		items = append(items, clone(item))
	}

	return items, nil
}

// FetchVideos は動画 ID をタイトルにした詳細を返す。
func (y *YouTube) FetchVideos(ids []library.VideoID) ([]*library.Video, []*library.Channel, error) {
	y.VideoRequests = append(y.VideoRequests, ids)

	var videos []*library.Video
	for _, id := range ids {
		videos = append(videos, &library.Video{YoutubeID: id, Title: string(id), Views: y.Views[id]})
	}

	return videos, nil, nil
}

func (y *YouTube) FetchCategories() ([]*library.Category, error) {
	return nil, nil
}

func (y *YouTube) FetchLiveVideos([]library.VideoID) (map[library.VideoID]*library.LiveVideo, error) {
	return map[library.VideoID]*library.LiveVideo{}, nil
}
