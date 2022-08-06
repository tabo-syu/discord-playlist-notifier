package service

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/repository"
	"errors"
)

type PlaylistService interface {
	FindByDiscordId(guildId string) ([]*domain.Playlist, error)
	Register(guildId string, playlistId string, needMention bool) error
	Unregister(guildId string, playlistId string) error
}

type playlistService struct {
	youtube  repository.YouTubeRepository
	playlist repository.PlaylistRepository
	guild    repository.GuildRepository
}

func NewPlaylistService(y repository.YouTubeRepository, p repository.PlaylistRepository, g repository.GuildRepository) PlaylistService {
	return &playlistService{y, p, g}
}

func (s *playlistService) FindByDiscordId(guildId string) ([]*domain.Playlist, error) {
	return s.playlist.FindByDiscordId(guildId)
}

func (s *playlistService) Register(guildId string, playlistId string, needMention bool) error {
	playlists, err := s.youtube.FindPlaylists(playlistId)
	if errors.Is(err, errs.ErrYouTubePlaylistCouldNotFound) {
		return err
	}
	if err != nil {
		return errs.ErrYouTubeGeneralError
	}

	playlist := playlists[0]
	playlist.Mention = needMention

	playlistExist, err := s.playlist.Exist(guildId, playlist.YoutubeID)
	if err != nil {
		return err
	}
	if playlistExist {
		return errs.ErrDBRecordAlreadyCreated
	}

	guild, err := s.guild.GetByDiscordId(guildId)
	if err != nil {
		return err
	}
	playlist.Guild = *guild

	return s.playlist.Add(playlist)
}

func (s *playlistService) Unregister(guildId string, playlistId string) error {
	playlistExist, err := s.playlist.Exist(guildId, playlistId)
	if err != nil {
		return err
	}
	if !playlistExist {
		return errs.ErrDBRecordCouldNotFound
	}

	playlists, err := s.playlist.FindByDiscordId(guildId)
	if err != nil {
		return err
	}

	var target *domain.Playlist
	for _, playlist := range playlists {
		if playlist.YoutubeID == playlistId {
			target = playlist
		}
	}

	err = s.playlist.DeleteAll([]*domain.Playlist{target})
	if err != nil {
		return err
	}

	return nil
}
