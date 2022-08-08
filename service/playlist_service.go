package service

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/repository"
	"errors"
	"log"
	"time"
)

type PlaylistService interface {
	FindAll() ([]*domain.Playlist, error)
	FindByGuild(guildId string) ([]*domain.Playlist, error)
	UpdateUpdatedAt(*domain.Playlist, time.Time) error
	Register(guildId string, channelId string, playlistId string) error
	Unregister(guildId string, playlistId string) error
	GetDiffFromLatest(playlist []*domain.Playlist) ([]*domain.Playlist, error)
}

type playlistService struct {
	youtube  repository.YouTubeRepository
	playlist repository.PlaylistRepository
	guild    repository.GuildRepository
}

func NewPlaylistService(y repository.YouTubeRepository, p repository.PlaylistRepository, g repository.GuildRepository) PlaylistService {
	return &playlistService{y, p, g}
}

func (s *playlistService) FindAll() ([]*domain.Playlist, error) {
	return s.playlist.FindAll()
}

func (s *playlistService) FindByGuild(guildId string) ([]*domain.Playlist, error) {
	return s.playlist.FindByDiscordId(guildId)
}

func (s *playlistService) UpdateUpdatedAt(playlist *domain.Playlist, time time.Time) error {
	playlist.UpdatedAt = time

	return s.playlist.Update(playlist)
}

func (s *playlistService) Register(guildId string, channelId string, playlistId string) error {
	playlists, err := s.youtube.FindPlaylists(playlistId)
	if errors.Is(err, errs.ErrYouTubePlaylistCouldNotFound) {
		return err
	}
	if err != nil {
		return errs.ErrYouTubeGeneralError
	}

	playlist := playlists[0]

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
	playlist.SendChannelID = channelId

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

func (s *playlistService) GetDiffFromLatest(lastPlaylists []*domain.Playlist) ([]*domain.Playlist, error) {
	var pids []string
	for _, playlist := range lastPlaylists {
		pids = append(pids, playlist.YoutubeID)
	}
	latestPlaylists, err := s.youtube.FindPlaylistsWithVideos(pids...)
	if err != nil {
		return nil, err
	}

	var updatedPlaylists []*domain.Playlist
	for _, last := range lastPlaylists {
		wasFound := false
		for _, latest := range latestPlaylists {
			if last.YoutubeID != latest.YoutubeID {
				continue
			}
			wasFound = true

			var updated []domain.Video
			for _, video := range latest.Videos {
				if video.PublishedAt.After(last.UpdatedAt) {
					updated = append(updated, video)
				}
			}
			if len(updated) != 0 {
				last.Title = latest.Title
				last.Videos = updated
				updatedPlaylists = append(updatedPlaylists, last)
			}

			break
		}

		if !wasFound {
			log.Println("Playlist(ID:", last.YoutubeID, ") may have been deleted from YouTube")
		}
	}

	return updatedPlaylists, nil
}
