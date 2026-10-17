package library

import "math/rand/v2"

// PlaylistVideo は動画と、それが見つかったプレイリストのアイテム。
type PlaylistVideo struct {
	PlaylistID PlaylistID
	Item       *PlaylistItem
	Video      *Video
}

// Snapshot は YouTube のプレイリストの現在の中身。
type Snapshot struct {
	YoutubeID PlaylistID
	Title     string
	// Deleted は YouTube 上にプレイリストがもう存在しないときに真になる。
	Deleted bool
	Items   []*PlaylistItem
	Videos  map[VideoID]*Video
	// 今回の同期で非公開・削除になった動画
	Hidden []*Video
}

// Contains はプレイリストにその動画のアイテムがあるかを返す。
func (s *Snapshot) Contains(videoID VideoID) bool {
	for _, item := range s.Items {
		if item.VideoYoutubeID == videoID {
			return true
		}
	}

	return false
}

// PickRandom は中身から、重複のない視聴可能な動画を最大 n 件選ぶ。
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
