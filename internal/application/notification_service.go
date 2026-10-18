package application

import (
	"errors"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NotificationService は定期実行するユースケースを持つ。返す相手がいないので、
// 失敗はエラーとして返さずログに出す。
type NotificationService struct {
	subscriptions subscription.Repository
	library       *LibraryService
	notifier      Notifier
	// 投稿の基準にする日付や年のタイムゾーン
	location *time.Location
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewNotificationService(s subscription.Repository, l *LibraryService, n Notifier, loc *time.Location, now func() time.Time) *NotificationService {
	return &NotificationService{s, l, n, loc, now}
}

// NotifyUpdates は通知登録された全てのプレイリストを同期し、非公開・削除に
// なった動画と、前回の通知以降に追加された動画を通知する。
//
// 通知設定の基準時刻は動画を送る前に保存し、保存に成功した通知設定だけを
// 通知する。通知を取りこぼすことはあっても、二重には送らない。
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

	snapshots, err := s.library.Sync(ids)
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}

	s.notifyHiddenVideos(subscriptions, snapshots)

	for _, sub := range subscriptions {
		if snapshot, ok := snapshots[sub.YoutubeID]; ok && snapshot.Deleted {
			log.Println("Playlist(ID:", sub.YoutubeID, ") may have been deleted from YouTube")
		}
	}

	updates := notification.FindNewVideos(subscriptions, snapshots)
	if len(updates) == 0 {
		log.Println("Playlist was not updated")

		return
	}

	now := s.now()

	var saved []*notification.NewVideos

	for _, update := range updates {
		sub := update.Subscription
		sub.Rename(update.PlaylistTitle)
		sub.MarkNotified(now)

		if err := s.subscriptions.Update(sub); err != nil {
			log.Println("Could not update playlist:", sub.YoutubeID, "cause:", err)
		} else {
			saved = append(saved, update)
		}
	}

	for _, update := range saved {
		sub := update.Subscription
		// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
		var videoIDs []library.VideoID
		for _, v := range update.Videos {
			videoIDs = append(videoIDs, v.Video.YoutubeID)
		}

		live, err := s.library.LiveVideos(videoIDs)
		if err != nil {
			log.Println("Could not fetch channels for playlist:", sub.YoutubeID, "cause:", err)
		}

		if err := s.notifier.NotifyNewVideos(update, live); err != nil {
			log.Println("Message could not send to", sub.SendChannelID, "for playlist:", sub.YoutubeID, "cause:", err)
		} else {
			log.Println("Successfully sent notification for playlist:", sub.YoutubeID, "to channel:", sub.SendChannelID)
		}
	}
}

// RefreshVideos は保存済みの全動画の詳細を更新し、再生数の新しい節目に
// 達した動画を通知する。
func (s *NotificationService) RefreshVideos() {
	reached, err := s.library.RefreshVideos()
	if err != nil {
		log.Println("Could not refresh videos cause:", err)

		return
	}

	if len(reached) == 0 {
		return
	}

	subscriptions, err := s.subscriptions.FindAll()
	if err != nil {
		log.Println("Could not notify milestones cause:", err)

		return
	}

	var videoIDs []library.VideoID
	for _, v := range reached {
		videoIDs = append(videoIDs, v.YoutubeID)
	}

	containing, err := s.library.PlaylistsContaining(videoIDs)
	if err != nil {
		log.Println("Could not notify milestones cause:", err)

		return
	}

	// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
	live, err := s.library.LiveVideos(videoIDs)
	if err != nil {
		log.Println("Could not fetch channels for milestones cause:", err)
	}

	for _, notice := range notification.FindMilestones(subscriptions, reached, containing) {
		if err := s.notifier.NotifyMilestone(notice, live); err != nil {
			log.Println("Milestone notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent milestone notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "milestone:", notice.Milestone)
		}
	}
}

// PostWrapped は、今年動画が追加された通知登録済みの全プレイリストについて、
// その年のまとめを投稿する。
func (s *NotificationService) PostWrapped() {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
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
	if errors.Is(err, domain.ErrDBRecordNotFound) {
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

func (s *NotificationService) notifyHiddenVideos(subscriptions []*subscription.Subscription, snapshots map[library.PlaylistID]*library.Snapshot) {
	for _, notice := range notification.FindHiddenVideos(subscriptions, snapshots) {
		if err := s.notifier.NotifyHiddenVideo(notice); err != nil {
			log.Println("Hidden video notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent hidden video notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "status:", notice.Video.PrivacyStatus)
		}
	}
}
