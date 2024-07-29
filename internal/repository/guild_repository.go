package repository

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"gorm.io/gorm"
)

type GuildRepository interface {
	Exist(guildId string) (bool, error)
	Add(guildId string) error
	GetByDiscordId(guildId string) (*domain.Guild, error)
	Delete(guild *domain.Guild) error
}

type guildRepository struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) GuildRepository {
	return &guildRepository{db}
}

func (r *guildRepository) Exist(guildId string) (bool, error) {
	var guild domain.Guild
	var count int64
	err := r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, err
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

func (r *guildRepository) Delete(guild *domain.Guild) error {
	result := r.db.Delete(guild)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
