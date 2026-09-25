package library

// Repository stores the contents of watched YouTube playlists.
type Repository interface {
	// FindItems returns the items of the given playlists, ordered by the time they were added.
	FindItems(playlistIDs ...PlaylistID) ([]*PlaylistItem, error)
	FindVideos(ids []VideoID) (map[VideoID]*Video, error)
	// FindVideosInPlaylists returns the videos of every item in the given playlists.
	FindVideosInPlaylists(playlistIDs ...PlaylistID) (map[VideoID]*Video, error)
	// FindListedVideos returns all videos that are in at least one playlist.
	FindListedVideos() (map[VideoID]*Video, error)
	// FindItemsByVideos returns the items of every playlist that contain the given videos.
	FindItemsByVideos(ids []VideoID) ([]*PlaylistItem, error)
	// ApplyItems replaces the stored items of a playlist with the given ones
	// and saves the videos, in one transaction.
	ApplyItems(playlistID PlaylistID, added []*PlaylistItem, removed []*PlaylistItem, videos []*Video) error
	// SaveVideos overwrites every column of the given videos.
	SaveVideos(videos []*Video) error
}
