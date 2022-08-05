package repository

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"

	"gorm.io/gorm"
)

type PlaylistRepository interface {
	Exist(guildId string, playlistId string) (bool, error)
	Add(*domain.Playlist) error
	FindByDiscordId(guildId string) (*[]domain.Playlist, error)
}

type playlistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) PlaylistRepository {
	return &playlistRepository{db}
}

func (r *playlistRepository) Exist(guildId string, playlistId string) (bool, error) {
	var guild domain.Guild
	r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	err := r.db.Model(&guild).Where(domain.Playlist{YoutubeID: playlistId}).Association("Playlists").Find(&guild.Playlists)
	if err != nil {
		return false, err
	}

	if len(guild.Playlists) == 0 {
		return false, nil
	}

	return true, nil
}

func (r *playlistRepository) FindByDiscordId(guildId string) (*[]domain.Playlist, error) {
	var guild domain.Guild
	var playlists []domain.Playlist
	r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	err := r.db.Model(&guild).Association("Playlists").Find(&playlists)
	if err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, errs.ErrDBRecordCouldNotFound
	}

	return &playlists, nil
}

func (r *playlistRepository) Add(playlist *domain.Playlist) error {
	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
