package notification

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Milestone tells a channel that a video in one of its playlists reached a
// view milestone.
type Milestone struct {
	ChannelID  subscription.ChannelID
	PlaylistID library.PlaylistID
	Video      *library.Video
	Milestone  library.ViewCount
}

// FindMilestones builds the notices for the videos that reached a milestone,
// once per channel even if several of its playlists contain the video.
// containing maps each video to the playlists that contain it.
func FindMilestones(subscriptions []*subscription.Subscription, videos []*library.Video, containing map[library.VideoID][]library.PlaylistID) []*Milestone {
	var notices []*Milestone
	for _, video := range videos {
		inPlaylist := map[library.PlaylistID]bool{}
		for _, id := range containing[video.YoutubeID] {
			inPlaylist[id] = true
		}

		sent := map[subscription.ChannelID]bool{}
		for _, sub := range subscriptions {
			if !inPlaylist[sub.YoutubeID] || sent[sub.SendChannelID] {
				continue
			}
			sent[sub.SendChannelID] = true
			notices = append(notices, &Milestone{
				ChannelID:  sub.SendChannelID,
				PlaylistID: sub.YoutubeID,
				Video:      video,
				Milestone:  video.Views.Milestone(),
			})
		}
	}

	return notices
}
