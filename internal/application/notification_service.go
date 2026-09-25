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

// NotificationService holds the use cases run on a schedule. Failures are
// logged rather than returned, since there is no one to return them to.
type NotificationService struct {
	subscriptions subscription.Repository
	library       *LibraryService
	notifier      Notifier
	// Time zone of the days and years the posts are based on
	location *time.Location
}

func NewNotificationService(s subscription.Repository, l *LibraryService, n Notifier, loc *time.Location) *NotificationService {
	return &NotificationService{s, l, n, loc}
}

// NotifyUpdates syncs every subscribed playlist, and notifies the videos that
// became hidden and the videos added since the last notification.
//
// The base time of a subscription is saved before its videos are sent, and
// only the subscriptions saved successfully are notified. A notification may
// be missed, but is never sent twice.
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

	now := time.Now()
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
		// The channel is not stored, so fetch it now. Send without it on failure.
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

func (s *NotificationService) notifyHiddenVideos(subscriptions []*subscription.Subscription, snapshots map[library.PlaylistID]*library.Snapshot) {
	for _, notice := range notification.FindHiddenVideos(subscriptions, snapshots) {
		if err := s.notifier.NotifyHiddenVideo(notice); err != nil {
			log.Println("Hidden video notice could not send to", notice.ChannelID, "video:", notice.Video.YoutubeID, "cause:", err)
		} else {
			log.Println("Sent hidden video notice to", notice.ChannelID, "video:", notice.Video.YoutubeID, "status:", notice.Video.PrivacyStatus)
		}
	}
}

// RefreshVideos updates the details of every stored video, and notifies the
// videos that reached a new view milestone.
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

	// The channel is not stored, so fetch it now. Send without it on failure.
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

// PostWrapped posts the summary of this year of every subscribed playlist
// that had videos added in the year.
func (s *NotificationService) PostWrapped() {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return
	}
	if err != nil {
		log.Println("Could not post wrapped cause:", err)
		return
	}

	year := time.Now().In(s.location).Year()
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

// PostPicks posts a random video of every subscribed playlist whose pick
// interval is due today.
func (s *NotificationService) PostPicks() {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return
	}
	if err != nil {
		log.Println("Could not post pick cause:", err)
		return
	}

	now := time.Now().In(s.location)
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

		// The channel is not stored, so fetch it now. Send without it on failure.
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
