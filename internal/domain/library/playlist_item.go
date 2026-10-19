package library

import "time"

// PlaylistItem は YouTube のプレイリストの 1 項目。
type PlaylistItem struct {
	PlaylistYoutubeID PlaylistID
	ItemID            ItemID
	VideoYoutubeID    VideoID
	// 動画がプレイリストに追加された時刻
	AddedAt time.Time
}
