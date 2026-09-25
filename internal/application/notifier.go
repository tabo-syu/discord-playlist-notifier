package application

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Notifier posts to the Discord channels of the subscriptions. live holds the
// channel and current views fetched right before posting; it may be nil or
// miss some videos, and then they are posted without them.
type Notifier interface {
	// NotifyNewVideos posts one message per video, and keeps posting the rest even if one fails.
	NotifyNewVideos(news *notification.NewVideos, live map[library.VideoID]*library.LiveVideo) error
	NotifyHiddenVideo(notice *notification.HiddenVideo) error
	NotifyMilestone(notice *notification.Milestone, live map[library.VideoID]*library.LiveVideo) error
	PostWrapped(sub *subscription.Subscription, wrapped *library.Wrapped) error
	PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, live map[library.VideoID]*library.LiveVideo) error
}
