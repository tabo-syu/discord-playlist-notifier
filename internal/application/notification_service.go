package application

import (
	"errors"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NotificationService は定期実行するユースケースを持つ。返す相手がいないので、
// 失敗はエラーとして返さずログに出す。
//
// 動画の非公開・節目・新着はドメインイベントとして events に発行し、通知は
// それを購読したハンドラ（NotificationHandlers）が送る。定期のまとめとピックは
// ドメインで起きたことではなく、時刻が来たら投稿するものなので、ここから直接送る。
type NotificationService struct {
	subscriptions subscription.Repository
	library       *LibraryService
	notifier      Notifier
	events        *EventPublisher
	// 投稿の基準にする日付や年のタイムゾーン
	location *time.Location
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewNotificationService(s subscription.Repository, l *LibraryService, n Notifier, e *EventPublisher, loc *time.Location, now func() time.Time) *NotificationService {
	return &NotificationService{s, l, n, e, loc, now}
}

// NotifyUpdates は通知登録された全てのプレイリストを同期し、非公開・削除に
// なった動画と、前回の通知以降に追加された動画のイベントを発行する。
//
// 新着のイベントは、通知設定の基準時刻を保存できた後にだけ発行する。全ての通知設定を
// 保存してから発行するので、送信は必ず保存の後になる。通知を取りこぼすことは
// あっても、二重には送らない。
func (s *NotificationService) NotifyUpdates() {
	subscriptions, err := s.subscriptions.FindAll()
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}

	var ids []library.PlaylistID
	for _, sub := range subscriptions {
		ids = append(ids, sub.YoutubeID)
	}

	snapshots, hidden, err := s.library.Sync(ids)
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}
	// 動画は Sync の中で保存済み
	Publish(s.events, hidden)

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

		if err := s.subscriptions.Update(sub); err != nil {
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

// RefreshVideos は保存済みの全動画の詳細を更新し、再生数の新しい節目に
// 達した動画のイベントを発行する。
func (s *NotificationService) RefreshVideos() {
	reached, err := s.library.RefreshVideos()
	if err != nil {
		log.Println("Could not refresh videos cause:", err)

		return
	}
	// 動画は RefreshVideos の中で保存済み
	Publish(s.events, reached)
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
		w, err := s.library.Wrapped(sub.YoutubeID, year)
		if err != nil {
			log.Println("Could not load playlist:", sub.YoutubeID, "cause:", err)

			continue
		}

		if w.Added == 0 {
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

		picked, err := s.library.RandomVideos([]library.PlaylistID{sub.YoutubeID}, 1)
		if err != nil {
			log.Println("Could not pick a video for playlist:", sub.YoutubeID, "cause:", err)

			continue
		}

		if len(picked) == 0 {
			log.Println("No video to pick for playlist:", sub.YoutubeID)

			continue
		}

		// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
		live, err := s.library.LiveVideos([]library.VideoID{picked[0].Video.YoutubeID})
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
