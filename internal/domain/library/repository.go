package library

import "time"

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
	// SaveChannels inserts the given channels, or updates their titles.
	SaveChannels(channels []*Channel) error
	// FindChannels returns the stored channels among the given IDs.
	FindChannels(ids []ChannelID) (map[ChannelID]*Channel, error)
	// AddSnapshots appends the given states of videos to the history.
	AddSnapshots(snapshots []*VideoSnapshot) error
	// FindSnapshots returns the states of the video recorded at or after
	// since, oldest first.
	FindSnapshots(id VideoID, since time.Time) ([]*VideoSnapshot, error)
	// FindEarliestSnapshots returns, for each given video, its first state
	// recorded at or after since. Videos with no such state are missing.
	FindEarliestSnapshots(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
	// FindLatestSnapshots returns, for each given video, its last state
	// recorded at or after since. Videos with no such state are missing.
	FindLatestSnapshots(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
}
