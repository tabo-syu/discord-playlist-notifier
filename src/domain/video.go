package domain

import "time"

type Video struct {
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
}
