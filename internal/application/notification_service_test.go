package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestNotificationServiceNotifyUpdates(t *testing.T) {
	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		storedItems       []*library.PlaylistItem
		storedVideos      []*library.Video
		playlists         []*fake.Playlist
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "新しいタイトル",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "新しいタイトル",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL1",
					Title: "PL1",
					Items: []*library.FetchedItem{
						fetched("1", "a", hoursAgo(1), library.PrivacyPublic),
					},
				},
				{
					ID:    "PL2",
					Title: "PL2",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "PL",
					Items: []*library.FetchedItem{
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
			playlists: []*fake.Playlist{
				{
					ID:    "PL",
					Title: "PL",
					Items: []*library.FetchedItem{
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
			subscriptions := &fake.SubscriptionRepository{Subscriptions: tt.subscriptions, UpdateFails: tt.updateFails, SavedAt: now}
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			yt := &fake.YouTube{Playlists: tt.playlists}
			n := &fake.Notifier{Fails: tt.sendFails}
			s := application.NewNotificationService(subscriptions, application.NewLibraryService(yt, lib, jst, clock), n, jst, clock)

			s.NotifyUpdates()

			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
			assert.Equal(t, tt.wantNewVideos, n.NewVideos)
			assert.Equal(t, tt.wantHidden, n.Hidden)
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
			subscriptions := &fake.SubscriptionRepository{Subscriptions: tt.subscriptions, SavedAt: now}
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			yt := &fake.YouTube{Views: tt.views}
			n := &fake.Notifier{}
			s := application.NewNotificationService(subscriptions, application.NewLibraryService(yt, lib, jst, clock), n, jst, clock)

			s.RefreshVideos()

			assert.Equal(t, tt.wantMilestones, n.Milestones)
		})
	}
}

func TestNotificationServicePostPicks(t *testing.T) {
	// 選ぶ順はランダムなので、各プレイリストの候補は 1 曲だけにする
	tests := []struct {
		name          string
		now           time.Time
		subscriptions []*subscription.Subscription
		storedItems   []*library.PlaylistItem
		storedVideos  []*library.Video
		want          []*fake.PickPost
	}{
		{
			name: "毎日のピックは毎日投稿する",
			now:  date(2026, time.October, 18, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickDaily),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: []*fake.PickPost{
				{Subscription: picking(1, "PL", subscription.PickDaily), Picked: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", hoursAgo(10)),
					Video:      video("a", library.PrivacyPublic, 0, 0),
				}},
			},
		},
		{
			name: "毎週のピックは月曜日に投稿する",
			now:  date(2026, time.October, 19, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickWeekly),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: []*fake.PickPost{
				{Subscription: picking(1, "PL", subscription.PickWeekly), Picked: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", hoursAgo(10)),
					Video:      video("a", library.PrivacyPublic, 0, 0),
				}},
			},
		},
		{
			name: "毎週のピックは月曜日以外は投稿しない",
			now:  date(2026, time.October, 18, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickWeekly),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: nil,
		},
		{
			name: "曜日は JST で決める",
			// UTC では日曜日の 15 時、JST では月曜日の 0 時
			now: time.Date(2026, time.October, 18, 15, 0, 0, 0, time.UTC),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickWeekly),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: []*fake.PickPost{
				{Subscription: picking(1, "PL", subscription.PickWeekly), Picked: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", hoursAgo(10)),
					Video:      video("a", library.PrivacyPublic, 0, 0),
				}},
			},
		},
		{
			name: "毎月のピックは 1 日に投稿する",
			now:  date(2026, time.November, 1, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickMonthly),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: []*fake.PickPost{
				{Subscription: picking(1, "PL", subscription.PickMonthly), Picked: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", hoursAgo(10)),
					Video:      video("a", library.PrivacyPublic, 0, 0),
				}},
			},
		},
		{
			name: "通知設定ごとに、そのプレイリストから選ぶ",
			now:  date(2026, time.October, 18, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL1", subscription.PickDaily),
				picking(2, "PL2", subscription.PickDaily),
			},
			storedItems: []*library.PlaylistItem{
				item("PL1", "1", "a", hoursAgo(10)),
				item("PL2", "2", "b", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
				video("b", library.PrivacyPublic, 0, 0),
			},
			want: []*fake.PickPost{
				{Subscription: picking(1, "PL1", subscription.PickDaily), Picked: &library.PlaylistVideo{
					PlaylistID: "PL1",
					Item:       item("PL1", "1", "a", hoursAgo(10)),
					Video:      video("a", library.PrivacyPublic, 0, 0),
				}},
				{Subscription: picking(2, "PL2", subscription.PickDaily), Picked: &library.PlaylistVideo{
					PlaylistID: "PL2",
					Item:       item("PL2", "2", "b", hoursAgo(10)),
					Video:      video("b", library.PrivacyPublic, 0, 0),
				}},
			},
		},
		{
			name: "ピックを設定していなければ投稿しない",
			now:  date(2026, time.November, 1, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickNone),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 0, 0),
			},
			want: nil,
		},
		{
			name: "選べる曲がなければ投稿しない",
			now:  date(2026, time.October, 18, 12),
			subscriptions: []*subscription.Subscription{
				picking(1, "PL", subscription.PickDaily),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(10)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPrivate, 0, 0),
			},
			want: nil,
		},
		{
			name: "通知設定がなければ何もしない",
			now:  date(2026, time.October, 18, 12),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := func() time.Time { return tt.now }
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			n := &fake.Notifier{}
			s := application.NewNotificationService(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), n, jst, clock)

			s.PostPicks()

			assert.Equal(t, tt.want, n.Picks)
		})
	}
}

