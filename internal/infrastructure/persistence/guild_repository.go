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
	var (
		rec   guildRecord
		count int64
	)

	err := r.db.Where(guildRecord{DiscordID: string(id)}).Find(&rec).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, err
}

// Add は保存したギルドに DB が振った ID を g に書き戻す。
func (r *guildRepository) Add(g *guild.Guild) error {
	rec := toGuildRecord(g)

	result := r.db.Save(rec)
	if err := result.Error; err != nil {
		return err
	}

	g.ID = rec.ID

	return nil
}

func (r *guildRepository) FindByDiscordID(id guild.DiscordID) (*guild.Guild, error) {
	var rec guildRecord

	result := r.db.Where(&guildRecord{DiscordID: string(id)}).Take(&rec)
	if err := result.Error; err != nil {
		return nil, err
	}

	return rec.toDomain(), nil
}

func (r *guildRepository) Delete(g *guild.Guild) error {
	result := r.db.Delete(toGuildRecord(g))
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
