package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"

	"gorm.io/gorm"
)

type guildRepository struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) *guildRepository {
	return &guildRepository{db}
}

func (r *guildRepository) Exists(id guild.DiscordID) (bool, error) {
	var g guild.Guild
	var count int64
	err := r.db.Where(guild.Guild{DiscordID: id}).Find(&g).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, err
}

func (r *guildRepository) Add(g *guild.Guild) error {
	result := r.db.Save(g)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *guildRepository) FindByDiscordID(id guild.DiscordID) (*guild.Guild, error) {
	var g guild.Guild
	result := r.db.Where(&guild.Guild{DiscordID: id}).Take(&g)
	if err := result.Error; err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *guildRepository) Delete(g *guild.Guild) error {
	result := r.db.Delete(g)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
