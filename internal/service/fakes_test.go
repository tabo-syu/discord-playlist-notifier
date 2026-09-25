package service

import (
	"sort"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

type fakeYouTube struct {
	metas  map[string]*repository.PlaylistMeta
	items  map[string][]*repository.PlaylistItemInfo
	videos map[string]*domain.YouTubeVideo

	itemCalls  int
	videoCalls [][]string
}

func (f *fakeYouTube) FindPlaylists(...string) ([]*domain.Playlist, error) { return nil, nil }

func (f *fakeYouTube) FetchPlaylistMetas(ids ...string) (map[string]*repository.PlaylistMeta, error) {
	result := map[string]*repository.PlaylistMeta{}
	for _, id := range ids {
		if m, ok := f.metas[id]; ok {
			result[id] = m
		}
	}
	return result, nil
}

func (f *fakeYouTube) FetchPlaylistItems(id string) ([]*repository.PlaylistItemInfo, error) {
	f.itemCalls++
	return f.items[id], nil
}

func (f *fakeYouTube) FetchVideos(ids []string, _ bool) ([]*domain.YouTubeVideo, error) {
	f.videoCalls = append(f.videoCalls, ids)
	var result []*domain.YouTubeVideo
	for _, id := range ids {
		if v, ok := f.videos[id]; ok {
			copied := *v
			result = append(result, &copied)
		}
	}
	return result, nil
}

type fakeLibrary struct {
	items  []*domain.PlaylistItem
	videos map[string]*domain.YouTubeVideo
}

func newFakeLibrary() *fakeLibrary {
	return &fakeLibrary{videos: map[string]*domain.YouTubeVideo{}}
}

func (f *fakeLibrary) FindItems(ids ...string) ([]*domain.PlaylistItem, error) {
	want := map[string]bool{}
	for _, id := range ids {
		want[id] = true
	}
	var result []*domain.PlaylistItem
	for _, item := range f.items {
		if want[item.PlaylistYoutubeID] {
			copied := *item
			result = append(result, &copied)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].AddedAt.Before(result[j].AddedAt) })
	return result, nil
}

func (f *fakeLibrary) FindVideos(ids []string) (map[string]*domain.YouTubeVideo, error) {
	result := map[string]*domain.YouTubeVideo{}
	for _, id := range ids {
		if v, ok := f.videos[id]; ok {
			copied := *v
			result[id] = &copied
		}
	}
	return result, nil
}

func (f *fakeLibrary) FindVideosInPlaylists(ids ...string) (map[string]*domain.YouTubeVideo, error) {
	items, _ := f.FindItems(ids...)
	var videoIds []string
	for _, item := range items {
		videoIds = append(videoIds, item.VideoYoutubeID)
	}
	return f.FindVideos(videoIds)
}

func (f *fakeLibrary) FindListedVideos() (map[string]*domain.YouTubeVideo, error) {
	var ids []string
	for _, item := range f.items {
		ids = append(ids, item.VideoYoutubeID)
	}
	return f.FindVideos(ids)
}

func (f *fakeLibrary) ApplyItems(_ string, added []*domain.PlaylistItem, removed []*domain.PlaylistItem, videos []*domain.YouTubeVideo) error {
	gone := map[string]bool{}
	for _, item := range removed {
		gone[item.ItemID] = true
	}
	var kept []*domain.PlaylistItem
	for _, item := range f.items {
		if !gone[item.ItemID] {
			kept = append(kept, item)
		}
	}
	f.items = append(kept, added...)
	return f.SaveVideos(videos)
}

func (f *fakeLibrary) SaveVideos(videos []*domain.YouTubeVideo) error {
	for _, v := range videos {
		copied := *v
		f.videos[v.YoutubeID] = &copied
	}
	return nil
}
