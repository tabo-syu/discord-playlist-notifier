package scheduler

import (
	"log"
	"runtime/debug"
	"time"

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
	// A panic in one run must not bring the whole bot down
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Notify: %v\n%s", r, debug.Stack())
		}
	}()

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
	var successfullyUpdated []*service.NewVideos
	
	// First update all playlists and track which ones were successful
	for _, diff := range diffs {
		if err := s.playlist.UpdateUpdatedAt(diff.Playlist, now); err != nil {
			log.Println("Could not update playlist:", diff.Playlist.YoutubeID, "cause:", err)
		} else {
			// Only add to successful list if update succeeded
			successfullyUpdated = append(successfullyUpdated, diff)
		}
	}
	
	// Only send notifications for playlists that were successfully updated
	for _, diff := range successfullyUpdated {
		playlist := diff.Playlist
		// The channel is not stored, so fetch it now. Send without it on failure.
		var videoIds []string
		for _, v := range diff.Videos {
			videoIds = append(videoIds, v.Video.YoutubeID)
		}
		live, err := s.library.LiveVideos(videoIds)
		if err != nil {
			log.Println("Could not fetch channels for playlist:", playlist.YoutubeID, "cause:", err)
		}

		// Process for sending to each registered channel
		if err := s.renderer.RenderUpdatedVideo(diff, live, location); err != nil {
			log.Println("Message could not send to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "cause:", err)
		} else {
			log.Println("Successfully sent notification for playlist:", playlist.YoutubeID, "to channel:", playlist.SendChannelID)
		}
	}
}

func (s *schedule) RefreshVideos() {
	// A panic in one run must not bring the whole bot down
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in RefreshVideos: %v\n%s", r, debug.Stack())
		}
	}()

	reached, err := s.library.RefreshVideos()
	if err != nil {
		log.Println("Could not refresh videos cause:", err)
		return
	}
	if len(reached) == 0 {
		return
	}

	playlists, err := s.playlist.FindAll()
	if err != nil {
		log.Println("Could not notify milestones cause:", err)
		return
	}
	var videoIds []string
	for _, v := range reached {
		videoIds = append(videoIds, v.YoutubeID)
	}
	containing, err := s.library.PlaylistsContaining(videoIds)
	if err != nil {
		log.Println("Could not notify milestones cause:", err)
		return
	}

	// The channel is not stored, so fetch it now. Send without it on failure.
	live, err := s.library.LiveVideos(videoIds)
	if err != nil {
		log.Println("Could not fetch channels for milestones cause:", err)
	}

	for _, notice := range service.MilestoneNotices(playlists, reached, containing) {
		if err := s.renderer.RenderMilestone(notice, live); err != nil {
			log.Println("Milestone notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent milestone notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "milestone:", notice.Milestone)
		}
	}
}
