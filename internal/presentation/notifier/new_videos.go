package notifier

import (
	"errors"
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"

	"github.com/bwmarrin/discordgo"
)

func (n *notifier) NotifyNewVideos(news *notification.NewVideos, live map[library.VideoID]*library.LiveVideo) error {
	red := color("ff0000")
	sub := news.Subscription

	var errs []error
	for _, added := range news.Videos {
		video := added.Video
		embed := &discordgo.MessageEmbed{
			Color: red,
			Author: &discordgo.MessageEmbedAuthor{
				Name: sub.Title + " に追加されました！",
			},
			Title: video.Title,
			URL:   videoURL(video.YoutubeID, sub.YoutubeID),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "追加日時",
					Value:  added.Item.AddedAt.In(n.location).Format("2006/01/02 15:04:05"),
					Inline: true,
				},
				{
					Name:   "再生回数",
					Value:  separator(views(video, live)),
					Inline: true,
				},
			},
			Image:  &discordgo.MessageEmbedImage{URL: video.Thumbnail()},
			Footer: channelFooter(video, live),
			// UTC の時刻を渡すと、Discord がユーザーに合ったタイムゾーンに変換する
			Timestamp: video.PublishedAt.Format(time.RFC3339),
		}

		// 動画 1 件につき 1 通送り、1 件失敗しても残りを送り続ける
		if err := n.send(sub.SendChannelID, embed); err != nil {
			errs = append(errs, fmt.Errorf("video %s: %w", video.YoutubeID, err))
		}
	}

	return errors.Join(errs...)
}
