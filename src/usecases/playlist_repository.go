package usecases

import "discord-playlist-notifier/src/domain"

type PlaylistRepository interface {
	FindById(string) (domain.Playlist, error)
	Save(domain.Playlist) error
}
