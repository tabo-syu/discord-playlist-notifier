package fake

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Notifier は application.Notifier の偽物。送った通知を記録する。
type Notifier struct {
	// true なら送信に失敗する
	Fails bool

	NewVideos  []*notification.NewVideos
	Hidden     []*notification.HiddenVideo
	Milestones []*notification.Milestone
}

func (n *Notifier) NotifyNewVideos(news *notification.NewVideos, _ map[library.VideoID]*library.LiveVideo) error {
	if n.Fails {
		return Err
	}
	n.NewVideos = append(n.NewVideos, news)
	return nil
}

func (n *Notifier) NotifyHiddenVideo(notice *notification.HiddenVideo) error {
	if n.Fails {
		return Err
	}
	n.Hidden = append(n.Hidden, notice)
	return nil
}

func (n *Notifier) NotifyMilestone(notice *notification.Milestone, _ map[library.VideoID]*library.LiveVideo) error {
	if n.Fails {
		return Err
	}
	n.Milestones = append(n.Milestones, notice)
	return nil
}

func (n *Notifier) PostWrapped(*subscription.Subscription, *library.Wrapped) error {
	return nil
}

func (n *Notifier) PostPick(*subscription.Subscription, *library.PlaylistVideo, map[library.VideoID]*library.LiveVideo) error {
	return nil
}
