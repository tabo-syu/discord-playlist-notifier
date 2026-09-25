package service

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

// HiddenVideoNotice tells a channel that a video in one of its playlists
// became private or was deleted.
type HiddenVideoNotice struct {
	ChannelID     string
	PlaylistID    string
	PlaylistTitle string
	Video         *domain.Video
}

// HiddenVideoNotices builds the notices for the videos that became hidden in
// this sync. A video is stored once even if it is in several playlists, so
// the change is only reported by the first synced playlist; every registered
// playlist containing the video is notified, once per channel.
func HiddenVideoNotices(playlists []*domain.Playlist, snapshots map[string]*PlaylistSnapshot) []*HiddenVideoNotice {
	hidden := map[string]*domain.Video{}
	var order []string
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

	var notices []*HiddenVideoNotice
	sent := map[[2]string]bool{}
	for _, videoId := range order {
		for _, playlist := range playlists {
			snapshot, ok := snapshots[playlist.YoutubeID]
			if !ok || !snapshot.contains(videoId) {
				continue
			}
			key := [2]string{playlist.SendChannelID, videoId}
			if sent[key] {
				continue
			}
			sent[key] = true
			notices = append(notices, &HiddenVideoNotice{
				ChannelID:     playlist.SendChannelID,
				PlaylistID:    playlist.YoutubeID,
				PlaylistTitle: snapshot.Title,
				Video:         hidden[videoId],
			})
		}
	}

	return notices
}

func (s *PlaylistSnapshot) contains(videoId string) bool {
	for _, item := range s.Items {
		if item.VideoYoutubeID == videoId {
			return true
		}
	}

	return false
}
