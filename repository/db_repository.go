package repository

import (
	"discord-playlist-notifier/domain"

	"gorm.io/gorm"
)

type DBRepository interface {
	// Guild
	ExistsGuild(string) (bool, error)
	AddGuild(*domain.Guild) error
	// Playlist
	ExistsPlaylist(string, string) (bool, error)
}

type dbRepository struct {
	db *gorm.DB
}

func NewDBRepository(db *gorm.DB) *dbRepository {
	return &dbRepository{db}
}

// Guild
func (r *dbRepository) ExistsGuild(guildId string) (bool, error) {
	return false, nil
}

func (r *dbRepository) AddGuild(guild *domain.Guild) error {
	return nil
}

// Playlist
func (r *dbRepository) ExistsPlaylist(guildId string, playlistId string) (bool, error) {
	return false, nil
}
