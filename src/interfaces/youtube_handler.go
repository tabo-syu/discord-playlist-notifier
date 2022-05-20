package interfaces

import "google.golang.org/api/youtube/v3"

type YouTubeHandler interface {
	Playlists() PlaylistsService
	PlaylistItems() PlaylistItemsService
}

type PlaylistsService interface {
	// TODO: interfaceで返せるようにする
	List(part []string) *youtube.PlaylistsListCall
}

type PlaylistItemsService interface {
	// TODO: interfaceで返せるようにする
	List(part []string) *youtube.PlaylistItemsListCall
}
