package domain

import "time"

type Playlist struct {
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	Videos    []Video   `json:"videos"`
	UpdatedAt time.Time `json:"updated_at"`
}
