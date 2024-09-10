package service

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

type GuildService struct {
	guild    repository.GuildRepository
	playlist repository.PlaylistRepository
}

func NewGuildService(g repository.GuildRepository, p repository.PlaylistRepository) *GuildService {
	return &GuildService{g, p}
}

func (s *GuildService) GetByDiscordId(guildId string) (*domain.Guild, error) {
	return s.guild.GetByDiscordId(guildId)
}

func (s *GuildService) Register(guildId string) error {
	guildExist, err := s.guild.Exist(guildId)
	if err != nil {
		return err
	}
	if guildExist {
		return domain.ErrDBRecordAlreadyCreated
	}

	return s.guild.Add(guildId)
}

func (s *GuildService) Unregister(guildId string) error {
	guildExist, err := s.guild.Exist(guildId)
	if err != nil {
		return err
	}
	if !guildExist {
		return domain.ErrDBRecordCouldNotFound
	}

	guild, err := s.guild.GetByDiscordId(guildId)
	if err != nil {
		return err
	}
	playlists, err := s.playlist.FindByDiscordId(guildId)
	if err != nil {
		return err
	}

	err = s.guild.Delete(guild)
	if err != nil {
		return err
	}

	return s.playlist.DeleteAll(playlists)
}
