package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"

	"github.com/bwmarrin/discordgo"
)

func (n *notifier) NotifyHiddenVideo(notice *notification.HiddenVideo) error {
	status := "非公開になりました"
	if notice.Video.PrivacyStatus == library.PrivacyDeleted {
		status = "削除されました"
	}

	embed := &discordgo.MessageEmbed{
		Color: color("808080"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: notice.PlaylistTitle + " の動画が" + status,
		},
		Title: notice.Video.Title,
		URL:   videoURL(notice.Video.YoutubeID, notice.PlaylistID),
		// The channel is not stored and can no longer be fetched, so there is no footer
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: notice.Video.Thumbnail()},
	}

	return n.send(notice.ChannelID, embed)
}
