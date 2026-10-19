package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestGuildServiceRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *deps)
		wantErr error
	}{
		{
			name: "参加したギルドを記録する",
			setup: func(m *deps) {
				m.guilds.EXPECT().Exists(guild.DiscordID("guild")).Return(false, nil)
				m.guilds.EXPECT().Add(guild.New("guild")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "記録済みのギルドは記録しない",
			setup: func(m *deps) {
				m.guilds.EXPECT().Exists(guild.DiscordID("guild")).Return(true, nil)
			},
			wantErr: guild.ErrAlreadyRegistered,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := application.NewGuildService(m.guilds, m.subscriptions)

			err := s.Register("guild")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestGuildServiceUnregister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *deps)
		wantErr error
	}{
		{
			name: "ギルドをその通知設定と一緒に削除する",
			setup: func(m *deps) {
				m.guilds.EXPECT().FindByDiscordID(guild.DiscordID("guild")).Return(guildOf(1, "guild"), nil)
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				deleted := m.subscriptions.EXPECT().DeleteAll([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}).Return(nil).Call
				// 通知設定を先に消す（GuildService.Unregister のコメント）
				m.guilds.EXPECT().Delete(guildOf(1, "guild")).Return(nil).NotBefore(deleted)
			},
			wantErr: nil,
		},
		{
			name: "通知設定を削除できなければ、ギルドを削除しない",
			setup: func(m *deps) {
				m.guilds.EXPECT().FindByDiscordID(guild.DiscordID("guild")).Return(guildOf(1, "guild"), nil)
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.subscriptions.EXPECT().DeleteAll([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}).Return(errMock)
			},
			wantErr: errMock,
		},
		{
			// 既知の挙動: FindByGuild が空のときにエラーを返すため、途中で止まる
			name: "通知設定のないギルドは削除されない",
			setup: func(m *deps) {
				m.guilds.EXPECT().FindByDiscordID(guild.DiscordID("guild")).Return(guildOf(1, "guild"), nil)
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return(nil, subscription.ErrNotSubscribed)
			},
			wantErr: subscription.ErrNotSubscribed,
		},
		{
			name: "記録していないギルドは削除できない",
			setup: func(m *deps) {
				m.guilds.EXPECT().FindByDiscordID(guild.DiscordID("guild")).Return(nil, guild.ErrNotRegistered)
			},
			wantErr: guild.ErrNotRegistered,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := application.NewGuildService(m.guilds, m.subscriptions)

			err := s.Unregister("guild")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
