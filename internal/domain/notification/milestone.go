package notification

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// Milestone は、チャンネルのプレイリストにある動画が再生数の節目に達したことを
// 伝える。
type Milestone struct {
	ChannelID  subscription.ChannelID
	PlaylistID library.PlaylistID
	Video      *library.Video
	Milestone  library.ViewCount
}

// FindMilestones は節目に達した動画の通知を、チャンネルの複数のプレイリストが
// その動画を含んでいてもチャンネルごとに 1 回ずつ組み立てる。
// containing は動画ごとに、それを含むプレイリストを対応付ける。
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
