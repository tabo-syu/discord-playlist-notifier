package library

// YouTube is where the contents of playlists come from. The daily quota is
// limited, so prefer the stored contents (Repository) when they are enough.
type YouTube interface {
	// FindPlaylist returns domain.ErrYouTubePlaylistNotFound when the playlist
	// does not exist or is private. The item count is not fetched.
	FindPlaylist(id PlaylistID) (*PlaylistMeta, error)
	// FetchPlaylistMetas omits the playlists that no longer exist.
	FetchPlaylistMetas(ids ...PlaylistID) (map[PlaylistID]*PlaylistMeta, error)
	FetchPlaylistItems(id PlaylistID) ([]*FetchedItem, error)
	// FetchVideos returns details of the given videos. Videos that cannot be
	// seen (private, deleted) are missing from the result.
	FetchVideos(ids []VideoID) ([]*Video, error)
	// FetchLiveVideos returns the channel and current views of the given
	// videos, keyed by video ID. Videos that cannot be seen are missing.
	FetchLiveVideos(ids []VideoID) (map[VideoID]*LiveVideo, error)
}
