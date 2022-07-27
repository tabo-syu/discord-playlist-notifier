package repository

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"

	"gorm.io/gorm"
)

type DBRepository interface {
	// Guild
	ExistGuild(string) (bool, error)
	GetGuildByDiscordId(string) (*domain.Guild, error)
	AddGuild(string) error
	// Playlist
	ExistPlaylist(string, string) (bool, error)
	GetPlaylistsByDiscordId(string) (*[]domain.Playlist, error)
	AddPlaylist(string, *domain.Playlist) error
}

type dbRepository struct {
	db *gorm.DB
}

func NewDBRepository(db *gorm.DB) *dbRepository {
	return &dbRepository{db}
}

// Guild
func (r *dbRepository) ExistGuild(guildId string) (bool, error) {
	var guild domain.Guild
	result := r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	if err := result.Error; err != nil {
		return false, err
	}

	return guild.ID != 0, nil
}

func (r *dbRepository) GetGuildByDiscordId(guildId string) (*domain.Guild, error) {
	guildExist, err := r.ExistGuild(guildId)
	if err != nil {
		return nil, err
	}
	if !guildExist {
		return nil, errs.ErrRecordCouldNotFound
	}

	var guild domain.Guild
	result := r.db.Where(&domain.Guild{DiscordID: guildId}).Take(&guild)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &guild, nil
}

func (r *dbRepository) AddGuild(guildId string) error {
	guildExist, err := r.ExistGuild(guildId)
	if err != nil {
		return err
	}
	if guildExist {
		return errs.ErrRecordAlreadyCreated
	}

	result := r.db.Save(&domain.Guild{DiscordID: guildId})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// Playlist
func (r *dbRepository) ExistPlaylist(guildId string, playlistId string) (bool, error) {
	var guild domain.Guild
	r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	err := r.db.Model(&guild).
		Where(domain.Playlist{YoutubeID: playlistId}).
		Association("Playlists").
		Find(&guild.Playlists)
	if err != nil {
		return false, err
	}

	if len(guild.Playlists) == 0 {
		return false, nil
	}

	return true, nil
}

func (r *dbRepository) GetPlaylistsByDiscordId(guildId string) (*[]domain.Playlist, error) {
	var guild domain.Guild
	var playlists []domain.Playlist
	r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	err := r.db.Model(&guild).
		Association("Playlists").
		Find(&playlists)
	if err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, errs.ErrRecordCouldNotFound
	}

	return &playlists, nil
}

func (r *dbRepository) AddPlaylist(guildId string, playlist *domain.Playlist) error {
	guildExist, err := r.ExistGuild(guildId)
	if err != nil {
		return err
	}

	playlistExist, err := r.ExistPlaylist(guildId, playlist.YoutubeID)
	if err != nil {
		return err
	}
	if playlistExist {
		return errs.ErrRecordAlreadyCreated
	}

	var guild *domain.Guild
	if !guildExist {
		if err = r.AddGuild(guildId); err != nil {
			return err
		}
	}
	guild, err = r.GetGuildByDiscordId(guildId)
	if err != nil {
		return err
	}

	playlist.Guild = *guild
	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
