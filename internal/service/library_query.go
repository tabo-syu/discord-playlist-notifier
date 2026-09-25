package service

import (
	"math/rand/v2"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

// PlaylistVideo is a video together with the playlist it was found in.
type PlaylistVideo struct {
	PlaylistID string
	Item       *domain.PlaylistItem
	Video      *domain.YouTubeVideo
}

// Contents returns the stored items of the given playlists, ordered by the
// time they were added, together with their videos.
func (s *LibraryService) Contents(playlistIds []string) ([]*PlaylistVideo, error) {
	items, err := s.library.FindItems(unique(playlistIds)...)
	if err != nil {
		return nil, err
	}

	var videoIds []string
	for _, item := range items {
		videoIds = append(videoIds, item.VideoYoutubeID)
	}
	videos, err := s.library.FindVideos(unique(videoIds))
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

// RandomVideos picks up to n distinct available videos from the given playlists.
func (s *LibraryService) RandomVideos(playlistIds []string, n int) ([]*PlaylistVideo, error) {
	contents, err := s.Contents(playlistIds)
	if err != nil {
		return nil, err
	}

	var candidates []*PlaylistVideo
	seen := map[string]bool{}
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

	return candidates, nil
}
