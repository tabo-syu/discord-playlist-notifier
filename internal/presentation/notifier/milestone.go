package notifier

import (
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"

	"github.com/bwmarrin/discordgo"
)

func (n *notifier) NotifyMilestone(notice *notification.Milestone, live map[library.VideoID]*library.LiveVideo) error {
	embed := &discordgo.MessageEmbed{
		Color: color("ffd700"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: "🎉 " + japaneseCount(notice.Milestone) + "回再生を突破しました！",
		},
		Title: notice.Video.Title,
		URL:   videoURL(notice.Video.YoutubeID, notice.PlaylistID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "再生回数",
				Value:  separator(views(notice.Video, live)),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: notice.Video.Thumbnail()},
		Footer:    channelFooter(notice.Video, live),
	}

	return n.send(notice.ChannelID, embed)
}

// japaneseCount formats round numbers with Japanese units, e.g. 1000万, 1億.
func japaneseCount(n library.ViewCount) string {
	switch {
	case n >= 100_000_000 && n%100_000_000 == 0:
		return fmt.Sprintf("%d億", n/100_000_000)
	case n >= 10_000 && n%10_000 == 0:
		return fmt.Sprintf("%d万", n/10_000)
	default:
		return separator(n)
	}
}
