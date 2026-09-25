package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

type renderer struct {
	session *discordgo.Session
}

func NewRenderer(s *discordgo.Session) *renderer {
	return &renderer{s}
}

func (r *renderer) RenderUpdatedVideo(diff *service.NewVideos, live map[string]*domain.LiveVideo, location *time.Location) error {
	red := color("ff0000")
	playlist := diff.Playlist

	var errs []error
	for _, added := range diff.Videos {
		video := added.Video
		embed := &discordgo.MessageEmbed{
			Color: red,
			Author: &discordgo.MessageEmbedAuthor{
				Name: playlist.Title + " に追加されました！",
			},
			Title: video.Title,
			URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", video.YoutubeID, playlist.YoutubeID),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "追加日時",
					Value:  added.Item.AddedAt.In(location).Format("2006/01/02 15:04:05"),
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
		if _, err := r.session.ChannelMessageSendEmbed(playlist.SendChannelID, embed); err != nil {
			errs = append(errs, fmt.Errorf("video %s: %w", video.YoutubeID, err))
		}
	}

	return errors.Join(errs...)
}

// views prefers the views fetched right before sending.
func views(video *domain.Video, live map[string]*domain.LiveVideo) uint64 {
	if l, ok := live[video.YoutubeID]; ok {
		return l.Views
	}
	return video.Views
}

// channelFooter shows the channel of the video, which is only known when it
// was fetched right before sending.
func channelFooter(video *domain.Video, live map[string]*domain.LiveVideo) *discordgo.MessageEmbedFooter {
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

func separator(integer uint64) string {
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
