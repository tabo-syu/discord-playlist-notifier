package scheduler

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Local time (in the scheduler location) to post the daily pick
const DAILY_PICK_TIME = "12:00"

func (s *schedule) DailyPick(location *time.Location) {
	// A panic in one run must not bring the whole bot down
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in DailyPick: %v\n%s", r, debug.Stack())
		}
	}()

	playlists, err := s.playlist.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return
	}
	if err != nil {
		log.Println("Could not post daily pick cause:", err)
		return
	}

	for _, playlist := range playlists {
		if !playlist.DailyPick {
			continue
		}

		picked, err := s.library.RandomVideos([]string{playlist.YoutubeID}, 1)
		if err != nil {
			log.Println("Could not pick a video for playlist:", playlist.YoutubeID, "cause:", err)
			continue
		}
		if len(picked) == 0 {
			log.Println("No video to pick for playlist:", playlist.YoutubeID)
			continue
		}

		// The channel is not stored, so fetch it now. Send without it on failure.
		live, err := s.library.LiveVideos([]string{picked[0].Video.YoutubeID})
		if err != nil {
			log.Println("Could not fetch the channel for daily pick cause:", err)
		}

		if err := s.renderer.RenderDailyPick(playlist, picked[0], live, location); err != nil {
			log.Println("Daily pick could not send to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent daily pick to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "video:", picked[0].Video.YoutubeID)
		}
	}
}

func (r *renderer) RenderDailyPick(playlist *domain.Playlist, picked *service.PlaylistVideo, live map[string]*domain.LiveVideo, location *time.Location) error {
	video := picked.Video
	embed := &discordgo.MessageEmbed{
		Color: color("1e90ff"),
		Author: &discordgo.MessageEmbedAuthor{
			Name: "🎵 今日の一曲 from " + playlist.Title,
		},
		Title: video.Title,
		URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", video.YoutubeID, playlist.YoutubeID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "追加日時",
				Value:  picked.Item.AddedAt.In(location).Format("2006/01/02"),
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

	_, err := r.session.ChannelMessageSendEmbed(playlist.SendChannelID, embed)

	return err
}
