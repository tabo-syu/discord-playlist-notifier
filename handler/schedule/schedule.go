package schedule

import (
	"discord-playlist-notifier/renderer"
	"discord-playlist-notifier/service"
	"log"
	"time"
)

type Schedule interface {
	Notify(*time.Location)
}

type schedule struct {
	playlist service.PlaylistService
	renderer renderer.Renderer
}

func NewSchedule(s service.PlaylistService, r renderer.Renderer) Schedule {
	return &schedule{s, r}
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
		// 登録済みの各チャンネルへの送信処理
		if err := s.renderer.RenderUpdatedVideo(playlist, location); err != nil {
			log.Println("Message could not send to", playlist.SendChannelID)
		}
	}
}
