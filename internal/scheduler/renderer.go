package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"github.com/bwmarrin/discordgo"
)

type renderer struct {
	session *discordgo.Session
}

func NewRenderer(s *discordgo.Session) *renderer {
	return &renderer{s}
}

func (r *renderer) RenderUpdatedVideo(playlist *domain.Playlist, location *time.Location) error {
	red := color("ff0000")

	var embeds []*discordgo.MessageEmbed
	for _, video := range playlist.Videos {
		embed := &discordgo.MessageEmbed{
			Color: red,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    playlist.Title + " に " + video.AddedByChannelName + " さんが追加しました！",
				IconURL: video.AddedByChannelIcon,
			},
			Title: video.Title,
			URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", video.YoutubeID, playlist.YoutubeID),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "追加日時",
					Value:  video.PublishedAt.In(location).Format("2006/01/02 15:04:05"),
					Inline: true,
				},
				{
					Name:   "再生回数",
					Value:  separator(video.Views),
					Inline: true,
				},
			},
			Image: &discordgo.MessageEmbedImage{URL: video.Thumbnail},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    video.ChannelName,
				IconURL: video.ChannelIcon,
			},
			// Passing UTC time allows Discord to convert to the user's appropriate timezone
			Timestamp: video.OwnerPublishedAt.Format(time.RFC3339),
		}

		embeds = append(embeds, embed)
	}

	_, err := r.session.ChannelMessageSendEmbeds(playlist.SendChannelID, embeds)
	if err != nil {
		return err
	}

	return nil
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
