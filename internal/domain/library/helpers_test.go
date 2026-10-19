package library_test

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// date は JST の日時を返す。
func date(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, jst)
}

// video は詳細を取得済みの動画を返す。タイトルは ID と同じにする。
func video(id string, status library.PrivacyStatus, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: library.VideoID(id), DetailsFetched: true, Title: id, PrivacyStatus: status, Views: views}
}

// added はプレイリストに addedAt に追加された動画を返す。
func added(v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: "PL",
		Item:       &library.PlaylistItem{VideoYoutubeID: v.YoutubeID, AddedAt: addedAt},
		Video:      v,
	}
}

// syncedAt は同期したアイテムが追加された時刻。
var syncedAt = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

// stored はプレイリスト PL の、syncedAt に追加された保存済みのアイテムを返す。
func stored(itemID, videoID string) *library.PlaylistItem {
	return storedAt(itemID, videoID, syncedAt)
}

// storedAt はプレイリスト PL の、addedAt に追加された保存済みのアイテムを返す。
func storedAt(itemID, videoID string, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{
		PlaylistYoutubeID: "PL",
		ItemID:            library.ItemID(itemID),
		VideoYoutubeID:    library.VideoID(videoID),
		AddedAt:           addedAt,
	}
}

// fetched は YouTube から取得した、syncedAt に追加されたアイテムを返す。
func fetched(itemID, videoID string, status library.PrivacyStatus) *library.FetchedItem {
	return fetchedAt(itemID, videoID, syncedAt, status)
}

// fetchedAt は YouTube から取得した、addedAt に追加されたアイテムを返す。
func fetchedAt(itemID, videoID string, addedAt time.Time, status library.PrivacyStatus) *library.FetchedItem {
	return &library.FetchedItem{
		ItemID:        library.ItemID(itemID),
		VideoID:       library.VideoID(videoID),
		AddedAt:       addedAt,
		PrivacyStatus: status,
	}
}

// playlist はプレイリスト PL を、指定したアイテムを持つ形で返す。
func playlist(items ...*library.PlaylistItem) *library.Playlist {
	return &library.Playlist{YoutubeID: "PL", Items: items}
}

// recorded は 2026-10-01 0 時 (UTC) から hours 時間後に記録された状態を返す。
func recorded(hours int, views library.ViewCount) *library.VideoSnapshot {
	return &library.VideoSnapshot{RecordedAt: hoursLater(hours), Views: views}
}

// hoursLater は 2026-10-01 0 時 (UTC) から hours 時間後の時刻を返す。
func hoursLater(hours int) time.Time {
	return time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(hours) * time.Hour)
}
