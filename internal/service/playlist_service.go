package service

import (
	"errors"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

type PlaylistService struct {
	youtube  repository.YouTubeRepository
	playlist repository.PlaylistRepository
	guild    repository.GuildRepository
}

func NewPlaylistService(y repository.YouTubeRepository, p repository.PlaylistRepository, g repository.GuildRepository) *PlaylistService {
	return &PlaylistService{y, p, g}
}

func (s *PlaylistService) FindAll() ([]*domain.Playlist, error) {
	return s.playlist.FindAll()
}

func (s *PlaylistService) FindByGuild(guildId string) ([]*domain.Playlist, error) {
	return s.playlist.FindByDiscordId(guildId)
}

func (s *PlaylistService) UpdateUpdatedAt(playlist *domain.Playlist, time time.Time) error {
	playlist.UpdatedAt = time

	return s.playlist.Update(playlist)
}

func (s *PlaylistService) Register(guildId string, channelId string, playlistId string) error {
	playlists, err := s.youtube.FindPlaylists(playlistId)
	if errors.Is(err, domain.ErrYouTubePlaylistCouldNotFound) {
		return err
	}
	if err != nil {
		return domain.ErrYouTubeGeneralError
	}

	playlist := playlists[0]

	playlistExist, err := s.playlist.Exist(guildId, playlist.YoutubeID)
	if err != nil {
		return err
	}
	if playlistExist {
		return domain.ErrDBRecordAlreadyCreated
	}

	guild, err := s.guild.GetByDiscordId(guildId)
	if err != nil {
		return err
	}
	playlist.Guild = *guild
	playlist.SendChannelID = channelId

	return s.playlist.Add(playlist)
}

func (s *PlaylistService) Unregister(guildId string, playlistId string) error {
	playlistExist, err := s.playlist.Exist(guildId, playlistId)
	if err != nil {
		return err
	}
	if !playlistExist {
		return domain.ErrDBRecordCouldNotFound
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

func (s *PlaylistService) GetDiffFromLatest(lastPlaylists []*domain.Playlist) ([]*domain.Playlist, error) {
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
