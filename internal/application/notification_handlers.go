package application

import (
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NotificationHandlers は、ドメインイベントを Discord への通知にするハンドラを持つ。
// 各メソッドが 1 種類のイベントのハンドラで、cmd/bot/main.go が EventPublisher に
// 登録する。通知の種類を足すときは、ここにハンドラを足して登録する。
//
// ハンドラは発行した側の文脈（同期の結果など）に頼らず、要るものを自分で読む。
// 失敗は返す相手がいないのでログに出す。
type NotificationHandlers struct {
	subscriptions subscription.Repository
	catalog       *library.Catalog
	live          LiveVideoSource
	notifier      Notifier
}

func NewNotificationHandlers(s subscription.Repository, c *library.Catalog, l LiveVideoSource, n Notifier) *NotificationHandlers {
	return &NotificationHandlers{s, c, l, n}
}

// NotifyHiddenVideos は見られなくなった動画を、それを含むプレイリストのチャンネルに
// 通知する。
func (h *NotificationHandlers) NotifyHiddenVideos(events []*library.VideoHidden) {
	videoIDs := make([]library.VideoID, 0, len(events))
	for _, e := range events {
		videoIDs = append(videoIDs, e.Video.YoutubeID)
	}

	subscriptions, containing, err := h.recipients(videoIDs)
	if err != nil {
		log.Println("Could not notify hidden videos cause:", err)

		return
	}

	for _, notice := range notification.FindHiddenVideos(subscriptions, events, containing) {
		if err = h.notifier.NotifyHiddenVideo(notice); err != nil {
			log.Println("Hidden video notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent hidden video notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "status:", notice.Video.PrivacyStatus)
		}
	}
}

// NotifyMilestones は再生数の新しい節目に達した動画を、それを含むプレイリストの
// チャンネルに通知する。
func (h *NotificationHandlers) NotifyMilestones(events []*library.MilestoneReached) {
	videoIDs := make([]library.VideoID, 0, len(events))
	for _, e := range events {
		videoIDs = append(videoIDs, e.Video.YoutubeID)
	}

	subscriptions, containing, err := h.recipients(videoIDs)
	if err != nil {
		log.Println("Could not notify milestones cause:", err)

		return
	}

	// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
	live, err := h.live.FetchLiveVideos(videoIDs)
	if err != nil {
		log.Println("Could not fetch channels for milestones cause:", err)
	}

	for _, notice := range notification.FindMilestones(subscriptions, events, containing) {
		if err = h.notifier.NotifyMilestone(notice, live); err != nil {
			log.Println("Milestone notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent milestone notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "milestone:", notice.Milestone)
		}
	}
}

// NotifyNewVideos は、通知設定のプレイリストに追加された動画を、その通知設定の
// チャンネルに通知する。
func (h *NotificationHandlers) NotifyNewVideos(events []*subscription.VideosAdded) {
	for _, e := range events {
		// チャンネルは保存していないのでここで取得する。失敗したらチャンネルなしで送る。
		videoIDs := make([]library.VideoID, 0, len(e.Videos))
		for _, v := range e.Videos {
			videoIDs = append(videoIDs, v.Video.YoutubeID)
		}

		live, err := h.live.FetchLiveVideos(videoIDs)
		if err != nil {
			log.Println("Could not fetch channels for playlist:", e.PlaylistID, "cause:", err)
		}

		if err = h.notifier.NotifyNewVideos(e, live); err != nil {
			log.Println("Message could not send to", e.ChannelID, "for playlist:", e.PlaylistID, "cause:", err)
		} else {
			log.Println("Successfully sent notification for playlist:", e.PlaylistID, "to channel:", e.ChannelID)
		}
	}
}

// recipients は、動画を通知する先を決めるのに要る、全ての通知設定と、動画ごとに
// それを含むプレイリストを読む。
func (h *NotificationHandlers) recipients(videoIDs []library.VideoID) ([]*subscription.Subscription, map[library.VideoID][]library.PlaylistID, error) {
	subscriptions, err := h.subscriptions.FindAll()
	if err != nil {
		return nil, nil, err
	}

	containing, err := h.catalog.PlaylistsContaining(videoIDs)
	if err != nil {
		return nil, nil, err
	}

	return subscriptions, containing, nil
}
