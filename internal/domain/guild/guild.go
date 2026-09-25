// Package guild is the aggregate of the Discord servers the bot has joined.
package guild

import "gorm.io/gorm"

// DiscordID is the ID Discord gives to a guild (server).
type DiscordID string

// Guild is a Discord server the bot has joined.
type Guild struct {
	gorm.Model
	DiscordID DiscordID
}

func (Guild) TableName() string {
	return "guilds"
}

func New(id DiscordID) *Guild {
	return &Guild{DiscordID: id}
}
