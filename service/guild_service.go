package service

import (
	"discord-playlist-notifier/domain"
	"discord-playlist-notifier/errs"
	"discord-playlist-notifier/repository"
)

type GuildService interface {
	GetByDiscordId(guildId string) (*domain.Guild, error)
	Register(guildId string) error
	Unregister(guildId string) error
}

type guildService struct {
	guild    repository.GuildRepository
	playlist repository.PlaylistRepository
}

func NewGuildService(g repository.GuildRepository, p repository.PlaylistRepository) GuildService {
	return &guildService{g, p}
}

func (s *guildService) GetByDiscordId(guildId string) (*domain.Guild, error) {
	return s.guild.GetByDiscordId(guildId)
}

func (s *guildService) Register(guildId string) error {
	guildExist, err := s.guild.Exist(guildId)
	if err != nil {
		return err
	}
	if guildExist {
		return errs.ErrDBRecordAlreadyCreated
	}

	return s.guild.Add(guildId)
}

func (s *guildService) Unregister(guildId string) error {
	guildExist, err := s.guild.Exist(guildId)
	if err != nil {
		return err
	}
	if !guildExist {
		return errs.ErrDBRecordCouldNotFound
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
