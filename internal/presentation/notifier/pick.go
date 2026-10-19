package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

func pickLabel(interval subscription.PickInterval) string {
	switch interval {
	case subscription.PickWeekly:
		return "今週の一本"
	case subscription.PickMonthly:
		return "今月の一本"
	default:
		return "今日の一本"
	}
}

func (n *notifier) PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, live map[library.VideoID]*library.LiveVideo) error {
	return n.sendVideo(sub.SendChannelID, picked.Video, sub.YoutubeID, live, videoNotice{
		Color:  BLUE,
		Author: "🎬 " + pickLabel(sub.PickInterval) + " from " + sub.Title,
		Fields: []field{
			{"追加日時", picked.Item.AddedAt.In(n.location).Format("2006/01/02")},
			{"再生回数", view.Number(views(picked.Video, live))},
		},
		Image: LARGE,
	})
}
