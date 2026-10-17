package library

import "time"

// CategoryID は YouTube の動画カテゴリの ID。例: 音楽は "10"
type CategoryID string

// Category は YouTube の動画カテゴリで、ボットの地域と言語で YouTube が表示する
// 名前を持つ。動画は ID で参照する。
type Category struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	YoutubeID CategoryID `gorm:"uniqueIndex"`
	Title     string
	// 新しい動画に設定できるか。そうでないものは、古い動画に設定されたままに
	// なっていることがある古いカテゴリ。
	Assignable bool
}

func (Category) TableName() string {
	return "categories"
}
