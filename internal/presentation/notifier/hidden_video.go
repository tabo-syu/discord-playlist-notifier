package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
)

func (n *notifier) NotifyHiddenVideo(notice *notification.HiddenVideo) error {
	status := "非公開になりました"
	if notice.Video.PrivacyStatus == library.PrivacyDeleted {
		status = "削除されました"
	}

	// チャンネルは保存しておらず、もう取得もできないので、投稿者は出さない（live は nil）
	return n.sendVideo(notice.ChannelID, notice.Video, notice.PlaylistID, nil, videoNotice{
		Color:  GRAY,
		Author: notice.PlaylistTitle + " の動画が" + status,
		Image:  SMALL,
	})
}
