package notification_test

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

var base = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

func at(hours int) time.Time {
	return base.Add(time.Duration(hours) * time.Hour)
}

func sub(playlist library.PlaylistID, channel subscription.ChannelID, notifiedAt time.Time) *subscription.Subscription {
	s := &subscription.Subscription{YoutubeID: playlist, SendChannelID: channel}
	s.MarkNotified(notifiedAt)

	return s
}

func video(id library.VideoID, status library.PrivacyStatus) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: status}
}

func item(id library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{VideoYoutubeID: id, AddedAt: addedAt}
}
