package application_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// 通知のユースケースは何も返さないので、mock に期待した呼び出し（保存と送信）だけを確かめる。

func TestNotificationServiceTakeNewVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []*library.PlaylistSynced
		setup  func(m *deps)
	}{
		{
			name: "前回の通知の後に追加された動画を通知し、送る前に基準時刻とタイトルを保存する",
			events: []*library.PlaylistSynced{
				synced("PL", "新しいタイトル", []*library.PlaylistItem{
					item("PL", "1", "old", hoursAgo(7)),
					item("PL", "2", "new", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"old": video("old", library.PrivacyPublic, 0, 0),
					"new": video("new", library.PrivacyPublic, 0, 0),
				}),
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "古いタイトル", lastNotified),
				}, nil)
				saved := m.subscriptions.EXPECT().Update(sub(1, "PL", "ch", "新しいタイトル", now)).Return(nil).Call
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"new"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&subscription.VideosAdded{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "新しいタイトル",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "2", "new", hoursAgo(1)), Video: video("new", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil).NotBefore(saved)
			},
		},
		{
			name: "新しい動画がなければ、通知も保存もしない",
			events: []*library.PlaylistSynced{
				synced("PL", "新しいタイトル", []*library.PlaylistItem{
					item("PL", "1", "old", hoursAgo(7)),
				}, map[library.VideoID]*library.Video{
					"old": video("old", library.PrivacyPublic, 0, 0),
				}),
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "古いタイトル", lastNotified),
				}, nil)
			},
		},
		{
			name: "基準時刻の保存に失敗した通知設定には通知しない",
			events: []*library.PlaylistSynced{
				synced("PL1", "PL1", []*library.PlaylistItem{
					item("PL1", "1", "a", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}),
				synced("PL2", "PL2", []*library.PlaylistItem{
					item("PL2", "2", "b", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"b": video("b", library.PrivacyPublic, 0, 0),
				}),
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch", "PL1", lastNotified),
					sub(2, "PL2", "ch", "PL2", lastNotified),
				}, nil)
				m.subscriptions.EXPECT().Update(sub(1, "PL1", "ch", "PL1", now)).Return(errMock)
				m.subscriptions.EXPECT().Update(sub(2, "PL2", "ch", "PL2", now)).Return(nil)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"b"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&subscription.VideosAdded{
					ChannelID:     "ch",
					PlaylistID:    "PL2",
					PlaylistTitle: "PL2",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil)
			},
		},
		{
			name: "同期できなかったプレイリストの通知設定は進めない",
			events: []*library.PlaylistSynced{
				synced("PL1", "PL1", []*library.PlaylistItem{
					item("PL1", "1", "a", hoursAgo(10)),
				}, map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}),
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch", "PL1", lastNotified),
					sub(2, "PL2", "ch", "PL2", lastNotified),
				}, nil)
			},
		},
		{
			name: "YouTube から消えたプレイリストには通知しない",
			events: []*library.PlaylistSynced{
				{Snapshot: &library.Snapshot{Playlist: playlist("PL"), Deleted: true}},
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
			},
		},
		{
			name: "通知設定を読めなければ何もしない",
			events: []*library.PlaylistSynced{
				synced("PL", "PL", []*library.PlaylistItem{
					item("PL", "1", "a", hoursAgo(1)),
				}, map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}),
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, errMock)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			s := newNotificationService(m, clock)

			s.TakeNewVideos(tt.events)
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
			s := newNotificationService(m, fixed)

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
			s := newNotificationService(m, fixed)

			s.PostWrapped()
		})
	}
}
