package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestSubscriptionServiceSubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		playlists         []*fake.Playlist
		playlistID        library.PlaylistID
		wantErr           error
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "YouTube のタイトルで登録する",
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
				},
			},
			playlistID: "PL",
			wantErr:    nil,
			wantSubscriptions: []*subscription.Subscription{
				{ID: 1, YoutubeID: "PL", SendChannelID: "ch", Title: "タイトル", GuildID: 1},
			},
		},
		{
			name: "登録済みのプレイリストは登録しない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "other", "タイトル", lastNotified),
			},
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "タイトル",
				},
			},
			playlistID: "PL",
			wantErr:    domain.ErrDBRecordAlreadyCreated,
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "other", "タイトル", lastNotified),
			},
		},
		{
			name:       "YouTube にないプレイリストは登録しない",
			playlistID: "PL",
			wantErr:    domain.ErrYouTubePlaylistNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			guilds := &fake.GuildRepository{Guilds: []*guild.Guild{guildOf(1, "guild")}}
			subscriptions := &fake.SubscriptionRepository{Guilds: guilds, Subscriptions: tt.subscriptions}
			s := application.NewSubscriptionService(&fake.YouTube{Playlists: tt.playlists}, subscriptions, guilds)

			err := s.Subscribe("guild", "ch", tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}

func TestSubscriptionServiceUnsubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		playlistID        library.PlaylistID
		wantErr           error
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "指定したプレイリストだけ解除する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch", "PL1", lastNotified),
				sub(2, "PL2", "ch", "PL2", lastNotified),
			},
			playlistID: "PL1",
			wantErr:    nil,
			wantSubscriptions: []*subscription.Subscription{
				sub(2, "PL2", "ch", "PL2", lastNotified),
			},
		},
		{
			name: "登録していないプレイリストは解除できない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch", "PL1", lastNotified),
			},
			playlistID: "PL2",
			wantErr:    domain.ErrDBRecordNotFound,
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch", "PL1", lastNotified),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			guilds := &fake.GuildRepository{Guilds: []*guild.Guild{guildOf(1, "guild")}}
			subscriptions := &fake.SubscriptionRepository{Guilds: guilds, Subscriptions: tt.subscriptions}
			s := application.NewSubscriptionService(&fake.YouTube{}, subscriptions, guilds)

			err := s.Unsubscribe("guild", tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}

func TestSubscriptionServiceSetPickInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		playlistID        library.PlaylistID
		interval          subscription.PickInterval
		wantErr           error
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "ピックの間隔だけを変え、新着動画の基準時刻は変えない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			playlistID: "PL",
			interval:   subscription.PickWeekly,
			wantErr:    nil,
			wantSubscriptions: []*subscription.Subscription{
				{
					ID:        1,
					UpdatedAt: lastNotified, YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1,
				},
			},
		},
		{
			name: "登録していないプレイリストは変えられない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			playlistID: "OTHER",
			interval:   subscription.PickWeekly,
			wantErr:    domain.ErrDBRecordNotFound,
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			guilds := &fake.GuildRepository{Guilds: []*guild.Guild{guildOf(1, "guild")}}
			subscriptions := &fake.SubscriptionRepository{Guilds: guilds, Subscriptions: tt.subscriptions}
			s := application.NewSubscriptionService(&fake.YouTube{}, subscriptions, guilds)

			err := s.SetPickInterval("guild", tt.playlistID, tt.interval)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}
