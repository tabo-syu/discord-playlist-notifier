package application

import (
	"errors"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

type SubscriptionService struct {
	youtube       library.YouTube
	subscriptions subscription.Repository
	guilds        guild.Repository
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewSubscriptionService(y library.YouTube, s subscription.Repository, g guild.Repository, now func() time.Time) *SubscriptionService {
	return &SubscriptionService{y, s, g, now}
}

func (s *SubscriptionService) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	return s.subscriptions.FindByGuild(guildID)
}

// Subscribe はプレイリストのチャンネルへの通知を始める。
func (s *SubscriptionService) Subscribe(guildID guild.DiscordID, channelID subscription.ChannelID, playlistID library.PlaylistID) error {
	playlist, err := s.youtube.FindPlaylist(playlistID)
	if errors.Is(err, domain.ErrYouTubePlaylistNotFound) {
		return err
	}

	if err != nil {
		return domain.ErrYouTubeGeneralError
	}

	exists, err := s.subscriptions.Exists(guildID, playlist.YoutubeID)
	if err != nil {
		return err
	}

	if exists {
		return domain.ErrDBRecordAlreadyCreated
	}

	g, err := s.guilds.FindByDiscordID(guildID)
	if err != nil {
		return err
	}

	return s.subscriptions.Add(subscription.New(g, channelID, playlist.YoutubeID, playlist.Title, s.now()))
}

// Unsubscribe はギルドへのプレイリストの通知をやめる。
func (s *SubscriptionService) Unsubscribe(guildID guild.DiscordID, playlistID library.PlaylistID) error {
	exists, err := s.subscriptions.Exists(guildID, playlistID)
	if err != nil {
		return err
	}

	if !exists {
		return domain.ErrDBRecordNotFound
	}

	subscriptions, err := s.subscriptions.FindByGuild(guildID)
	if err != nil {
		return err
	}

	var target *subscription.Subscription

	for _, sub := range subscriptions {
		if sub.YoutubeID == playlistID {
			target = sub
		}
	}

	return s.subscriptions.DeleteAll([]*subscription.Subscription{target})
}

func (s *SubscriptionService) SetPickInterval(guildID guild.DiscordID, playlistID library.PlaylistID, interval subscription.PickInterval) error {
	subscriptions, err := s.subscriptions.FindByGuild(guildID)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		if sub.YoutubeID == playlistID {
			sub.ChangePickInterval(interval)

			return s.subscriptions.Update(sub)
		}
	}

	return domain.ErrDBRecordNotFound
}
