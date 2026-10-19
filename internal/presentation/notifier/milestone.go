package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

func (n *notifier) NotifyMilestone(notice *notification.Milestone, live map[library.VideoID]*application.LiveVideo) error {
	return n.sendVideo(notice.ChannelID, notice.Video, notice.PlaylistID, live, videoNotice{
		Color:  GOLD,
		Author: "🎉 " + view.JapaneseCount(notice.Milestone) + "回再生を突破しました！",
		Fields: []field{
			{"再生回数", view.Number(views(notice.Video, live))},
		},
		Image: SMALL,
	})
}
