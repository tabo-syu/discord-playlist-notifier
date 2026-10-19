package library

import (
	"slices"
	"sort"
)

// Catalog は、保存済みのプレイリストの中身を、Playlist と Video の 2 つの集約を
// またいで引くドメインサービス。どちらの集約のメソッドにもならない読み取りと、
// 「指定した範囲のプレイリストにある動画だけを見せる」判断を持つ。
//
// アプリケーションサービスは、互いを呼ぶ代わりにこれを共有する。YouTube は
// 呼ばないので、クォータは使わない。
type Catalog struct {
	playlists PlaylistRepository
	videos    VideoRepository
}

func NewCatalog(p PlaylistRepository, v VideoRepository) *Catalog {
	return &Catalog{p, v}
}

// Contents は指定したプレイリストの保存済みのアイテムを、追加された時刻の順に
// 動画と一緒に返す。動画を保存していないアイテムは除く。
func (c *Catalog) Contents(playlistIDs []PlaylistID) ([]*PlaylistVideo, error) {
	ids := unique(playlistIDs)

	var items []*PlaylistItem

	for _, id := range ids {
		p, err := c.playlists.Find(id)
		if err != nil {
			return nil, err
		}

		items = append(items, p.Items...)
	}
	// 各プレイリストのアイテムは追加された時刻の順なので、まとめて並べ直す
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].AddedAt.Before(items[j].AddedAt)
	})

	videos, err := c.videos.FindInPlaylists(ids...)
	if err != nil {
		return nil, err
	}

	var contents []*PlaylistVideo

	for _, item := range items {
		video, ok := videos[item.VideoYoutubeID]
		if !ok {
			continue
		}

		contents = append(contents, &PlaylistVideo{PlaylistID: item.PlaylistYoutubeID, Item: item, Video: video})
	}

	return contents, nil
}

// PlaylistsContaining は指定した動画ごとに、それを含むプレイリストを返す。
// 同じ動画が 1 つのプレイリストに 2 回入っていても、プレイリストは 1 回だけ返す。
func (c *Catalog) PlaylistsContaining(videoIDs []VideoID) (map[VideoID][]PlaylistID, error) {
	playlists, err := c.playlists.FindContaining(videoIDs)
	if err != nil {
		return nil, err
	}

	result := map[VideoID][]PlaylistID{}

	ids := unique(videoIDs)
	for _, p := range playlists {
		for _, id := range ids {
			if p.Contains(id) {
				result[id] = append(result[id], p.YoutubeID)
			}
		}
	}

	return result, nil
}

// FindVideo は、scope のプレイリストのどれかにある動画を返す。scope の外にしか
// ない動画（別のギルドだけが登録したプレイリストの動画など）と、詳細をまだ取得して
// いない動画は、見せられないので ErrVideoNotFound を返す。
func (c *Catalog) FindVideo(scope []PlaylistID, id VideoID) (*Video, error) {
	containing, err := c.PlaylistsContaining([]VideoID{id})
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(containing[id], func(p PlaylistID) bool { return slices.Contains(scope, p) }) {
		return nil, ErrVideoNotFound
	}

	videos, err := c.videos.Find([]VideoID{id})
	if err != nil {
		return nil, err
	}

	video, ok := videos[id]
	if !ok || !video.HasDetails() {
		return nil, ErrVideoNotFound
	}

	return video, nil
}

func unique[T comparable](ids []T) []T {
	seen := map[T]bool{}

	var result []T

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}
