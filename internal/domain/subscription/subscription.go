// Package subscription は、YouTube のプレイリストを通知するための
// ギルドごとの設定の集約。
package subscription

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
)

// ChannelID は通知を送る Discord のテキストチャンネルの ID。
type ChannelID string

// Subscription はギルドが YouTube のプレイリストの通知を受けるための設定。
// プレイリスト自体の中身は library に属する。
//
// UpdatedAt は新着動画の通知の基準時刻で、それ以降にプレイリストへ追加された
// 動画が新着になる（IsNew を参照）。
type Subscription struct {
	gorm.Model
	YoutubeID     library.PlaylistID
	SendChannelID ChannelID
	Title         string
	// プレイリストからランダムな動画を投稿する頻度
	PickInterval PickInterval
	// 通知設定が属する guild.Guild。他の集約は ID でのみ
	// 参照する。
	GuildID uint
}

// テーブル名はモデルのリネーム前の名前のままにしている。
func (Subscription) TableName() string {
	return "playlists"
}

func New(g *guild.Guild, channelID ChannelID, youtubeID library.PlaylistID, title string) *Subscription {
	return &Subscription{
		YoutubeID:     youtubeID,
		SendChannelID: channelID,
		Title:         title,
		GuildID:       g.ID,
	}
}

// IsNew はアイテムが前回の通知以降に追加されたかを返す。
// ちょうど同時刻に追加されたアイテムを含めるため、After ではなく !Before を使う。
func (s *Subscription) IsNew(item *library.PlaylistItem) bool {
	return !item.AddedAt.Before(s.UpdatedAt)
}

// MarkNotified は新着動画の通知の基準時刻を進める。過去に戻すと、
// その時刻以降に追加された動画をもう一度通知する。
func (s *Subscription) MarkNotified(at time.Time) {
	s.UpdatedAt = at
}

// Rename は YouTube 上のプレイリストのタイトルに合わせる。
func (s *Subscription) Rename(title string) {
	s.Title = title
}

func (s *Subscription) ChangePickInterval(interval PickInterval) {
	s.PickInterval = interval
}
