// Package notification は、動画に起きたこと（ドメインイベント）を、どのチャンネルに
// 通知するかに振り分ける。動画は全サーバーで共有するので、複数の通知設定をまたいで
// 判断する。
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

// FindHiddenVideos は見られなくなった動画の通知を、チャンネルの複数のプレイリストが
// その動画を含んでいてもチャンネルごとに 1 回ずつ組み立てる。プレイリストの
// タイトルは通知設定に保存したもの。containing は動画ごとに、それを含む
// プレイリストを対応付ける。
func FindHiddenVideos(subscriptions []*subscription.Subscription, events []*library.VideoHidden, containing map[library.VideoID][]library.PlaylistID) []*HiddenVideo {
	var notices []*HiddenVideo

	sent := map[sentKey]bool{}

	for _, e := range events {
		for _, sub := range recipients(subscriptions, e.Video.YoutubeID, containing, sent) {
			notices = append(notices, &HiddenVideo{
				ChannelID:     sub.SendChannelID,
				PlaylistID:    sub.YoutubeID,
				PlaylistTitle: sub.Title,
				Video:         e.Video,
			})
		}
	}

	return notices
}

// sentKey は、チャンネルにその動画を通知したかを覚えておくためのキー。
type sentKey struct {
	channel subscription.ChannelID
	video   library.VideoID
}

// recipients は動画を含むプレイリストの通知設定を、チャンネルごとに 1 つずつ返す。
// sent に記録したチャンネルと動画の組は除き、返した組を sent に記録する。
// 同じ動画のイベントが 2 回あっても、同じチャンネルには 1 回だけ通知する。
func recipients(subscriptions []*subscription.Subscription, videoID library.VideoID, containing map[library.VideoID][]library.PlaylistID, sent map[sentKey]bool) []*subscription.Subscription {
	inPlaylist := map[library.PlaylistID]bool{}
	for _, id := range containing[videoID] {
		inPlaylist[id] = true
	}

	var result []*subscription.Subscription

	for _, sub := range subscriptions {
		k := sentKey{sub.SendChannelID, videoID}
		if !inPlaylist[sub.YoutubeID] || sent[k] {
			continue
		}

		sent[k] = true

		result = append(result, sub)
	}

	return result
}
