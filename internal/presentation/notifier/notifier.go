// Package notifier implements application.Notifier by posting embeds to
// Discord channels.
package notifier

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"github.com/bwmarrin/discordgo"
)

type notifier struct {
	session *discordgo.Session
	// Time zone used to show dates
	location *time.Location
}

func New(s *discordgo.Session, loc *time.Location) *notifier {
	return &notifier{s, loc}
}

func (n *notifier) send(channelID subscription.ChannelID, embed *discordgo.MessageEmbed) error {
	_, err := n.session.ChannelMessageSendEmbed(string(channelID), embed)

	return err
}

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
			// Passing UTC time allows Discord to convert to the user's appropriate timezone
			Timestamp: video.PublishedAt.Format(time.RFC3339),
		}

		// Send one message per video, and keep sending the rest even if one fails
		if err := n.send(sub.SendChannelID, embed); err != nil {
			errs = append(errs, fmt.Errorf("video %s: %w", video.YoutubeID, err))
		}
	}

	return errors.Join(errs...)
}

func videoURL(videoID library.VideoID, playlistID library.PlaylistID) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", videoID, playlistID)
}

// views prefers the views fetched right before sending.
func views(video *library.Video, live map[library.VideoID]*library.LiveVideo) library.ViewCount {
	if l, ok := live[video.YoutubeID]; ok {
		return l.Views
	}
	return video.Views
}

// channelFooter shows the channel of the video, which is only known when it
// was fetched right before sending.
func channelFooter(video *library.Video, live map[library.VideoID]*library.LiveVideo) *discordgo.MessageEmbedFooter {
	l, ok := live[video.YoutubeID]
	if !ok {
		return nil
	}
	return &discordgo.MessageEmbedFooter{Text: l.ChannelName, IconURL: l.ChannelIcon}
}

func color(hex string) int {
	color, err := strconv.ParseInt(hex, 16, 0)
	if err != nil {
		// Default to red if there's an error
		return 0xff0000
	}

	return int(color)
}

func separator(integer library.ViewCount) string {
	// Use a more efficient approach with strings.Builder
	var sb strings.Builder
	str := fmt.Sprintf("%d", integer)

	// Calculate the number of commas needed
	commas := (len(str) - 1) / 3

	// Pre-allocate the buffer to avoid reallocations
	sb.Grow(len(str) + commas)

	// Add digits with commas
	for i, char := range str {
		// Add a comma before every 3rd digit from the right, except at the beginning
		if i > 0 && (len(str)-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(char)
	}

	return sb.String()
}
