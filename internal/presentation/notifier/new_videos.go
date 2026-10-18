package notifier

import (
	"errors"
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

func (n *notifier) NotifyNewVideos(news *notification.NewVideos, live map[library.VideoID]*library.LiveVideo) error {
	sub := news.Subscription

	var errs []error
	for _, added := range news.Videos {
		video := added.Video
		// 動画 1 件につき 1 通送り、1 件失敗しても残りを送り続ける
		err := n.sendVideo(sub.SendChannelID, video, sub.YoutubeID, live, videoNotice{
			Color:  RED,
			Author: sub.Title + " に追加されました！",
			Fields: []field{
				{"追加日時", added.Item.AddedAt.In(n.location).Format("2006/01/02 15:04:05")},
				{"再生回数", view.Number(views(video, live))},
			},
			Image:     LARGE,
			Timestamp: video.PublishedAt,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("video %s: %w", video.YoutubeID, err))
		}
	}

	return errors.Join(errs...)
}
