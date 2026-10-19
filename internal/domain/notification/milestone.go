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
func FindMilestones(subscriptions []*subscription.Subscription, events []*library.MilestoneReached, containing map[library.VideoID][]library.PlaylistID) []*Milestone {
	var notices []*Milestone

	sent := map[sentKey]bool{}

	for _, e := range events {
		for _, sub := range recipients(subscriptions, e.Video.YoutubeID, containing, sent) {
			notices = append(notices, &Milestone{
				ChannelID:  sub.SendChannelID,
				PlaylistID: sub.YoutubeID,
				Video:      e.Video,
				Milestone:  e.Milestone,
			})
		}
	}

	return notices
}
