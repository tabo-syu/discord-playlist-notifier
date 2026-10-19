package application

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Notifier は通知設定の Discord チャンネルに投稿する。live は投稿の直前に取得した
// チャンネルと現在の再生数を持つ。nil のことも一部の動画が欠けていることもあり、
// その場合はそれらなしで投稿する。
type Notifier interface {
	// NotifyNewVideos は動画 1 件につき 1 通投稿し、1 件失敗しても残りの投稿を続ける。
	NotifyNewVideos(added *subscription.VideosAdded, live map[library.VideoID]*library.LiveVideo) error
	NotifyHiddenVideo(notice *notification.HiddenVideo) error
	NotifyMilestone(notice *notification.Milestone, live map[library.VideoID]*library.LiveVideo) error
	PostWrapped(sub *subscription.Subscription, wrapped *library.Wrapped) error
	PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, live map[library.VideoID]*library.LiveVideo) error
}
