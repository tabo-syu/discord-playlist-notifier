package library

import "time"

// PlaylistItem は YouTube のプレイリストの 1 項目。
type PlaylistItem struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PlaylistYoutubeID PlaylistID `gorm:"index"`
	ItemID            ItemID     `gorm:"uniqueIndex"`
	VideoYoutubeID    VideoID    `gorm:"index"`
	// 動画がプレイリストに追加された時刻
	AddedAt time.Time
}

func (PlaylistItem) TableName() string {
	return "playlist_items"
}
