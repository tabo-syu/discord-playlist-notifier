// Package application はボットのユースケースを持つ。リポジトリから集約を
// 読み込み、ドメインに判断させ、その結果を保存する。
package application

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

type GuildService struct {
	guilds        guild.Repository
	subscriptions subscription.Repository
}

func NewGuildService(g guild.Repository, s subscription.Repository) *GuildService {
	return &GuildService{g, s}
}

// Register はボットが参加したギルドを記録する。
func (s *GuildService) Register(id guild.DiscordID) error {
	exists, err := s.guilds.Exists(id)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrDBRecordAlreadyCreated
	}

	return s.guilds.Add(guild.New(id))
}

// Unregister はボットが退出したギルドを、その通知設定と一緒に削除する。
func (s *GuildService) Unregister(id guild.DiscordID) error {
	exists, err := s.guilds.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrDBRecordNotFound
	}

	g, err := s.guilds.FindByDiscordID(id)
	if err != nil {
		return err
	}
	subscriptions, err := s.subscriptions.FindByGuild(id)
	if err != nil {
		return err
	}

	err = s.guilds.Delete(g)
	if err != nil {
		return err
	}

	return s.subscriptions.DeleteAll(subscriptions)
}
