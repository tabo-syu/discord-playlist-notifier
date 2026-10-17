package library

import "time"

// CategoryID is the ID of a YouTube video category, e.g. "10" for music.
type CategoryID string

// Category is a YouTube video category, with its name as YouTube shows it in
// the bot's region and language. Videos refer to it by ID.
type Category struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	YoutubeID CategoryID `gorm:"uniqueIndex"`
	Title     string
	// Whether it can be set on new videos. The others are old categories that
	// may still be set on old videos.
	Assignable bool
}

func (Category) TableName() string {
	return "categories"
}
