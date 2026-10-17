// Package guild はボットが参加している Discord サーバーの集約。
package guild

import "gorm.io/gorm"

// DiscordID は Discord がギルド（サーバー）に割り当てる ID。
type DiscordID string

// Guild はボットが参加している Discord サーバー。
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
