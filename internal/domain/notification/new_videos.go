// Package notification は通知設定とそのプレイリストの中身から、どのチャンネルに
// 何を通知するかを決める。
package notification

import (
	"sort"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NewVideos は通知登録されたプレイリストに、前回の通知以降に追加された動画。
type NewVideos struct {
	Subscription *subscription.Subscription
	// YouTube 上のプレイリストの現在のタイトル
	PlaylistTitle string
	// 追加された時刻の順
	Videos []*library.PlaylistVideo
}

// FindNewVideos は通知設定ごとに、前回の通知以降にプレイリストへ追加された
// 視聴可能な動画を選ぶ。プレイリストの同期に失敗したか、プレイリストが
// 削除された通知設定は除く。
func FindNewVideos(subscriptions []*subscription.Subscription, snapshots map[library.PlaylistID]*library.Snapshot) []*NewVideos {
	var updates []*NewVideos

	for _, sub := range subscriptions {
		latest, ok := snapshots[sub.YoutubeID]
		if !ok || latest.Deleted {
			continue
		}

		var added []*library.PlaylistVideo

		for _, item := range latest.Items {
			if !sub.IsNew(item) {
				continue
			}

			video, ok := latest.Videos[item.VideoYoutubeID]
			if !ok || !video.Available() {
				continue
			}

			added = append(added, &library.PlaylistVideo{PlaylistID: sub.YoutubeID, Item: item, Video: video})
		}

		if len(added) != 0 {
			// プレイリストの順に関係なく、動画が追加された順に通知する
			sort.SliceStable(added, func(i, j int) bool {
				return added[i].Item.AddedAt.Before(added[j].Item.AddedAt)
			})
			updates = append(updates, &NewVideos{Subscription: sub, PlaylistTitle: latest.Title, Videos: added})
		}
	}

	return updates
}
