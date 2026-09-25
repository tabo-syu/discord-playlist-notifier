// Package library holds the contents of the watched YouTube playlists: their
// items and videos. Unlike subscriptions, which are per-guild settings, the
// contents are stored once per YouTube playlist and shared by every guild.
package library

// PlaylistID is the ID of a YouTube playlist.
type PlaylistID string

// VideoID is the ID of a YouTube video.
type VideoID string

// ItemID is the ID of a playlist item, which differs from the ID of its video.
type ItemID string
