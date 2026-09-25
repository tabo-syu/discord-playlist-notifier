package scheduler

import (
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

func (r *renderer) RenderMilestone(notice *service.MilestoneNotice, live map[string]*domain.LiveVideo) error {
	embed := &discordgo.MessageEmbed{
		Color: color("ffd700"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: "🎉 " + japaneseCount(notice.Milestone) + "回再生を突破しました！",
		},
		Title: notice.Video.Title,
		URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", notice.Video.YoutubeID, notice.PlaylistID),
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

	_, err := r.session.ChannelMessageSendEmbed(notice.ChannelID, embed)

	return err
}

// japaneseCount formats round numbers with Japanese units, e.g. 1000万, 1億.
func japaneseCount(n uint64) string {
	switch {
	case n >= 100_000_000 && n%100_000_000 == 0:
		return fmt.Sprintf("%d億", n/100_000_000)
	case n >= 10_000 && n%10_000 == 0:
		return fmt.Sprintf("%d万", n/10_000)
	default:
		return separator(n)
	}
}
