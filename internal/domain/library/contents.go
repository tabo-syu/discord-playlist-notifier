package library

import "math/rand/v2"

// PlaylistVideo は動画と、それが見つかったプレイリストのアイテム。
type PlaylistVideo struct {
	PlaylistID PlaylistID
	Item       *PlaylistItem
	Video      *Video
}

// Snapshot はプレイリストを同期した結果。保存した Playlist に、YouTube と DB から
// その時点で読んだものを足したもの。
type Snapshot struct {
	*Playlist

	// YouTube 上のプレイリストの現在のタイトル
	Title string
	// Deleted は YouTube 上にプレイリストがもう存在しないときに真になる。
	Deleted bool
	Videos  map[VideoID]*Video
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
