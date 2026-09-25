package scheduler

import (
	"errors"
	"log"
	"runtime/debug"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
	"github.com/tabo-syu/discord-playlist-notifier/internal/view"

	"github.com/bwmarrin/discordgo"
)

// Cron expression (in the scheduler location) to post the yearly summary: Dec 31 21:00
const WRAPPED_CRON = "0 21 31 12 *"

// Discord rejects embed descriptions longer than this
const MAX_EMBED_DESCRIPTION_LENGTH = 4096

func (s *schedule) Wrapped(location *time.Location) {
	// A panic in one run must not bring the whole bot down
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Wrapped: %v\n%s", r, debug.Stack())
		}
	}()

	playlists, err := s.playlist.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return
	}
	if err != nil {
		log.Println("Could not post wrapped cause:", err)
		return
	}

	year := time.Now().In(location).Year()
	for _, playlist := range playlists {
		contents, err := s.library.Contents([]string{playlist.YoutubeID})
		if err != nil {
			log.Println("Could not load playlist:", playlist.YoutubeID, "cause:", err)
			continue
		}
		w := service.ComputeWrapped(contents, year, location)
		if w.Added == 0 {
			continue
		}

		if err := s.renderer.RenderWrapped(playlist, w, location); err != nil {
			log.Println("Wrapped could not send to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent wrapped to", playlist.SendChannelID, "for playlist:", playlist.YoutubeID, "year:", year)
		}
	}
}

func (r *renderer) RenderWrapped(playlist *domain.Playlist, w *service.Wrapped, location *time.Location) error {
	embed := &discordgo.MessageEmbed{
		Color:       color("9b59b6"),
		Title:       view.WrappedTitle(playlist.Title, w.Year),
		URL:         "https://www.youtube.com/playlist?list=" + playlist.YoutubeID,
		Description: view.Truncate(view.Wrapped(w, location, view.InEmbed), MAX_EMBED_DESCRIPTION_LENGTH),
	}

	_, err := r.session.ChannelMessageSendEmbed(playlist.SendChannelID, embed)

	return err
}
