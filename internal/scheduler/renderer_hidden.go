package scheduler

import (
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

func (r *renderer) RenderHiddenVideo(notice *service.HiddenVideoNotice) error {
	status := "非公開になりました"
	if notice.Video.PrivacyStatus == domain.PrivacyDeleted {
		status = "削除されました"
	}

	embed := &discordgo.MessageEmbed{
		Color: color("808080"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: notice.PlaylistTitle + " の動画が" + status,
		},
		Title:     notice.Video.Title,
		URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", notice.Video.YoutubeID, notice.PlaylistID),
		// The channel is not stored and can no longer be fetched, so there is no footer
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: notice.Video.Thumbnail()},
	}

	_, err := r.session.ChannelMessageSendEmbed(notice.ChannelID, embed)

	return err
}
