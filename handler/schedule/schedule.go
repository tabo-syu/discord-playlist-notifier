package schedule

import (
	"discord-playlist-notifier/service"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Schedule interface {
	Notify(*time.Location)
}

type schedule struct {
	session  *discordgo.Session
	playlist service.PlaylistService
}

func NewSchedule(s *discordgo.Session, y service.PlaylistService) Schedule {
	return &schedule{s, y}
}

func (s *schedule) Notify(location *time.Location) {
	playlists, err := s.playlist.FindAll()
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}
	diffs, err := s.playlist.GetDiffFromLatest(playlists)
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}
	if len(diffs) == 0 {
		log.Println("Playlist was not updated")

		return
	}

	now := time.Now()
	for _, playlist := range diffs {
		if err := s.playlist.UpdateUpdatedAt(playlist, now); err != nil {
			log.Println("Could not update cause:", err)
		}
	}

	for _, playlist := range diffs {
		message := "動画が" + playlist.Title + "に追加されました！\n"
		for _, video := range playlist.Videos {
			message += fmt.Sprintf(
				"https://www.youtube.com/watch?v=%s&list=%s (%s)\n",
				// 世界標準時のため、9時間プラスして JST に合わせる
				video.YoutubeID, playlist.YoutubeID, video.PublishedAt.In(location).Format("2006-01-02 15:04:05"),
			)
		}
		// 登録済みの各チャンネルへの送信処理
		_, err := s.session.ChannelMessageSend(playlist.SendChannelID, message)
		if err != nil {
			log.Println("Message could not send to", playlist.SendChannelID)
		}
	}
}
