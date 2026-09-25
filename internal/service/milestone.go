package service

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

// View counts that are worth a notification
var ViewMilestones = []uint64{
	100_000,
	500_000,
	1_000_000,
	5_000_000,
	10_000_000,
	50_000_000,
	100_000_000,
}

// MilestoneFor returns the highest milestone the views have reached, or 0.
func MilestoneFor(views uint64) uint64 {
	var reached uint64
	for _, m := range ViewMilestones {
		if views >= m {
			reached = m
		}
	}

	return reached
}

// MilestoneNotice tells a channel that a video in one of its playlists
// reached a view milestone.
type MilestoneNotice struct {
	ChannelID  string
	PlaylistID string
	Video      *domain.Video
	Milestone  uint64
}

// MilestoneNotices builds the notices for the videos that reached a
// milestone, once per channel even if several of its playlists contain the video.
func MilestoneNotices(playlists []*domain.Playlist, videos []*domain.Video, containing map[string][]string) []*MilestoneNotice {
	var notices []*MilestoneNotice
	for _, video := range videos {
		inPlaylist := map[string]bool{}
		for _, id := range containing[video.YoutubeID] {
			inPlaylist[id] = true
		}

		sent := map[string]bool{}
		for _, playlist := range playlists {
			if !inPlaylist[playlist.YoutubeID] || sent[playlist.SendChannelID] {
				continue
			}
			sent[playlist.SendChannelID] = true
			notices = append(notices, &MilestoneNotice{
				ChannelID:  playlist.SendChannelID,
				PlaylistID: playlist.YoutubeID,
				Video:      video,
				Milestone:  MilestoneFor(video.Views),
			})
		}
	}

	return notices
}
