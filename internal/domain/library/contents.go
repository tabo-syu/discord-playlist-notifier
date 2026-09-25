package library

import "math/rand/v2"

// PlaylistVideo is a video together with the playlist item it was found in.
type PlaylistVideo struct {
	PlaylistID PlaylistID
	Item       *PlaylistItem
	Video      *Video
}

// Snapshot is the current content of a YouTube playlist.
type Snapshot struct {
	YoutubeID PlaylistID
	Title     string
	// Deleted is true when the playlist no longer exists on YouTube.
	Deleted bool
	Items   []*PlaylistItem
	Videos  map[VideoID]*Video
	// Videos that became private or deleted in this sync
	Hidden []*Video
}

// Contains reports whether the playlist has an item of the video.
func (s *Snapshot) Contains(videoID VideoID) bool {
	for _, item := range s.Items {
		if item.VideoYoutubeID == videoID {
			return true
		}
	}

	return false
}

// PickRandom picks up to n distinct available videos from the contents.
func PickRandom(contents []*PlaylistVideo, n int) []*PlaylistVideo {
	var candidates []*PlaylistVideo
	seen := map[VideoID]bool{}
	for _, c := range contents {
		if c.Video.Available() && !seen[c.Video.YoutubeID] {
			seen[c.Video.YoutubeID] = true
			candidates = append(candidates, c)
		}
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}

	return candidates
}
