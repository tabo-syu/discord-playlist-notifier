package notifier_test

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// errMock は mock に失敗させるときに返すエラー。
var errMock = errors.New("mock error")

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// sub はチャンネル ch にプレイリスト PL を通知する設定を返す。
func sub(title string, interval subscription.PickInterval) *subscription.Subscription {
	return &subscription.Subscription{Model: gorm.Model{ID: 1}, YoutubeID: "PL", SendChannelID: "ch", Title: title, PickInterval: interval, GuildID: 1}
}

// video は公開の動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: library.PrivacyPublic, Views: views}
}

// added はプレイリスト PL に addedAt に追加された動画を返す。
func added(v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: "PL",
		Item:       &library.PlaylistItem{PlaylistYoutubeID: "PL", VideoYoutubeID: v.YoutubeID, AddedAt: addedAt},
		Video:      v,
	}
}
