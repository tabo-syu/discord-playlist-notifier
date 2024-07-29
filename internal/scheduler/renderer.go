package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"github.com/bwmarrin/discordgo"
)

type Renderer interface {
	RenderUpdatedVideo(playlist *domain.Playlist, location *time.Location) error
}

type renderer struct {
	session *discordgo.Session
}

func NewRenderer(s *discordgo.Session) Renderer {
	return &renderer{s}
}

func (r *renderer) RenderUpdatedVideo(playlist *domain.Playlist, location *time.Location) error {
	red := color("ff0000")

	var embeds []*discordgo.MessageEmbed
	for _, video := range playlist.Videos {
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
			// 世界標準時を渡すことでユーザーの適切なタイムゾーンに変換してくれる
			Timestamp: video.OwnerPublishedAt.Format("2006-01-02 15:04:05"),
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
	color, _ := strconv.ParseInt(hex, 16, 0)

	return int(color)
}

func separator(integer uint64) string {
	arr := strings.Split(fmt.Sprintf("%d", integer), "")
	var (
		str string
		i2  int
	)
	for i := len(arr) - 1; i >= 0; i-- {
		if i2 > 2 && i2%3 == 0 {
			str = fmt.Sprintf(",%s", str)
		}
		str = fmt.Sprintf("%s%s", arr[i], str)
		i2++
	}

	return str
}
