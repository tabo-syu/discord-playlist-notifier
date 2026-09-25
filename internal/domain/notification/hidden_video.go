package notification

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// HiddenVideo tells a channel that a video in one of its playlists became
// private or was deleted.
type HiddenVideo struct {
	ChannelID     subscription.ChannelID
	PlaylistID    library.PlaylistID
	PlaylistTitle string
	Video         *library.Video
}

// FindHiddenVideos builds the notices for the videos that became hidden in
// this sync. A video is stored once even if it is in several playlists, so
// the change is only reported by the first synced playlist; every subscribed
// playlist containing the video is notified, once per channel.
func FindHiddenVideos(subscriptions []*subscription.Subscription, snapshots map[library.PlaylistID]*library.Snapshot) []*HiddenVideo {
	hidden := map[library.VideoID]*library.Video{}
	var order []library.VideoID
	for _, snapshot := range snapshots {
		for _, v := range snapshot.Hidden {
			if _, ok := hidden[v.YoutubeID]; !ok {
				order = append(order, v.YoutubeID)
			}
			hidden[v.YoutubeID] = v
		}
	}
	if len(hidden) == 0 {
		return nil
	}

	type key struct {
		channel subscription.ChannelID
		video   library.VideoID
	}
	var notices []*HiddenVideo
	sent := map[key]bool{}
	for _, videoID := range order {
		for _, sub := range subscriptions {
			snapshot, ok := snapshots[sub.YoutubeID]
			if !ok || !snapshot.Contains(videoID) {
				continue
			}
			k := key{sub.SendChannelID, videoID}
			if sent[k] {
				continue
			}
			sent[k] = true
			notices = append(notices, &HiddenVideo{
				ChannelID:     sub.SendChannelID,
				PlaylistID:    sub.YoutubeID,
				PlaylistTitle: snapshot.Title,
				Video:         hidden[videoID],
			})
		}
	}

	return notices
}
