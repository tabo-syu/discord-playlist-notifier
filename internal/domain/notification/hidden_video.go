package notification

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// HiddenVideo は、チャンネルのプレイリストにある動画が非公開・削除になったことを
// 伝える。
type HiddenVideo struct {
	ChannelID     subscription.ChannelID
	PlaylistID    library.PlaylistID
	PlaylistTitle string
	Video         *library.Video
}

// FindHiddenVideos は今回の同期で見られなくなった動画の通知を組み立てる。
// 動画は複数のプレイリストにあっても 1 つしか保存しないので、変化は最初に
// 同期したプレイリストでしか報告されない。その動画を含む通知登録済みの
// 全プレイリストに、チャンネルごとに 1 回ずつ通知する。
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
