package repository

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/errs"

	"gorm.io/gorm"
)

type PlaylistRepository interface {
	Exist(guildId string, playlistId string) (bool, error)
	Add(*domain.Playlist) error
	Update(*domain.Playlist) error
	FindAll() ([]*domain.Playlist, error)
	FindByDiscordId(guildId string) ([]*domain.Playlist, error)
	DeleteAll([]*domain.Playlist) error
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

func (r *playlistRepository) FindAll() ([]*domain.Playlist, error) {
	// プレイリストを取得
	var playlists []*domain.Playlist
	if err := r.db.Find(&playlists).Error; err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, errs.ErrDBRecordCouldNotFound
	}

	// プレイリストに紐づく動画も取得
	for _, playlist := range playlists {
		var videos []domain.Video
		err := r.db.Model(&playlist).Association("Videos").Find(&videos)
		if err != nil {
			return nil, err
		}
		playlist.Videos = videos
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
		return nil, errs.ErrDBRecordCouldNotFound
	}

	// プレイリストに紐づく動画も取得
	for _, playlist := range playlists {
		var videos []domain.Video
		err := r.db.Model(&playlist).Association("Videos").Find(&videos)
		if err != nil {
			return nil, err
		}
		playlist.Videos = videos
	}

	return playlists, nil
}

func (r *playlistRepository) Add(playlist *domain.Playlist) error {
	if playlist.ID != 0 {
		return errs.ErrDBRecordAlreadyCreated
	}

	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *playlistRepository) Update(playlist *domain.Playlist) error {
	if playlist.ID == 0 {
		return errs.ErrDBRecordCouldNotFound
	}

	result := r.db.Save(&playlist)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *playlistRepository) DeleteAll(playlists []*domain.Playlist) error {
	var pids, vids []uint
	var videos []domain.Video
	for _, playlist := range playlists {
		if playlist.ID == 0 {
			return errs.ErrDBRecordCouldNotFound
		}
		pids = append(pids, playlist.ID)
		for _, video := range playlist.Videos {
			if video.ID == 0 {
				return errs.ErrDBRecordCouldNotFound
			}
			vids = append(vids, video.ID)
			videos = append(videos, video)
		}
	}

	err := r.db.Delete(playlists, pids).Error
	if err != nil {
		return err
	}
	// 物理削除
	err = r.db.Unscoped().Delete(videos, vids).Error
	if err != nil {
		return err
	}

	return nil
}
