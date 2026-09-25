package scheduler

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
)

type schedule struct {
	playlist *service.PlaylistService
	library  *service.LibraryService
	renderer *renderer
}

func NewSchedule(s *service.PlaylistService, l *service.LibraryService, r *renderer) *schedule {
	return &schedule{s, l, r}
}

func (s *schedule) Notify(location *time.Location) {
	defer recoverJob("Notify")

	playlists, err := s.playlist.FindAll()
	if err != nil {
		log.Println("Could not notify cause:", err)
		return
	}
	
	var ids []string
	for _, playlist := range playlists {
		ids = append(ids, playlist.YoutubeID)
	}
	snapshots, err := s.library.Sync(ids)
	if err != nil {
		log.Println("Could not notify cause:", err)
		return
	}

	diffs := s.playlist.GetDiffFromLatest(playlists, snapshots)
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
			log.Println("Message could not send to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "cause:", err)
		} else {
			log.Println("Successfully sent notification for playlist:", playlist.YoutubeID, "to channel:", playlist.SendChannelID)
		}
	}
}

func (s *schedule) RefreshVideos() {
	defer recoverJob("RefreshVideos")

	if err := s.library.RefreshVideos(); err != nil {
		log.Println("Could not refresh videos cause:", err)
	}
}

// A panic in one run of a job must not bring the whole bot down
func recoverJob(name string) {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic in %s: %v\n%s", name, r, debug.Stack())
	}
}
