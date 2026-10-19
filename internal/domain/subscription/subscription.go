// Package subscription は、YouTube のプレイリストを通知するための
// ギルドごとの設定の集約。
package subscription

import (
	"sort"
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

// TakeNewVideos は同期したプレイリストから、前回の通知の後に追加された視聴できる
// 動画を選ぶ。あれば基準時刻を at に進め、タイトルを YouTube 上のものに合わせて、
// VideosAdded を返す。なければ何も変えずに nil を返す。YouTube から削除された
// プレイリストは通知しない。
//
// 返したイベントは、この通知設定を保存できた後にだけ配る。保存の前に送ると、
// 保存に失敗したときに次の回でもう一度送ってしまう。
func (s *Subscription) TakeNewVideos(latest *library.Snapshot, at time.Time) *VideosAdded {
	if latest.Deleted {
		return nil
	}

	var added []*library.PlaylistVideo

	for _, item := range latest.Items {
		if !s.IsNew(item) {
			continue
		}

		video, ok := latest.Videos[item.VideoYoutubeID]
		if !ok || !video.Available() {
			continue
		}

		added = append(added, &library.PlaylistVideo{PlaylistID: s.YoutubeID, Item: item, Video: video})
	}

	if len(added) == 0 {
		return nil
	}
	// プレイリストの順に関係なく、動画が追加された順に通知する
	sort.SliceStable(added, func(i, j int) bool {
		return added[i].Item.AddedAt.Before(added[j].Item.AddedAt)
	})

	// 次の回は、この時刻以降に追加された動画を新着にする。基準時刻を変えるのは
	// ここだけで、過去に戻すと、その時刻以降の動画をもう一度通知する。
	s.NotifiedUntil = at
	// タイトルは YouTube 上のものに合わせる
	s.Title = latest.Title

	return &VideosAdded{ChannelID: s.SendChannelID, PlaylistID: s.YoutubeID, PlaylistTitle: s.Title, Videos: added}
}

func (s *Subscription) ChangePickInterval(interval PickInterval) {
	s.PickInterval = interval
}
