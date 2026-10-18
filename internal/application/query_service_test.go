package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestQueryServicePlaylists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		want          []*application.WatchedPlaylist
	}{
		{
			name: "通知登録されたプレイリストを、1 つずつタイトル順に返す",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "ch1", "作業用", lastNotified),
				sub(2, "PL2", "ch1", "お気に入り", lastNotified),
				// 別のギルドが同じプレイリストを登録している
				sub(3, "PL1", "ch2", "作業用", lastNotified),
			},
			want: []*application.WatchedPlaylist{
				{ID: "PL2", Title: "お気に入り"},
				{ID: "PL1", Title: "作業用"},
			},
		},
		{
			name: "通知設定がなければ空のスライスを返す",
			want: []*application.WatchedPlaylist{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := application.NewQueryService(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, application.NewLibraryService(&fake.YouTube{}, &fake.LibraryRepository{}, jst, clock), &fake.LibraryRepository{})

			got, err := s.Playlists()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceFindVideos(t *testing.T) {
	t.Parallel()

	// 通知登録されているのは PL1 と PL2 だけで、UNWATCHED はどこも登録していない
	subscriptions := []*subscription.Subscription{
		sub(1, "PL1", "ch", "PL1", lastNotified),
		sub(2, "PL2", "ch", "PL2", lastNotified),
	}
	short := &library.Video{YoutubeID: "short", Title: "Short Song", PrivacyStatus: library.PrivacyPublic, Views: 300, Duration: 2 * time.Minute, Tags: []string{"J-Pop"}, CategoryID: "10", ChannelYoutubeID: "UC1"}
	long := &library.Video{YoutubeID: "long", Title: "Long Mix", PrivacyStatus: library.PrivacyPublic, Views: 100, Duration: 10 * time.Minute, Tags: []string{"Lo-Fi", "作業用"}, CategoryID: "10", ChannelYoutubeID: "UC2"}
	live := &library.Video{YoutubeID: "live", Title: "Live Stream", PrivacyStatus: library.PrivacyPublic, Views: 200, CategoryID: "20", ChannelYoutubeID: "UC1"}
	private := &library.Video{YoutubeID: "private", Title: "Secret", PrivacyStatus: library.PrivacyPrivate}
	unwatched := &library.Video{YoutubeID: "unwatched", Title: "Unwatched Song", PrivacyStatus: library.PrivacyPublic}

	shortInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "1", "short", hoursAgo(5)), Video: short}
	longInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "2", "long", hoursAgo(4)), Video: long}
	liveInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "3", "live", hoursAgo(3)), Video: live}
	privateInPL1 := &library.PlaylistVideo{PlaylistID: "PL1", Item: item("PL1", "4", "private", hoursAgo(2)), Video: private}
	// 複数のプレイリストにある動画は、アイテムごとに現れる
	shortInPL2 := &library.PlaylistVideo{PlaylistID: "PL2", Item: item("PL2", "5", "short", hoursAgo(1)), Video: short}

	newLibrary := func() *fake.LibraryRepository {
		return &fake.LibraryRepository{
			Items: []*library.PlaylistItem{
				shortInPL1.Item,
				longInPL1.Item,
				liveInPL1.Item,
				privateInPL1.Item,
				shortInPL2.Item,
				item("UNWATCHED", "6", "unwatched", hoursAgo(1)),
			},
			Videos: []*library.Video{
				short,
				long,
				live,
				private,
				unwatched,
			},
			Channels: []*library.Channel{
				{YoutubeID: "UC1", Title: "Official Channel"},
				{YoutubeID: "UC2", Title: "Lofi Girl"},
			},
		}
	}

	tests := []struct {
		name    string
		query   application.VideoQuery
		want    *application.VideoPage
		wantErr error
	}{
		{
			name:  "条件がなければ、通知登録されたプレイリストの見られる動画を新しい順に返す",
			query: application.VideoQuery{},
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
				PlaylistID:         "PL1",
				IncludeUnavailable: true,
			},
			want: &application.VideoPage{
				Total: 4,
				Videos: []*library.PlaylistVideo{
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
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
		{
			name: "タイトルの一部で絞り込む（大文字・小文字を区別しない）",
			query: application.VideoQuery{
				Title: "song",
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
			want: &application.VideoPage{
				Total:  4,
				Videos: []*library.PlaylistVideo{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := newLibrary()
			s := application.NewQueryService(&fake.SubscriptionRepository{Subscriptions: subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			got, err := s.FindVideos(tt.query)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceFindVideosWithoutSubscriptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query application.VideoQuery
		want  *application.VideoPage
	}{
		{
			name:  "通知設定がなければ空のスライスを返す",
			query: application.VideoQuery{},
			want: &application.VideoPage{
				Total:  0,
				Videos: []*library.PlaylistVideo{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{}
			s := application.NewQueryService(&fake.SubscriptionRepository{}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			got, err := s.FindVideos(tt.query)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceWrapped(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "ch", "PL", lastNotified),
	}
	a := video("a", library.PrivacyPublic, 100, 0)

	tests := []struct {
		name        string
		playlistID  library.PlaylistID
		storedItems []*library.PlaylistItem
		want        *library.Wrapped
		wantErr     error
	}{
		{
			name:       "通知登録されたプレイリストのまとめを返す",
			playlistID: "PL",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", date(2026, time.March, 1, 0)),
			},
			want: &library.Wrapped{
				Year:  2026,
				Added: 1,
				BusiestMonth: library.MonthCount{
					Month: date(2026, time.March, 1, 0),
					Count: 1,
				},
				TopViewed: []*library.Video{
					a,
				},
				First: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", date(2026, time.March, 1, 0)),
					Video:      a,
				},
				Last: &library.PlaylistVideo{
					PlaylistID: "PL",
					Item:       item("PL", "1", "a", date(2026, time.March, 1, 0)),
					Video:      a,
				},
			},
		},
		{
			name:       "通知登録されていないプレイリストはまとめない",
			playlistID: "UNWATCHED",
			storedItems: []*library.PlaylistItem{
				item("UNWATCHED", "1", "a", date(2026, time.March, 1, 0)),
			},
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: []*library.Video{a}}
			s := application.NewQueryService(&fake.SubscriptionRepository{Subscriptions: subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			got, err := s.Wrapped(tt.playlistID, 2026)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryServiceStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		playlistID library.PlaylistID
		wantErr    error
	}{
		{
			name:       "通知登録されたプレイリストは集計する",
			playlistID: "PL",
			wantErr:    nil,
		},
		{
			name:       "通知登録されていないプレイリストは集計しない",
			playlistID: "UNWATCHED",
			wantErr:    domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{}
			subscriptions := &fake.SubscriptionRepository{Subscriptions: []*subscription.Subscription{sub(1, "PL", "ch", "PL", lastNotified)}}
			s := application.NewQueryService(subscriptions, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			_, err := s.Stats(tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestQueryServiceRandomVideos(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "ch", "PL", lastNotified),
	}
	a := &library.PlaylistVideo{PlaylistID: "PL", Item: item("PL", "1", "a", hoursAgo(2)), Video: video("a", library.PrivacyPublic, 0, 0)}
	b := &library.PlaylistVideo{PlaylistID: "UNWATCHED", Item: item("UNWATCHED", "2", "b", hoursAgo(1)), Video: video("b", library.PrivacyPublic, 0, 0)}

	// 選ぶ順はランダムなので、全て選ぶ数を指定して中身だけを比べる
	tests := []struct {
		name       string
		playlistID library.PlaylistID
		want       []*library.PlaylistVideo
		wantErr    error
	}{
		{
			name:       "通知登録されたプレイリストからだけ選ぶ",
			playlistID: "",
			want: []*library.PlaylistVideo{
				a,
			},
		},
		{
			name:       "通知登録されていないプレイリストは指定できない",
			playlistID: "UNWATCHED",
			want:       nil,
			wantErr:    domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{Items: []*library.PlaylistItem{a.Item, b.Item}, Videos: []*library.Video{a.Video, b.Video}}
			s := application.NewQueryService(&fake.SubscriptionRepository{Subscriptions: subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			got, err := s.RandomVideos(tt.playlistID, 10)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestQueryServiceHistory(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "ch", "PL", lastNotified),
	}

	tests := []struct {
		name        string
		videoID     library.VideoID
		storedItems []*library.PlaylistItem
		want        *library.VideoHistory
		wantErr     error
	}{
		{
			name:    "通知登録されたプレイリストにある動画の推移を返す",
			videoID: "a",
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", hoursAgo(100)),
			},
			want: &library.VideoHistory{
				Video: video("a", library.PrivacyPublic, 0, 0),
				Snapshots: []*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100},
				},
			},
		},
		{
			name:    "通知登録されていないプレイリストにしかない動画は返さない",
			videoID: "a",
			storedItems: []*library.PlaylistItem{
				item("UNWATCHED", "1", "a", hoursAgo(100)),
			},
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{
				Items:     tt.storedItems,
				Videos:    []*library.Video{video("a", library.PrivacyPublic, 0, 0)},
				Snapshots: []*library.VideoSnapshot{{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100}},
			}
			s := application.NewQueryService(&fake.SubscriptionRepository{Subscriptions: subscriptions}, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)

			got, err := s.History(tt.videoID, hoursAgo(48))

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
