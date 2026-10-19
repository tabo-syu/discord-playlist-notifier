package subscription

import "github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

// VideosAdded は、前回の通知の後に、通知設定のプレイリストへ視聴できる動画が
// 追加されたことを表すドメインイベント。Subscription.TakeNewVideos が基準時刻を
// 進めたときに返し、アプリケーションサービスが通知設定を保存できた後にだけ配る。
//
// 起きた時点の事実なので、後で変わる Subscription そのものではなく、送るのに要る
// 値を持つ。
type VideosAdded struct {
	ChannelID  ChannelID
	PlaylistID library.PlaylistID
	// YouTube 上のプレイリストの、検知した時点のタイトル
	PlaylistTitle string
	// 追加された時刻の順
	Videos []*library.PlaylistVideo
}
