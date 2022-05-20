package domain

import "time"

type Video struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
}