func TestNotificationServicePostWrapped(t *testing.T) {
	tests := []struct {
		name          string
		now           time.Time
		subscriptions []*subscription.Subscription
		storedItems   []*library.PlaylistItem
		storedVideos  []*library.Video
		want          []*fake.WrappedPost
	}{
		{
			name: "今年追加された曲があるプレイリストのまとめを投稿する",
			now:  date(2026, time.December, 31, 21),
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", date(2026, time.March, 1, 0)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 100, 0),
			},
			want: []*fake.WrappedPost{
				{
					Subscription: sub(1, "PL", "ch", "PL", lastNotified),
					Wrapped: &library.Wrapped{
						Year:  2026,
						Added: 1,
						BusiestMonth: library.MonthCount{
							Month: date(2026, time.March, 1, 0),
							Count: 1,
						},
						TopViewed: []*library.Video{
							video("a", library.PrivacyPublic, 100, 0),
						},
						First: &library.PlaylistVideo{
							PlaylistID: "PL",
							Item:       item("PL", "1", "a", date(2026, time.March, 1, 0)),
							Video:      video("a", library.PrivacyPublic, 100, 0),
						},
						Last: &library.PlaylistVideo{
							PlaylistID: "PL",
							Item:       item("PL", "1", "a", date(2026, time.March, 1, 0)),
							Video:      video("a", library.PrivacyPublic, 100, 0),
						},
					},
				},
			},
		},
		{
			name: "今年追加された曲がないプレイリストには投稿しない",
			now:  date(2026, time.December, 31, 21),
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", date(2025, time.March, 1, 0)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 100, 0),
			},
			want: nil,
		},
		{
			name: "年は JST で決める",
			// UTC では 2026 年 12 月 31 日の 15 時、JST では 2027 年 1 月 1 日の 0 時
			now: time.Date(2026, time.December, 31, 15, 0, 0, 0, time.UTC),
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "ch", "PL", lastNotified),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", date(2026, time.March, 1, 0)),
			},
			storedVideos: []*library.Video{
				video("a", library.PrivacyPublic, 100, 0),
			},
			want: nil,
		},
		{
			name: "通知設定がなければ何もしない",
			now:  date(2026, time.December, 31, 21),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := func() time.Time { return tt.now }
			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			n := &fake.Notifier{}
			s := application.NewNotificationService(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), n, jst, clock)

			s.PostWrapped()

			assert.Equal(t, tt.want, n.Wrapped)
		})
	}
}
