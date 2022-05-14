package domain

type Playlist struct {
	Id     string  `json:"id"`
	Title  string  `json:"title"`
	Videos []Video `json:"videos"`
}
