// Package application はボットのユースケースを持つ。リポジトリから集約を
// 読み込み、ドメインに判断させ、その結果を保存する。
package application

import (
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
		return guild.ErrAlreadyRegistered
	}

	return s.guilds.Add(guild.New(id))
}

// Unregister はボットが退出したギルドを、その通知設定と一緒に削除する。
//
// ギルドと通知設定は別の集約なので、1 つのトランザクションにせず、通知設定を先に消す。
// 通知設定の削除に失敗したときは、何も消していない（GuildDelete を受け取らなかったのと
// 同じ）状態で、ギルドの削除に失敗したときは、通知設定のないギルドが残るだけになる。
// 逆の順にすると、通知設定の削除に失敗したとき、ギルドだけが論理削除されて通知設定が残る。
// その通知設定は送り続けるのに、ギルドを招待し直すと新しい ID のギルドになるので、
// 一覧にも出ず、コマンドで消せなくなる。
func (s *GuildService) Unregister(id guild.DiscordID) error {
	g, err := s.guilds.FindByDiscordID(id)
	if err != nil {
		return err
	}

	subscriptions, err := s.subscriptions.FindByGuild(id)
	if err != nil {
		return err
	}

	err = s.subscriptions.DeleteAll(subscriptions)
	if err != nil {
		return err
	}

	return s.guilds.Delete(g)
}
