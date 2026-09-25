package repository

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"gorm.io/gorm"
)

type PlaylistRepository interface {
	Exist(guildId string, playlistId string) (bool, error)
	Add(*domain.Playlist) error
	Update(*domain.Playlist) error
	SetDailyPick(playlist *domain.Playlist, enabled bool) error
	FindAll() ([]*domain.Playlist, error)
	FindByDiscordId(guildId string) ([]*domain.Playlist, error)
	DeleteAll([]*domain.Playlist) error
}

type playlistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *playlistRepository {
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

func (r *playlistRepository) FindAll() ([]*domain.Playlist, error) {
	// プレイリストを取得
	var playlists []*domain.Playlist
	if err := r.db.Find(&playlists).Error; err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return playlists, nil
}

func (r *playlistRepository) FindByDiscordId(guildId string) ([]*domain.Playlist, error) {
	// プレイリストを取得
	var guild domain.Guild
	var playlists []*domain.Playlist
	r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild)
	err := r.db.Model(&guild).Association("Playlists").Find(&playlists)
	if err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return playlists, nil
}

func (r *playlistRepository) Add(playlist *domain.Playlist) error {
	if playlist.ID != 0 {
		return domain.ErrDBRecordAlreadyCreated
	}

	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *playlistRepository) Update(playlist *domain.Playlist) error {
	if playlist.ID == 0 {
		return domain.ErrDBRecordNotFound
	}

	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// SetDailyPick must not touch UpdatedAt, which is the base time of new video
// notifications, so UpdateColumn is used instead of Save.
func (r *playlistRepository) SetDailyPick(playlist *domain.Playlist, enabled bool) error {
	if playlist.ID == 0 {
		return domain.ErrDBRecordNotFound
	}

	return r.db.Model(playlist).UpdateColumn("daily_pick", enabled).Error
}

func (r *playlistRepository) DeleteAll(playlists []*domain.Playlist) error {
	var pids []uint
	for _, playlist := range playlists {
		if playlist.ID == 0 {
			return domain.ErrDBRecordNotFound
		}
		pids = append(pids, playlist.ID)
	}

	err := r.db.Delete(playlists, pids).Error
	if err != nil {
		return err
	}

	return nil
}
