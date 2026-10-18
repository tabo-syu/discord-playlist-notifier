package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestGuildServiceRegister(t *testing.T) {
	tests := []struct {
		name       string
		guilds     []*guild.Guild
		wantErr    error
		wantGuilds []*guild.Guild
	}{
		{
			name:    "参加したギルドを記録する",
			wantErr: nil,
			wantGuilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
		},
		{
			name: "記録済みのギルドは記録しない",
			guilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
			wantErr: domain.ErrDBRecordAlreadyCreated,
			wantGuilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guilds := &fake.GuildRepository{Guilds: tt.guilds}
			s := application.NewGuildService(guilds, &fake.SubscriptionRepository{Guilds: guilds})

			err := s.Register("guild")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantGuilds, guilds.Guilds)
		})
	}
}

func TestGuildServiceUnregister(t *testing.T) {
	tests := []struct {
		name              string
		guilds            []*guild.Guild
		subscriptions     []*subscription.Subscription
		wantErr           error
		wantGuilds        []*guild.Guild
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "ギルドをその通知設定と一緒に削除する",
			guilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			wantErr:           nil,
			wantGuilds:        []*guild.Guild{},
			wantSubscriptions: []*subscription.Subscription{},
		},
		{
			// 既知の挙動: FindByGuild が空のときにエラーを返すため、途中で止まる
			name: "通知設定のないギルドは削除されない",
			guilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
			wantErr: domain.ErrDBRecordNotFound,
			wantGuilds: []*guild.Guild{
				guildOf(1, "guild"),
			},
		},
		{
			name:    "記録していないギルドは削除できない",
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guilds := &fake.GuildRepository{Guilds: tt.guilds}
			subscriptions := &fake.SubscriptionRepository{Guilds: guilds, Subscriptions: tt.subscriptions}
			s := application.NewGuildService(guilds, subscriptions)

			err := s.Unregister("guild")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantGuilds, guilds.Guilds)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}
