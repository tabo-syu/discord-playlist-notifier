package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestNotificationServiceNotifyUpdates(t *testing.T) {
	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		storedItems       []*library.PlaylistItem
		storedVideos      []*library.Video
		playlists         []*youtubePlaylist
		updateFails       []library.PlaylistID
		sendFails         bool
		wantSubscriptions []*subscription.Subscription
		wantNewVideos     []*notification.NewVideos
		wantHidden        []*notification.HiddenVideo
	}{
		{
			name: "前回の通知の後に追加された動画を通知し、基準時刻とタイトルを保存する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "古いタイトル", lastNotified),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "新しいタイトル",
					items: []*library.FetchedItem{
						fetched("1", "old", hoursAgo(7), library.PrivacyPublic),
						fetched("2", "new", hoursAgo(1), library.PrivacyPublic),
					},
				},
			},
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "新しいタイトル", now),
			},
			wantNewVideos: []*notification.NewVideos{
				{
					Subscription:  sub(1, "PL", "ch", "新しいタイトル", now),
					PlaylistTitle: "新しいタイトル",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "2", "new", hoursAgo(1)), Video: video("new", library.PrivacyPublic, 0, 0)},
					},
				},
			},
		},
		{
			name: "新しい動画がなければ、通知も保存もしない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "古いタイトル", lastNotified),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "新しいタイトル",
					items: []*library.FetchedItem{
						fetched("1", "old", hoursAgo(7), library.PrivacyPublic),
					},
				},
			},
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "古いタイトル", lastNotified),
			},
		},
		{
			name: "基準時刻の保存に失敗した通知設定には通知しない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch", "PL1", lastNotified),
				sub(2, "PL2", "ch", "PL2", lastNotified),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL1",
					title: "PL1",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
					},
				},
				{
					id:    "PL2",
					title: "PL2",
					items: []*library.FetchedItem{
						fetched("2", "b", hoursAgo(1), library.PrivacyPublic),
					},
				},
			},
			updateFails: []library.PlaylistID{"PL1"},
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch", "PL1", lastNotified),
				sub(2, "PL2", "ch", "PL2", now),
			},
			wantNewVideos: []*notification.NewVideos{
				{
					Subscription:  sub(2, "PL2", "ch", "PL2", now),
					PlaylistTitle: "PL2",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)},
					},
				},
			},
		},
		{
			name: "送信に失敗しても、二重に通知しないように基準時刻は戻さない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "PL",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
					},
				},
			},
			sendFails: true,
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", now),
			},
		},
		{
			name: "見られなくなった動画を通知する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			playlists: []*youtubePlaylist{
				{
					id:    "PL",
					title: "PL",
					items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(10), library.PrivacyPrivate),
					},
				},
			},
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			wantHidden: []*notification.HiddenVideo{
				{ChannelID: "ch", PlaylistID: "PL", PlaylistTitle: "PL", Video: video("a", library.PrivacyPrivate, 0, 0)},
			},
		},
		{
			name: "YouTube から消えたプレイリストには通知しない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
		},
		{
			name: "通知設定がなければ何もしない",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptions := &subscriptionRepository{subscriptions: tt.subscriptions, updateFails: tt.updateFails}
			lib := &libraryRepository{items: tt.storedItems, videos: tt.storedVideos}
			yt := &youtube{playlists: tt.playlists}
			n := &notifier{fails: tt.sendFails}
			s := application.NewNotificationService(subscriptions, application.NewLibraryService(yt, lib, jst, clock), n, jst, clock)

			s.NotifyUpdates()

			assert.Equal(t, tt.wantSubscriptions, subscriptions.subscriptions)
			assert.Equal(t, tt.wantNewVideos, n.newVideos)
			assert.Equal(t, tt.wantHidden, n.hidden)
		})
	}
}

func TestNotificationServiceRefreshVideos(t *testing.T) {
	tests := []struct {
		name           string
		subscriptions  []*subscription.Subscription
		storedItems    []*library.PlaylistItem
		storedVideos   []*library.Video
		views          map[library.VideoID]library.ViewCount
		wantMilestones []*notification.Milestone
	}{
		{
			name: "新しい節目に達した動画を、それを含むプレイリストのチャンネルに通知する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch1", "PL1", lastNotified),
				sub(2, "PL2", "ch2", "PL2", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL1", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 90_000, 0),
			},
			views: map[library.VideoID]library.ViewCount{
				"a": 120_000,
			},
			wantMilestones: []*notification.Milestone{
				{ChannelID: "ch1", PlaylistID: "PL1", Video: video("a", library.PrivacyPublic, 120_000, 100_000), Milestone: 100_000},
			},
		},
		{
			name: "同じ節目のままなら通知しない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch1", "PL1", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL1", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 110_000, 100_000),
			},
			views: map[library.VideoID]library.ViewCount{
				"a": 120_000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptions := &subscriptionRepository{subscriptions: tt.subscriptions}
			lib := &libraryRepository{items: tt.storedItems, videos: tt.storedVideos}
			yt := &youtube{views: tt.views}
			n := &notifier{}
			s := application.NewNotificationService(subscriptions, application.NewLibraryService(yt, lib, jst, clock), n, jst, clock)

			s.RefreshVideos()

			assert.Equal(t, tt.wantMilestones, n.milestones)
		})
	}
}
