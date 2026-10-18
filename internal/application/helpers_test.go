package application_test

import (
	"io"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// TestMain はテストの出力を読みやすくするため、サービスのログを捨てる。
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// now はテストの現在時刻。
var now = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

func clock() time.Time {
	return now
}

// hoursAgo は now の hours 時間前の時刻を返す。
func hoursAgo(hours int) time.Time {
	return now.Add(-time.Duration(hours) * time.Hour)
}

// lastNotified は前回の通知の時刻。
var lastNotified = hoursAgo(6)

// guildOf は ID が id のギルドを返す。
func guildOf(id uint, discordID guild.DiscordID) *guild.Guild {
	return &guild.Guild{Model: gorm.Model{ID: id}, DiscordID: discordID}
}

// sub はギルド 1 の通知設定を返す。notifiedAt は新着動画の通知の基準時刻。
func sub(id uint, playlistID library.PlaylistID, channelID subscription.ChannelID, title string, notifiedAt time.Time) *subscription.Subscription {
	return &subscription.Subscription{
		Model:         gorm.Model{ID: id, UpdatedAt: notifiedAt},
		YoutubeID:     playlistID,
		SendChannelID: channelID,
		Title:         title,
		GuildID:       1,
	}
}

// fetched は YouTube 上のプレイリストのアイテムを返す。
func fetched(itemID library.ItemID, videoID library.VideoID, addedAt time.Time, status library.PrivacyStatus) *library.FetchedItem {
	return &library.FetchedItem{ItemID: itemID, VideoID: videoID, AddedAt: addedAt, PrivacyStatus: status}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// video は保存した動画を返す。タイトルは ID と同じにし、再生数の節目は milestone まで記録済みにする。
func video(id library.VideoID, status library.PrivacyStatus, views, milestone library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: status, Views: views, ViewMilestone: &milestone}
}

// picking はピックの間隔を interval にした、ギルド 1 の通知設定を返す。
func picking(id uint, playlistID library.PlaylistID, interval subscription.PickInterval) *subscription.Subscription {
	s := sub(id, playlistID, "ch", string(playlistID), lastNotified)
	s.ChangePickInterval(interval)

	return s
}

// date は JST の日時を返す。
func date(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, jst)
}
