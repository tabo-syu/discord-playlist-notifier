package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestQueryServicePlaylists(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *deps)
		want  []*application.WatchedPlaylist
	}{
		{
			name: "通知登録されたプレイリストを、1 つずつタイトル順に返す",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL1", "ch1", "作業用", lastNotified),
					sub(2, "PL2", "ch1", "お気に入り", lastNotified),
					// 別のギルドが同じプレイリストを登録している
					sub(3, "PL1", "ch2", "作業用", lastNotified),
				}, nil)
			},
			want: []*application.WatchedPlaylist{
				{ID: "PL2", Title: "お気に入り"},
				{ID: "PL1", Title: "作業用"},
			},
		},
		{
			name: "通知設定がなければ空のスライスを返す",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, domain.ErrDBRecordNotFound)
			},
			want: []*application.WatchedPlaylist{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got, err := newQueryService(m).Playlists()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceFindVideos(t *testing.T) {
	short := func() *library.Video {
		return &library.Video{YoutubeID: "short", Title: "Short Song", PrivacyStatus: library.PrivacyPublic, Views: 300, Duration: 2 * time.Minute, Tags: []string{"J-Pop"}, CategoryID: "10", ChannelYoutubeID: "UC1"}
	}
	long := func() *library.Video {
		return &library.Video{YoutubeID: "long", Title: "Long Mix", PrivacyStatus: library.PrivacyPublic, Views: 100, Duration: 10 * time.Minute, Tags: []string{"Lo-Fi", "作業用"}, CategoryID: "10", ChannelYoutubeID: "UC2"}
	}
	live := func() *library.Video {
		return &library.Video{YoutubeID: "live", Title: "Live Stream", PrivacyStatus: library.PrivacyPublic, Views: 200, CategoryID: "20", ChannelYoutubeID: "UC1"}
	}
	private := func() *library.Video {
		return &library.Video{YoutubeID: "private", Title: "Secret", PrivacyStatus: library.PrivacyPrivate}
	}
	shortInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "1", "short", hoursAgo(5)), Video: short()}
	longInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "2", "long", hoursAgo(4)), Video: long()}
	liveInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "3", "live", hoursAgo(3)), Video: live()}
	privateInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "4", "private", hoursAgo(2)), Video: private()}
	// 複数のプレイリストにある動画は、アイテムごとに現れる
	shortInPL2 := &library.PlaylistVideo{PlaylistID: "PL2", Item: item("PL2", "5", "short", hoursAgo(1)), Video: short()}

	// 通知登録されているのは PL1 と PL2
	watched := func(m *deps) {
		m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
			sub(1, "PL1", "ch", "PL1", lastNotified),
			sub(2, "PL2", "ch", "PL2", lastNotified),
		}, nil)
	}
	// 通知登録されたプレイリストの中身。通知登録されていないプレイリストは読まない
	contents := func(m *deps) {
		m.library.EXPECT().FindItems(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return([]*library.PlaylistItem{
			item("PL1", "1", "short", hoursAgo(5)),
			item("PL1", "2", "long", hoursAgo(4)),
			item("PL1", "3", "live", hoursAgo(3)),
			item("PL1", "4", "private", hoursAgo(2)),
			item("PL2", "5", "short", hoursAgo(1)),
		}, nil)
		m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL1"), library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
			"short":   short(),
			"long":    long(),
			"live":    live(),
			"private": private(),
		}, nil)
	}

	tests := []struct {
		name    string
		query   application.VideoQuery
		setup   func(m *deps)
		want    *application.VideoPage
		wantErr error
	}{
		{
			name:  "条件がなければ、通知登録されたプレイリストの見られる動画を新しい順に返す",
			query: application.VideoQuery{},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 4,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
					liveInPL1,
					longInPL1,
					shortInPL1,
				},
			},
		},
		{
			name: "見られない動画も含める",
			query: application.VideoQuery{
				IncludeUnavailable: true,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 5,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
					privateInPL1,
					liveInPL1,
					longInPL1,
					shortInPL1,
				},
			},
		},
		{
			name: "プレイリストを指定する",
			query: application.VideoQuery{
				PlaylistID: "PL2",
			},
			setup: func(m *deps) {
				watched(m)
				m.library.EXPECT().FindItems(library.PlaylistID("PL2")).Return([]*library.PlaylistItem{
					item("PL2", "5", "short", hoursAgo(1)),
				}, nil)
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL2")).Return(map[library.VideoID]*library.Video{
					"short": short(),
				}, nil)
			},
			want: &application.VideoPage{
				Total: 1,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
				},
			},
		},
		{
			name: "通知登録されていないプレイリストは指定できない",
			query: application.VideoQuery{
				PlaylistID: "UNWATCHED",
			},
			setup:   watched,
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
		{
			name: "タイトルの一部で絞り込む（大文字・小文字を区別しない）",
			query: application.VideoQuery{
				Title: "song",
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 2,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
					shortInPL1,
				},
			},
		},
		{
			name: "チャンネル名の一部で絞り込む",
			query: application.VideoQuery{
				Channel: "lofi",
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
				m.library.EXPECT().FindChannels([]library.ChannelID{"UC1", "UC2", "UC1", "UC1"}).Return(map[library.ChannelID]*library.Channel{
					"UC1": {YoutubeID: "UC1", Title: "Official Channel"},
					"UC2": {YoutubeID: "UC2", Title: "Lofi Girl"},
				}, nil)
			},
			want: &application.VideoPage{
				Total: 1,
				Videos: []*library.PlaylistVideo{
					longInPL1,
				},
			},
		},
		{
			name: "チャンネル ID で絞り込む",
			query: application.VideoQuery{
				ChannelID: "UC1",
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 3,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
					liveInPL1,
					shortInPL1,
				},
			},
		},
		{
			name: "タグの一部で絞り込む",
			query: application.VideoQuery{
				Tag: "lo-fi",
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 1,
				Videos: []*library.PlaylistVideo{
					longInPL1,
				},
			},
		},
		{
			name: "カテゴリで絞り込む",
			query: application.VideoQuery{
				CategoryID: "20",
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 1,
				Videos: []*library.PlaylistVideo{
					liveInPL1,
				},
			},
		},
		{
			name: "長さの下限で絞り込むと、長さのわからない動画は除く",
			query: application.VideoQuery{
				MinDuration: 5 * time.Minute,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 1,
				Videos: []*library.PlaylistVideo{
					longInPL1,
				},
			},
		},
		{
			name: "長さの上限で絞り込むと、長さのわからない動画は除く",
			query: application.VideoQuery{
				MaxDuration: 5 * time.Minute,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 2,
				Videos: []*library.PlaylistVideo{
					shortInPL2,
					shortInPL1,
				},
			},
		},
		{
			name: "追加された時刻で絞り込む（開始は含み、終了は含まない）",
			query: application.VideoQuery{
				AddedFrom:  hoursAgo(4),
				AddedUntil: hoursAgo(1),
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 2,
				Videos: []*library.PlaylistVideo{
					liveInPL1,
					longInPL1,
				},
			},
		},
		{
			name: "古い順に並べる",
			query: application.VideoQuery{
				Sort: application.SortAddedAsc,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 4,
				Videos: []*library.PlaylistVideo{
					shortInPL1,
					longInPL1,
					liveInPL1,
					shortInPL2,
				},
			},
		},
		{
			name: "再生数の多い順に並べる（同じなら古い順）",
			query: application.VideoQuery{
				Sort: application.SortViewsDesc,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 4,
				Videos: []*library.PlaylistVideo{
					shortInPL1,
					shortInPL2,
					liveInPL1,
					longInPL1,
				},
			},
		},
		{
			name: "ページに分ける",
			query: application.VideoQuery{
				Offset: 1,
				Limit:  2,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total: 4,
				Videos: []*library.PlaylistVideo{
					liveInPL1,
					longInPL1,
				},
			},
		},
		{
			name: "全件より後のページは空のスライスを返す",
			query: application.VideoQuery{
				Offset: 10,
			},
			setup: func(m *deps) {
				watched(m)
				contents(m)
			},
			want: &application.VideoPage{
				Total:  4,
				Videos: []*library.PlaylistVideo{},
			},
		},
		{
			name:  "通知設定がなければ空のスライスを返す",
			query: application.VideoQuery{},
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, domain.ErrDBRecordNotFound)
			},
			want: &application.VideoPage{
				Total:  0,
				Videos: []*library.PlaylistVideo{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got, err := newQueryService(m).FindVideos(tt.query)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceWrapped(t *testing.T) {
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		setup      func(m *deps)
		want       *library.Wrapped
		wantErr    error
	}{
		{
			name:       "通知登録されたプレイリストのまとめを返す",
			playlistID: "PL",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.library.EXPECT().FindItems(library.PlaylistID("PL")).Return([]*library.PlaylistItem{
					item("PL", "1", "a", date(2026, time.March, 1, 0)),
				}, nil)
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 100, 0),
				}, nil)
			},
			want: &library.Wrapped{
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
			},
		},
		{
			name:       "通知登録されていないプレイリストは、中身を読まずに断る",
			playlistID: "UNWATCHED",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
			},
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got, err := newQueryService(m).Wrapped(tt.playlistID, 2026)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceStats(t *testing.T) {
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		setup      func(m *deps)
		wantErr    error
	}{
		{
			name:       "通知登録されたプレイリストは集計する",
			playlistID: "PL",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.library.EXPECT().FindItems(library.PlaylistID("PL")).Return([]*library.PlaylistItem{}, nil)
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{}, nil)
			},
			wantErr: nil,
		},
		{
			name:       "通知登録されていないプレイリストは、中身を読まずに断る",
			playlistID: "UNWATCHED",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
			},
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			_, err := newQueryService(m).Stats(tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestQueryServiceRandomVideos(t *testing.T) {
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		setup      func(m *deps)
		want       []*library.PlaylistVideo
		wantErr    error
	}{
		{
			name:       "プレイリストを省略すると、通知登録された全てのプレイリストから選ぶ",
			playlistID: "",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.library.EXPECT().FindItems(library.PlaylistID("PL")).Return([]*library.PlaylistItem{
					item("PL", "1", "a", hoursAgo(2)),
				}, nil)
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
			},
			want: []*library.PlaylistVideo{
				{PlaylistID: "PL", Item: item("PL", "1", "a", hoursAgo(2)), Video: video("a", library.PrivacyPublic, 0, 0)},
			},
		},
		{
			name:       "通知登録されていないプレイリストは、中身を読まずに断る",
			playlistID: "UNWATCHED",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
			},
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
		{
			name:       "通知設定がなければ空のスライスを返す",
			playlistID: "",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return(nil, domain.ErrDBRecordNotFound)
			},
			want: []*library.PlaylistVideo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got, err := newQueryService(m).RandomVideos(tt.playlistID, 10)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceHistory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *deps)
		want    *library.VideoHistory
		wantErr error
	}{
		{
			name: "通知登録されたプレイリストにある動画の推移を返す",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.library.EXPECT().FindItemsByVideos([]library.VideoID{"a"}).Return([]*library.PlaylistItem{
					item("PL", "1", "a", hoursAgo(100)),
				}, nil)
				m.library.EXPECT().FindVideos([]library.VideoID{"a"}).Return(map[library.VideoID]*library.Video{
					"a": video("a", library.PrivacyPublic, 0, 0),
				}, nil)
				m.library.EXPECT().FindSnapshots(library.VideoID("a"), hoursAgo(48)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100},
				}, nil)
			},
			want: &library.VideoHistory{
				Video: video("a", library.PrivacyPublic, 0, 0),
				Snapshots: []*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100},
				},
			},
		},
		{
			name: "通知登録されていないプレイリストにしかない動画は返さない",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindAll().Return([]*subscription.Subscription{
					sub(1, "PL", "ch", "PL", lastNotified),
				}, nil)
				m.library.EXPECT().FindItemsByVideos([]library.VideoID{"a"}).Return([]*library.PlaylistItem{
					item("UNWATCHED", "1", "a", hoursAgo(100)),
				}, nil)
			},
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got, err := newQueryService(m).History("a", hoursAgo(48))

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
