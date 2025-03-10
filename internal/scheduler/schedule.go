package scheduler

import (
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
)

type schedule struct {
	playlist *service.PlaylistService
	renderer *renderer
}

func NewSchedule(s *service.PlaylistService, r *renderer) *schedule {
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
	var successfullyUpdated []*domain.Playlist
	
	// First update all playlists and track which ones were successful
	for _, playlist := range diffs {
		if err := s.playlist.UpdateUpdatedAt(playlist, now); err != nil {
			log.Println("Could not update playlist:", playlist.YoutubeID, "cause:", err)
		} else {
			// Only add to successful list if update succeeded
			successfullyUpdated = append(successfullyUpdated, playlist)
		}
	}
	
	// Only send notifications for playlists that were successfully updated
	for _, playlist := range successfullyUpdated {
		// Process for sending to each registered channel
		if err := s.renderer.RenderUpdatedVideo(playlist, location); err != nil {
			log.Println("Message could not send to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID)
		} else {
			log.Println("Successfully sent notification for playlist:", playlist.YoutubeID, "to channel:", playlist.SendChannelID)
		}
	}
}
