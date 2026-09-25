package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"github.com/bwmarrin/discordgo"
)

func pickLabel(interval subscription.PickInterval) string {
	switch interval {
	case subscription.PickWeekly:
		return "今週の一曲"
	case subscription.PickMonthly:
		return "今月の一曲"
	default:
		return "今日の一曲"
	}
}

func (n *notifier) PostPick(sub *subscription.Subscription, picked *library.PlaylistVideo, live map[library.VideoID]*library.LiveVideo) error {
	video := picked.Video
	embed := &discordgo.MessageEmbed{
		Color: color("1e90ff"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: "🎵 " + pickLabel(sub.PickInterval) + " from " + sub.Title,
		},
		Title: video.Title,
		URL:   videoURL(video.YoutubeID, sub.YoutubeID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "追加日時",
				Value:  picked.Item.AddedAt.In(n.location).Format("2006/01/02"),
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
	}

	return n.send(sub.SendChannelID, embed)
}
