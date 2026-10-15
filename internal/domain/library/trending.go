package library

import (
	"sort"
	"time"
)

const TRENDING_TOP = 3

// TrendingVideo is a video whose views grew in a period.
type TrendingVideo struct {
	Video *Video
	// Views gained from the first to the last recorded state in the period
	Increase ViewCount
	// The recorded states in the period, oldest first
	Snapshots []*VideoSnapshot
}

// Trending is the videos of a playlist that gained the most views in a period.
type Trending struct {
	// The first and last time any video was recorded in the period. Zero when
	// nothing was recorded.
	From time.Time
	To   time.Time
	// Most gained first
	Videos []*TrendingVideo
}

// ComputeTrending ranks the available videos by the views gained from their
// earliest to their latest recorded state, and keeps the top n. Videos with
// fewer than 2 states or no gain are left out. Snapshots are not filled.
func ComputeTrending(videos map[VideoID]*Video, earliest, latest map[VideoID]*VideoSnapshot, n int) *Trending {
	t := &Trending{}
	for id, first := range earliest {
		if t.From.IsZero() || first.RecordedAt.Before(t.From) {
			t.From = first.RecordedAt
		}
		if last, ok := latest[id]; ok && last.RecordedAt.After(t.To) {
			t.To = last.RecordedAt
		}

		video, ok := videos[id]
		if !ok || !video.Available() {
			continue
		}
		last, ok := latest[id]
		if !ok || !last.RecordedAt.After(first.RecordedAt) || last.Views <= first.Views {
			continue
		}
		t.Videos = append(t.Videos, &TrendingVideo{Video: video, Increase: last.Views - first.Views})
	}

	sort.Slice(t.Videos, func(i, j int) bool {
		a, b := t.Videos[i], t.Videos[j]
		if a.Increase != b.Increase {
			return a.Increase > b.Increase
		}
		return a.Video.YoutubeID < b.Video.YoutubeID
	})
	if len(t.Videos) > n {
		t.Videos = t.Videos[:n]
	}

	return t
}
