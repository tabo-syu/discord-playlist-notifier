package usecases

import "discord-playlist-notifier/src/domain"

type PlaylistRepository interface {
	Insert(id string) (*domain.Playlist, error)
	FindById(id string) (*domain.Playlist, error)
	Delete(id string) error
	Exists(id string) (bool, error)
}
