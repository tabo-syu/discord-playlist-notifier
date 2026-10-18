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
	Wrapped    []*WrappedPost
	Picks      []*PickPost
}

// WrappedPost は投稿した年間のまとめ。
type WrappedPost struct {
	Subscription *subscription.Subscription
	Wrapped      *library.Wrapped
}

// PickPost は投稿したランダムな一曲。
type PickPost struct {
	Subscription *subscription.Subscription
	Picked       *library.PlaylistVideo
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

func (n *Notifier) PostWrapped(sub *subscription.Subscription, wrapped *library.Wrapped) error {
	if n.Fails {
		return Err
	}
	n.Wrapped = append(n.Wrapped, &WrappedPost{Subscription: sub, Wrapped: wrapped})
	return nil
}

func (n *Notifier) PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, _ map[library.VideoID]*library.LiveVideo) error {
	if n.Fails {
		return Err
	}
	n.Picks = append(n.Picks, &PickPost{Subscription: sub, Picked: picked})
	return nil
}
