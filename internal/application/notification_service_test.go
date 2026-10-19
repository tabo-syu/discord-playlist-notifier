package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// 定期実行のユースケースは何も返さないので、mock に期待した呼び出し（保存と送信）だけを確かめる。

func TestNotificationServiceNotifyUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
	}{
		{
			name: "前回の通知の後に追加された動画を通知し、送る前に基準時刻とタイトルを保存する",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "古いタイトル", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": meta("PL", "新しいタイトル", 2),
				}, nil)
				m.expectUnchangedSync("PL", []*library.PlaylistItem{
					item("PL", "1", "old", hoursAgo(7)),
					item("PL", "2", "new", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"old": video("old", library.PrivacyPublic, 0, 0),
					"new": video("new", library.PrivacyPublic, 0, 0),
				})
				saved := m.subscriptions.EXPECT().Update(sub(1, "PL", "ch", "新しいタイトル", now)).Return(nil).Call
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"new"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&notification.NewVideos{
					Subscription:  sub(1, "PL", "ch", "新しいタイトル", now),
					PlaylistTitle: "新しいタイトル",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "2", "new", hoursAgo(1)), Video: video("new", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil).NotBefore(saved)
			},
		},
		{
			name: "新しい動画がなければ、通知も保存もしない",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "古いタイトル", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": meta("PL", "新しいタイトル", 1),
				}, nil)
				m.expectUnchangedSync("PL", []*library.PlaylistItem{
					item("PL", "1", "old", hoursAgo(7)),
				}, map[library.VideoID]*library.Video{
					"old": video("old", library.PrivacyPublic, 0, 0),
				})
			},
		},
		{
			name: "基準時刻の保存に失敗した通知設定には通知しない",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch", "PL1", lastNotified),
					sub(2, "PL2", "ch", "PL2", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL1": meta("PL1", "PL1", 1),
					"PL2": meta("PL2", "PL2", 1),
				}, nil)
				m.expectUnchangedSync("PL1", []*library.PlaylistItem{
					item("PL1", "1", "a", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				})
				m.expectUnchangedSync("PL2", []*library.PlaylistItem{
					item("PL2", "2", "b", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"b": video("b", library.PrivacyPublic, 0, 0),
				})
				m.subscriptions.EXPECT().Update(sub(1, "PL1", "ch", "PL1", now)).Return(errMock)
				m.subscriptions.EXPECT().Update(sub(2, "PL2", "ch", "PL2", now)).Return(nil)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"b"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&notification.NewVideos{
					Subscription:  sub(2, "PL2", "ch", "PL2", now),
					PlaylistTitle: "PL2",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil)
			},
		},
		{
			name: "チャンネルの取得や送信に失敗しても、送る",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": meta("PL", "PL", 1),
				}, nil)
				m.expectUnchangedSync("PL", []*library.PlaylistItem{
					item("PL", "1", "a", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				})
				m.subscriptions.EXPECT().Update(sub(1, "PL", "ch", "PL", now)).Return(nil)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(nil, errMock)
				m.notifier.EXPECT().NotifyNewVideos(&notification.NewVideos{
					Subscription:  sub(1, "PL", "ch", "PL", now),
					PlaylistTitle: "PL",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo(nil)).Return(errMock)
			},
		},
		{
			name: "見られなくなった動画を通知する",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{
					"PL": meta("PL", "PL", 1),
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(10)),
				), nil)
				m.youtube.EXPECT().FetchPlaylistItems(library.PlaylistID("PL")).Return([]*library.FetchedItem{
					fetched("1", "a", hoursAgo(10), library.PrivacyPrivate),
				}, nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
				m.playlists.EXPECT().Save(playlist("PL",
					item("PL", "1", "a", hoursAgo(10)),
				)).Return(nil)
				m.videos.EXPECT().Save([]*library.Video{
					video("a", library.PrivacyPrivate, 0, 0),
				}).Return(nil)
				m.notifier.EXPECT().NotifyHiddenVideo(&notification.HiddenVideo{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "PL",
					Video:         video("a", library.PrivacyPrivate, 0, 0),
				}).Return(nil)
			},
		},
		{
			name: "YouTube から消えたプレイリストには通知しない",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.youtube.EXPECT().FetchPlaylistMetas(library.PlaylistID("PL")).Return(map[library.PlaylistID]*library.PlaylistMeta{}, nil)
			},
		},
		{
			name: "通知設定がなければ何もしない",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, subscription.ErrNotSubscribed)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := application.NewNotificationService(m.subscriptions, newLibraryService(m, clock), m.notifier, jst, clock)

			s.NotifyUpdates()
		})
	}
}

func TestNotificationServiceRefreshVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
	}{
		{
			name: "新しい節目に達した動画を、それを含むプレイリストのチャンネルに通知する",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 90_000, 0),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 120_000},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save(mock.Anything).Return(nil)
				m.snapshots.EXPECT().Add(mock.Anything).Return(nil)
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch1", "PL1", lastNotified),
					sub(2, "PL2", "ch2", "PL2", lastNotified),
				}, nil)
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("PL1", item("PL1", "1", "a", hoursAgo(10))),
				}, nil)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyMilestone(&notification.Milestone{
					ChannelID:  "ch1",
					PlaylistID: "PL1",
					Video:      video("a", library.PrivacyPublic, 120_000, 100_000),
					Milestone:  100_000,
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil)
			},
		},
		{
			name: "節目に達した動画がなければ、通知設定を読まない",
			setup: func(m *deps) {
				m.videos.EXPECT().FindListed().Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 110_000, 100_000),
				}, nil)
				m.youtube.EXPECT().FetchVideos([]library.VideoID{"a"}).Return([]*library.Video{
					{YoutubeID: "a", Title: "a", Views: 120_000},
				}, []*library.Channel{}, nil)
				m.channels.EXPECT().Save([]*library.Channel{}).Return(nil)
				m.videos.EXPECT().Save(mock.Anything).Return(nil)
				m.snapshots.EXPECT().Add(mock.Anything).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := application.NewNotificationService(m.subscriptions, newLibraryService(m, clock), m.notifier, jst, clock)

			s.RefreshVideos()
		})
	}
}

func TestNotificationServicePostPicks(t *testing.T) {
	t.Parallel()

	// 選ぶ順はランダムなので、各プレイリストの候補は 1 本だけにする
	tests := []struct {
		name  string
		now   time.Time
		setup func(m *deps)
	}{
		{
			name: "毎日のピックは毎日投稿する",
			now:  date(2026, time.October, 18, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickDaily),
				}, nil)
				m.expectPick(picking(1, "PL", subscription.PickDaily))
			},
		},
		{
			name: "毎週のピックは月曜日に投稿する",
			now:  date(2026, time.October, 19, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickWeekly),
				}, nil)
				m.expectPick(picking(1, "PL", subscription.PickWeekly))
			},
		},
		{
			name: "毎週のピックは月曜日以外は投稿しない",
			now:  date(2026, time.October, 18, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickWeekly),
				}, nil)
			},
		},
		{
			name: "曜日は JST で決める",
			// UTC では日曜日の 15 時、JST では月曜日の 0 時
			now: time.Date(2026, time.October, 18, 15, 0, 0, 0, time.UTC),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickWeekly),
				}, nil)
				m.expectPick(picking(1, "PL", subscription.PickWeekly))
			},
		},
		{
			name: "毎月のピックは 1 日に投稿する",
			now:  date(2026, time.November, 1, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickMonthly),
				}, nil)
				m.expectPick(picking(1, "PL", subscription.PickMonthly))
			},
		},
		{
			name: "ピックを設定していなければ投稿しない",
			now:  date(2026, time.November, 1, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickNone),
				}, nil)
			},
		},
		{
			name: "選べる動画がなければ投稿しない",
			now:  date(2026, time.October, 18, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					picking(1, "PL", subscription.PickDaily),
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", hoursAgo(10)),
				), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPrivate, 0, 0),
				}, nil)
			},
		},
		{
			name: "通知設定がなければ何もしない",
			now:  date(2026, time.October, 18, 12),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, subscription.ErrNotSubscribed)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)

			fixed := func() time.Time { return tt.now }
			s := application.NewNotificationService(m.subscriptions, newLibraryService(m, fixed), m.notifier, jst, fixed)

			s.PostPicks()
		})
	}
}

func TestNotificationServicePostWrapped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		now   time.Time
		setup func(m *deps)
	}{
		{
			name: "今年追加された動画があるプレイリストのまとめを投稿する",
			now:  date(2026, time.December, 31, 21),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", date(2026, time.March, 1, 0)),
				), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 100, 0),
				}, nil)
				m.notifier.EXPECT().PostWrapped(sub(1, "PL", "ch", "PL", lastNotified), &library.Wrapped{
					Year:  2026,
					Added: 1,
					BusiestMonth: library.MonthCount{
						Month: date(2026, time.March, 1, 0),
						Count: 1,
					},
					TopViewed: []*library.Video{
						video("a", library.PrivacyPublic, 100, 0),
					},
					First: &library.PlaylistVideo{PlaylistID: "PL", Item: item("PL", "1", "a", date(2026, time.March, 1, 0)), Video: video("a", library.PrivacyPublic, 100, 0)},
					Last:  &library.PlaylistVideo{PlaylistID: "PL", Item: item("PL", "1", "a", date(2026, time.March, 1, 0)), Video: video("a", library.PrivacyPublic, 100, 0)},
				}).Return(nil)
			},
		},
		{
			name: "今年追加された動画がないプレイリストには投稿しない",
			now:  date(2026, time.December, 31, 21),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", date(2025, time.March, 1, 0)),
				), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 100, 0),
				}, nil)
			},
		},
		{
			name: "年は JST で決める",
			// UTC では 2026 年 12 月 31 日の 15 時、JST では 2027 年 1 月 1 日の 0 時
			now: time.Date(2026, time.December, 31, 15, 0, 0, 0, time.UTC),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().Find(library.PlaylistID("PL")).Return(playlist("PL",
					item("PL", "1", "a", date(2026, time.March, 1, 0)),
				), nil)
				m.videos.EXPECT().FindInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 100, 0),
				}, nil)
			},
		},
		{
			name: "通知設定がなければ何もしない",
			now:  date(2026, time.December, 31, 21),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, subscription.ErrNotSubscribed)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)

			fixed := func() time.Time { return tt.now }
			s := application.NewNotificationService(m.subscriptions, newLibraryService(m, fixed), m.notifier, jst, fixed)

			s.PostWrapped()
		})
	}
}
