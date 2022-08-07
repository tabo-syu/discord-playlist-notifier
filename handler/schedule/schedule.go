package schedule

import (
	"discord-playlist-notifier/service"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Schedule interface {
	Notify()
}

type schedule struct {
	session  *discordgo.Session
	playlist service.PlaylistService
}

func NewSchedule(s *discordgo.Session, y service.PlaylistService) Schedule {
	return &schedule{s, y}
}

func (s *schedule) Notify() {
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

	now := time.Now()
	for _, playlist := range playlists {
		if err := s.playlist.UpdateUpdatedAt(playlist, now); err != nil {
			log.Println("Could not update cause:", err)
		}
	}

	for _, playlist := range diffs {
		message := "動画がプレイリストに追加されました！\n"
		for _, video := range playlist.Videos {
			message += fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s\n", video.YoutubeID, playlist.YoutubeID)
		}

		// 登録済みの各チャンネルへの送信処理
		_, err := s.session.ChannelMessageSend(playlist.SendChannelID, message)
		if err != nil {
			log.Println("Message could not send to", playlist.SendChannelID)
		}
	}
}
