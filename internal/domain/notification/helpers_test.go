package notification_test

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// hoursLater は前回の通知（2026-10-18 12 時 UTC）から hours 時間後の時刻を返す。
func hoursLater(hours int) time.Time {
	return time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC).Add(time.Duration(hours) * time.Hour)
}

// sub は前回の通知が hoursLater(0) の通知設定を返す。
func sub(playlistID library.PlaylistID, channelID subscription.ChannelID) *subscription.Subscription {
	s := &subscription.Subscription{YoutubeID: playlistID, SendChannelID: channelID}
	s.MarkNotified(hoursLater(0))

	return s
}

// video は詳細を取得済みの動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, status library.PrivacyStatus) *library.Video {
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: status}
}

// added はプレイリストに hoursLater(hours) に追加された動画を返す。
func added(playlistID library.PlaylistID, v *library.Video, hours int) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: playlistID,
		Item:       &library.PlaylistItem{VideoYoutubeID: v.YoutubeID, AddedAt: hoursLater(hours)},
		Video:      v,
	}
}

// snapshot は contents を中身に持つプレイリストを返す。タイトルは ID と同じにする。
func snapshot(playlistID library.PlaylistID, contents ...*library.PlaylistVideo) *library.Snapshot {
	s := &library.Snapshot{Playlist: &library.Playlist{YoutubeID: playlistID}, Title: string(playlistID), Videos: map[library.VideoID]*library.Video{}}
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

// withHidden は今回の同期で videos が見られなくなったプレイリストを返す。
func withHidden(s *library.Snapshot, videos ...*library.Video) *library.Snapshot {
	s.Hidden = videos

	return s
}
