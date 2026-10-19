package subscription_test

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// hoursLater は前回の通知（2026-10-18 12 時 UTC）から hours 時間後の時刻を返す。
func hoursLater(hours int) time.Time {
	return time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC).Add(time.Duration(hours) * time.Hour)
}

// sub はプレイリスト PL をチャンネル ch に通知する、タイトルが title で前回の通知が
// notifiedAt の通知設定を返す。
func sub(title string, notifiedAt time.Time) *subscription.Subscription {
	return &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: title, NotifiedUntil: notifiedAt}
}

// video は詳細を取得済みの動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, status library.PrivacyStatus) *library.Video {
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: status}
}

// added はプレイリスト PL に hoursLater(hours) に追加された動画を返す。
func added(v *library.Video, hours int) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: "PL",
		Item:       &library.PlaylistItem{VideoYoutubeID: v.YoutubeID, AddedAt: hoursLater(hours)},
		Video:      v,
	}
}

// snapshot は contents を中身に持つ、YouTube 上のタイトルが「新しいタイトル」の
// プレイリスト PL を返す。
func snapshot(contents ...*library.PlaylistVideo) *library.Snapshot {
	s := &library.Snapshot{Playlist: &library.Playlist{YoutubeID: "PL"}, Title: "新しいタイトル", Videos: map[library.VideoID]*library.Video{}}
	for _, c := range contents {
		s.Items = append(s.Items, c.Item)
		s.Videos[c.Video.YoutubeID] = c.Video
	}

	return s
}

// deleted は YouTube 上で削除されたプレイリストを返す。
func deleted(s *library.Snapshot) *library.Snapshot {
	s.Deleted = true

	return s
}
