package application_test

import (
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// ハンドラは何も返さないので、mock に期待した呼び出し（読み取りと送信）だけを確かめる。

func TestNotificationHandlersNotifyHiddenVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []*library.VideoHidden
		setup  func(m *deps)
	}{
		{
			name: "動画を含むプレイリストのチャンネルに、通知設定のタイトルで 1 回ずつ通知する",
			events: []*library.VideoHidden{
				{Video: video("a", library.PrivacyPrivate, 0, 0)},
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch1", "お気に入り", lastNotified),
					sub(2, "PL2", "ch1", "作業用", lastNotified),
					sub(3, "PL3", "ch2", "ドライブ", lastNotified),
				}, nil)
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("PL1", item("PL1", "1", "a", hoursAgo(10))),
					playlist("PL2", item("PL2", "2", "a", hoursAgo(10))),
				}, nil)
				m.notifier.EXPECT().NotifyHiddenVideo(&notification.HiddenVideo{
					ChannelID:     "ch1",
					PlaylistID:    "PL1",
					PlaylistTitle: "お気に入り",
					Video:         video("a", library.PrivacyPrivate, 0, 0),
				}).Return(nil)
			},
		},
		{
			name: "送信に失敗しても、残りを送る",
			events: []*library.VideoHidden{
				{Video: video("a", library.PrivacyPrivate, 0, 0)},
				{Video: video("b", library.PrivacyDeleted, 0, 0)},
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a", "b"}).Return([]*library.Playlist{
					playlist("PL",
						item("PL", "1", "a", hoursAgo(10)),
						item("PL", "2", "b", hoursAgo(9)),
					),
				}, nil)
				m.notifier.EXPECT().NotifyHiddenVideo(&notification.HiddenVideo{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "PL",
					Video:         video("a", library.PrivacyPrivate, 0, 0),
				}).Return(errMock)
				m.notifier.EXPECT().NotifyHiddenVideo(&notification.HiddenVideo{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "PL",
					Video:         video("b", library.PrivacyDeleted, 0, 0),
				}).Return(nil)
			},
		},
		{
			name: "動画を含むプレイリストを読めなければ、送らない",
			events: []*library.VideoHidden{
				{Video: video("a", library.PrivacyPrivate, 0, 0)},
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return(nil, errMock)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			h := newNotificationHandlers(m)

			h.NotifyHiddenVideos(tt.events)
		})
	}
}

func TestNotificationHandlersNotifyMilestones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []*library.MilestoneReached
		setup  func(m *deps)
	}{
		{
			name: "チャンネルを取得できなくても、チャンネルなしで送る",
			events: []*library.MilestoneReached{
				{Video: video("a", library.PrivacyPublic, 120_000, 100_000), Milestone: 100_000},
			},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.playlists.EXPECT().FindContaining([]library.VideoID{"a"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "a", hoursAgo(10))),
				}, nil)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(nil, errMock)
				m.notifier.EXPECT().NotifyMilestone(&notification.Milestone{
					ChannelID:  "ch",
					PlaylistID: "PL",
					Video:      video("a", library.PrivacyPublic, 120_000, 100_000),
					Milestone:  100_000,
				}, map[library.VideoID]*library.LiveVideo(nil)).Return(nil)
			},
		},
		{
			name: "通知設定を読めなければ、送らない",
			events: []*library.MilestoneReached{
				{Video: video("a", library.PrivacyPublic, 120_000, 100_000), Milestone: 100_000},
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
			h := newNotificationHandlers(m)

			h.NotifyMilestones(tt.events)
		})
	}
}

func TestNotificationHandlersNotifyNewVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []*subscription.VideosAdded
		setup  func(m *deps)
	}{
		{
			name: "1 つの送信に失敗しても、残りの通知設定のチャンネルに送る",
			events: []*subscription.VideosAdded{
				{
					ChannelID:     "ch1",
					PlaylistID:    "PL1",
					PlaylistTitle: "PL1",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
					},
				},
				{
					ChannelID:     "ch2",
					PlaylistID:    "PL2",
					PlaylistTitle: "PL2",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)},
					},
				},
			},
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&subscription.VideosAdded{
					ChannelID:     "ch1",
					PlaylistID:    "PL1",
					PlaylistTitle: "PL1",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL1", Item: item("PL1", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(errMock)
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"b"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
				m.notifier.EXPECT().NotifyNewVideos(&subscription.VideosAdded{
					ChannelID:     "ch2",
					PlaylistID:    "PL2",
					PlaylistTitle: "PL2",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL2", Item: item("PL2", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo{}).Return(nil)
			},
		},
		{
			name: "チャンネルを取得できなくても、チャンネルなしで送る",
			events: []*subscription.VideosAdded{
				{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "PL",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
					},
				},
			},
			setup: func(m *deps) {
				m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(nil, errMock)
				m.notifier.EXPECT().NotifyNewVideos(&subscription.VideosAdded{
					ChannelID:     "ch",
					PlaylistID:    "PL",
					PlaylistTitle: "PL",
					Videos: []*library.PlaylistVideo{
						{PlaylistID: "PL", Item: item("PL", "1", "a", hoursAgo(1)), Video: video("a", library.PrivacyPublic, 0, 0)},
					},
				}, map[library.VideoID]*library.LiveVideo(nil)).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			h := newNotificationHandlers(m)

			h.NotifyNewVideos(tt.events)
		})
	}
}
