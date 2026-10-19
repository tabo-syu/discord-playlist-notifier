package notification_test

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// sub はプレイリストをチャンネルに通知する通知設定を返す。タイトルはプレイリストの
// ID と同じにする。
func sub(playlistID library.PlaylistID, channelID subscription.ChannelID) *subscription.Subscription {
	return &subscription.Subscription{YoutubeID: playlistID, SendChannelID: channelID, Title: string(playlistID)}
}

// video は詳細を取得済みの動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, status library.PrivacyStatus) *library.Video {
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: status}
}
