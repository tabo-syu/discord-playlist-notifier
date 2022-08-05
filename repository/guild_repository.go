package repository

import (
	"discord-playlist-notifier/domain"

	"gorm.io/gorm"
)

type GuildRepository interface {
	Add(string) error
	GetByDiscordId(string) (*domain.Guild, error)
}

type guildRepository struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) GuildRepository {
	return &guildRepository{db}
}

func (r *guildRepository) Add(guildId string) error {
	result := r.db.Save(&domain.Guild{DiscordID: guildId})
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *guildRepository) GetByDiscordId(guildId string) (*domain.Guild, error) {
	var guild domain.Guild
	result := r.db.Where(&domain.Guild{DiscordID: guildId}).Take(&guild)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &guild, nil
}
