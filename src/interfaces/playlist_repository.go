package interfaces

import (
	"discord-playlist-notifier/src/domain"
)

type PlaylistRepository struct {
	RedisHandler RedisHandler
}

func (r *PlaylistRepository) FindById(id string) (domain.Playlist, error) {}

func (r *PlaylistRepository) Save(playlist domain.Playlist) error {}
