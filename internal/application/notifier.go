package application

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Notifier は通知設定の Discord チャンネルに投稿する。live は投稿の直前に
// LiveVideoSource から取得したもので、nil のことも一部の動画が欠けていることもあり、
// その場合はそれらなしで投稿する。
type Notifier interface {
	// NotifyNewVideos は動画 1 件につき 1 通投稿し、1 件失敗しても残りの投稿を続ける。
	NotifyNewVideos(added *subscription.VideosAdded, live map[library.VideoID]*LiveVideo) error
	NotifyHiddenVideo(notice *notification.HiddenVideo) error
	NotifyMilestone(notice *notification.Milestone, live map[library.VideoID]*LiveVideo) error
	PostWrapped(sub *subscription.Subscription, wrapped *library.Wrapped) error
	PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, live map[library.VideoID]*LiveVideo) error
}

// LiveVideoSource は、通知に添える動画の情報を投稿の直前に取得する。ドメインの
// 判断には使わない表示のための値なので、ドメイン（library.Source）ではなく
// ここで宣言する。
type LiveVideoSource interface {
	// FetchLiveVideos は指定した動画の情報を、動画 ID をキーにして返す。見られない
	// 動画は結果に含まれない。同じ ID を 2 回指定しても 1 回だけ取得する。
	FetchLiveVideos(ids []library.VideoID) (map[library.VideoID]*LiveVideo, error)
}

// LiveVideo は、動画を通知する直前に取得して Notifier に渡す表示用の値。
// チャンネルは保存せず、再生数は保存済みのものより新しい。
type LiveVideo struct {
	ChannelName string
	// チャンネルのアイコンの URL。取得できなかったときは空
	ChannelIcon string
	Views       library.ViewCount
}
