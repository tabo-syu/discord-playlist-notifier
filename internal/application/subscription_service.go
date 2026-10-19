package application

import (
	"errors"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// ErrYouTubeUnavailable は YouTube からプレイリストを取得できず、ユースケースを
// 続けられなかったことを表す。外部サービスの可用性は業務のルールではないので、
// ドメインではなくアプリケーション層で表す。
var ErrYouTubeUnavailable = errors.New("youtube is unavailable")

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

// Subscribe はプレイリストのチャンネルへの通知を始める。登録済みかは事前に確かめず、
// リポジトリの Add が ErrAlreadySubscribed を返す（同時に登録しても重複しないように）。
func (s *SubscriptionService) Subscribe(guildID guild.DiscordID, channelID subscription.ChannelID, playlistID library.PlaylistID) error {
	playlist, err := s.youtube.FindPlaylist(playlistID)
	if errors.Is(err, library.ErrPlaylistNotFound) {
		return err
	}

	if err != nil {
		return ErrYouTubeUnavailable
	}

	g, err := s.guilds.FindByDiscordID(guildID)
	if err != nil {
		return err
	}

	return s.subscriptions.Add(subscription.New(g, channelID, playlist.YoutubeID, playlist.Title, s.now()))
}

// Unsubscribe はギルドへのプレイリストの通知をやめる。
func (s *SubscriptionService) Unsubscribe(guildID guild.DiscordID, playlistID library.PlaylistID) error {
	sub, err := s.subscriptions.FindByGuildAndPlaylist(guildID, playlistID)
	if err != nil {
		return err
	}

	return s.subscriptions.DeleteAll([]*subscription.Subscription{sub})
}

func (s *SubscriptionService) SetPickInterval(guildID guild.DiscordID, playlistID library.PlaylistID, interval subscription.PickInterval) error {
	sub, err := s.subscriptions.FindByGuildAndPlaylist(guildID, playlistID)
	if err != nil {
		return err
	}

	sub.ChangePickInterval(interval)

	return s.subscriptions.Update(sub)
}
