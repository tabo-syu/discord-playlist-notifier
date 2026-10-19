// Package subscription は、YouTube のプレイリストを通知するための
// ギルドごとの設定の集約。
package subscription

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// ChannelID は通知を送る Discord のテキストチャンネルの ID。
type ChannelID string

// Subscription はギルドが YouTube のプレイリストの通知を受けるための設定。
// プレイリスト自体の中身は library に属する。
type Subscription struct {
	// 保存したときにリポジトリが振る識別子
	ID            uint
	YoutubeID     library.PlaylistID
	SendChannelID ChannelID
	Title         string
	// プレイリストからランダムな動画を投稿する頻度
	PickInterval PickInterval
	// 通知設定が属する guild.Guild。他の集約は ID でのみ
	// 参照する。
	GuildID uint
	// 新着動画の通知の基準時刻。これより前にプレイリストへ追加された動画は
	// 通知済みか、登録より前からあったので通知しない。ちょうどこの時刻に
	// 追加された動画は新着に含める（IsNew を参照）。
	NotifiedUntil time.Time
}

// New は subscribedAt に登録した通知設定を返す。登録より前からプレイリストに
// あった動画は通知しない。
func New(g *guild.Guild, channelID ChannelID, youtubeID library.PlaylistID, title string, subscribedAt time.Time) *Subscription {
	return &Subscription{
		YoutubeID:     youtubeID,
		SendChannelID: channelID,
		Title:         title,
		GuildID:       g.ID,
		NotifiedUntil: subscribedAt,
	}
}

// IsNew はアイテムが前回の通知以降に追加されたかを返す。
// ちょうど同時刻に追加されたアイテムを含めるため、After ではなく !Before を使う。
func (s *Subscription) IsNew(item *library.PlaylistItem) bool {
	return !item.AddedAt.Before(s.NotifiedUntil)
}

// MarkNotified は新着動画の通知の基準時刻を進める。過去に戻すと、
// その時刻以降に追加された動画をもう一度通知する。
func (s *Subscription) MarkNotified(at time.Time) {
	s.NotifiedUntil = at
}

// Rename は YouTube 上のプレイリストのタイトルに合わせる。
func (s *Subscription) Rename(title string) {
	s.Title = title
}

func (s *Subscription) ChangePickInterval(interval PickInterval) {
	s.PickInterval = interval
}
