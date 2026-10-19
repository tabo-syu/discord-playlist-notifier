package application

import (
	"errors"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NotificationService は通知のユースケースを持つ。返す相手がいないので、失敗は
// エラーとして返さずログに出す。
//
// 新着は、LibraryService が発行した同期の結果（PlaylistSynced）を TakeNewVideos で
// 受け取り、通知設定を進めて VideosAdded を発行する。通知はそれを購読したハンドラ
// （NotificationHandlers）が送る。定期のまとめとピックはドメインで起きたことでは
// なく、時刻が来たら投稿するものなので、スケジューラから呼ばれてここから直接送る。
type NotificationService struct {
	subscriptions subscription.Repository
	catalog       *library.Catalog
	live          LiveVideoSource
	notifier      Notifier
	events        *EventPublisher
	// 投稿の基準にする日付や年のタイムゾーン
	location *time.Location
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewNotificationService(s subscription.Repository, c *library.Catalog, l LiveVideoSource, n Notifier, e *EventPublisher, loc *time.Location, now func() time.Time) *NotificationService {
	return &NotificationService{s, c, l, n, e, loc, now}
}

// TakeNewVideos は PlaylistSynced のハンドラで、同期したプレイリストの通知設定ごとに
// 前回の通知以降に追加された動画を選び、基準時刻を進めて保存してから、VideosAdded を
// 発行する。
//
// 全ての通知設定を保存してから発行するので、送信は必ず保存の後になる。通知を
// 取りこぼすことはあっても、二重には送らない。
func (s *NotificationService) TakeNewVideos(events []*library.PlaylistSynced) {
	snapshots := map[library.PlaylistID]*library.Snapshot{}
	for _, e := range events {
		snapshots[e.Snapshot.YoutubeID] = e.Snapshot
	}

	subscriptions, err := s.subscriptions.FindAll()
	if err != nil {
		log.Println("Could not take new videos cause:", err)

		return
	}

	now := s.now()

	var (
		found bool
		saved []*subscription.VideosAdded
	)

	for _, sub := range subscriptions {
		latest, ok := snapshots[sub.YoutubeID]
		if !ok {
			continue
		}

		if latest.Deleted {
			log.Println("Playlist(ID:", sub.YoutubeID, ") may have been deleted from YouTube")
		}

		added := sub.TakeNewVideos(latest, now)
		if added == nil {
			continue
		}

		found = true

		if err = s.subscriptions.Update(sub); err != nil {
			log.Println("Could not update playlist:", sub.YoutubeID, "cause:", err)
		} else {
			saved = append(saved, added)
		}
	}

	if !found {
		log.Println("Playlist was not updated")

		return
	}

	Publish(s.events, saved)
}

// PostWrapped は、今年動画が追加された通知登録済みの全プレイリストについて、
// その年のまとめを投稿する。
func (s *NotificationService) PostWrapped() {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, subscription.ErrNotSubscribed) {
		return
	}

	if err != nil {
		log.Println("Could not post wrapped cause:", err)

		return
	}

	year := s.now().In(s.location).Year()
	for _, sub := range subscriptions {
		contents, err := s.catalog.Contents([]library.PlaylistID{sub.YoutubeID})
		if err != nil {
			log.Println("Could not load playlist:", sub.YoutubeID, "cause:", err)

			continue
		}

		w := library.ComputeWrapped(contents, year, s.location)
		if w.Empty() {
			continue
		}

		if err := s.notifier.PostWrapped(sub, w); err != nil {
			log.Println("Wrapped could not send to", sub.SendChannelID, "for playlist:", sub.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent wrapped to", sub.SendChannelID, "for playlist:", sub.YoutubeID, "year:", year)
		}
	}
}

// PostPicks は、ピックの間隔が今日に当たる通知登録済みの全プレイリストについて、
// ランダムな動画を投稿する。
func (s *NotificationService) PostPicks() {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, subscription.ErrNotSubscribed) {
		return
	}

	if err != nil {
		log.Println("Could not post pick cause:", err)

		return
	}

	now := s.now().In(s.location)
	for _, sub := range subscriptions {
		if !sub.PickInterval.DueOn(now) {
			continue
		}

		contents, err := s.catalog.Contents([]library.PlaylistID{sub.YoutubeID})
		if err != nil {
			log.Println("Could not pick a video for playlist:", sub.YoutubeID, "cause:", err)

			continue
		}

		picked := library.PickRandom(contents, 1)
		if len(picked) == 0 {
			log.Println("No video to pick for playlist:", sub.YoutubeID)

			continue
		}

		// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
		live, err := s.live.FetchLiveVideos([]library.VideoID{picked[0].Video.YoutubeID})
		if err != nil {
			log.Println("Could not fetch the channel for pick cause:", err)
		}

		if err := s.notifier.PostPick(sub, picked[0], live); err != nil {
			log.Println("Pick could not send to", sub.SendChannelID, "for playlist:", sub.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent pick to", sub.SendChannelID, "for playlist:", sub.YoutubeID, "interval:", sub.PickInterval, "video:", picked[0].Video.YoutubeID)
		}
	}
}
