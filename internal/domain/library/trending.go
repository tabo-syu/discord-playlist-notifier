package library

import (
	"sort"
	"time"
)

const TRENDING_TOP = 3

// TrendingVideo は期間内に再生数が伸びた動画。
type TrendingVideo struct {
	Video *Video
	// 期間内で最初に記録された状態から最後の状態までに伸びた再生数
	Increase ViewCount
	// 期間内に記録された状態、古い順
	Snapshots []*VideoSnapshot
}

// Trending は期間内に再生数が最も伸びたプレイリストの動画。
type Trending struct {
	// 期間内にいずれかの動画が記録された最初と最後の時刻。何も記録されて
	// いなければゼロ値。
	From time.Time
	To   time.Time
	// 伸びが大きい順
	Videos []*TrendingVideo
}

// ComputeTrending は視聴可能な動画を、最初に記録された状態から最後の状態までに
// 伸びた再生数で順位付けし、上位 n 件を残す。状態が 2 つ未満の動画や伸びが
// ない動画は除く。Snapshots は埋めない。
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
