package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestSubscriptionServiceSubscribe(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *deps)
		wantErr error
	}{
		{
			name: "YouTube のタイトルで登録する",
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.subscriptions.EXPECT().Exists(guild.DiscordID("guild"), library.PlaylistID("PL")).Return(false, nil)
				m.guilds.EXPECT().FindByDiscordID(guild.DiscordID("guild")).Return(guildOf(1, "guild"), nil)
				m.subscriptions.EXPECT().Add(&subscription.Subscription{
					YoutubeID:     "PL",
					SendChannelID: "ch",
					Title:         "タイトル",
					GuildID:       1,
				}).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "登録済みのプレイリストは登録しない",
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.subscriptions.EXPECT().Exists(guild.DiscordID("guild"), library.PlaylistID("PL")).Return(true, nil)
			},
			wantErr: domain.ErrDBRecordAlreadyCreated,
		},
		{
			name: "YouTube にないプレイリストは登録しない",
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, domain.ErrYouTubePlaylistNotFound)
			},
			wantErr: domain.ErrYouTubePlaylistNotFound,
		},
		{
			name: "YouTube API の失敗は一般的なエラーにする",
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, errMock)
			},
			wantErr: domain.ErrYouTubeGeneralError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)
			s := application.NewSubscriptionService(m.youtube, m.subscriptions, m.guilds)

			err := s.Subscribe("guild", "ch", "PL")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSubscriptionServiceUnsubscribe(t *testing.T) {
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		setup      func(m *deps)
		wantErr    error
	}{
		{
			name:       "指定したプレイリストだけ解除する",
			playlistID: "PL1",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().Exists(guild.DiscordID("guild"), library.PlaylistID("PL1")).Return(true, nil)
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return([]*subscription.Subscription{
					sub(1, "PL1", "ch", "PL1", lastNotified),
					sub(2, "PL2", "ch", "PL2", lastNotified),
				}, nil)
				m.subscriptions.EXPECT().DeleteAll([]*subscription.Subscription{
					sub(1, "PL1", "ch", "PL1", lastNotified),
				}).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "登録していないプレイリストは解除できない",
			playlistID: "PL2",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().Exists(guild.DiscordID("guild"), library.PlaylistID("PL2")).Return(false, nil)
			},
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)
			s := application.NewSubscriptionService(m.youtube, m.subscriptions, m.guilds)

			err := s.Unsubscribe("guild", tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSubscriptionServiceSetPickInterval(t *testing.T) {
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		setup      func(m *deps)
		wantErr    error
	}{
		{
			name:       "ピックの間隔だけを保存する",
			playlistID: "PL",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.subscriptions.EXPECT().UpdatePickInterval(picking(1, "PL", subscription.PickWeekly)).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "登録していないプレイリストは変えられない",
			playlistID: "OTHER",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
			},
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)
			s := application.NewSubscriptionService(m.youtube, m.subscriptions, m.guilds)

			err := s.SetPickInterval("guild", tt.playlistID, subscription.PickWeekly)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
