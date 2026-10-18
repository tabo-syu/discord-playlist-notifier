package mcpserver_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
)

func TestListPlaylists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		storedItems   []*library.PlaylistItem
		storedVideos  []*library.Video
		want          string
	}{
		{
			name: "プレイリストを曲数と一緒に返す",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "お気に入り"),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "1", "a", date(2026, time.October, 1, 0)),
				item("PL", "2", "b", date(2026, time.October, 2, 0)),
			},
			storedVideos: []*library.Video{
				video("a", 100),
				{YoutubeID: "b", Title: "b", PrivacyStatus: library.PrivacyDeleted},
			},
			want: `{"playlists": [{
				"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL",
				"items": 2, "available": 1, "private": 0, "deleted": 1,
				"first_added_at": "2026-10-01T00:00:00+09:00", "last_added_at": "2026-10-02T00:00:00+09:00"
			}]}`,
		},
		{
			name: "通知設定がなければ空の配列を返す",
			want: `{"playlists": []}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := connect(t, &fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos})

			got, gotError := call(t, session, "list_playlists", nil)

			assert.JSONEq(t, tt.want, got)
			assert.Empty(t, gotError)
		})
	}
}

func TestFindVideos(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "お気に入り"),
	}
	// 詳細を全て持つ動画。タグは 12 個あり、返すのは最初の 10 個だけ
	song := &library.Video{
		YoutubeID:        "song",
		Title:            "夜に駆ける",
		PrivacyStatus:    library.PrivacyPublic,
		Views:            1234,
		PublishedAt:      date(2019, time.December, 15, 0),
		Duration:         261 * time.Second,
		Tags:             []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		CategoryID:       "10",
		ChannelYoutubeID: "UC1",
	}
	// 詳細のほとんどがわからない動画
	game := &library.Video{
		YoutubeID:     "game",
		Title:         "実況",
		PrivacyStatus: library.PrivacyPublic,
		CategoryID:    "20",
	}
	newLibrary := func() *fake.LibraryRepository {
		return &fake.LibraryRepository{
			Items: []*library.PlaylistItem{
				// UTC では 10/17 の 15 時、JST では 10/18 の 0 時。JST で返すことを確かめる
				item("PL", "1", "song", time.Date(2026, time.October, 17, 15, 0, 0, 0, time.UTC)),
				item("PL", "2", "game", date(2026, time.October, 17, 23)),
			},
			Videos: []*library.Video{
				song,
				game,
			},
			Channels: []*library.Channel{
				{YoutubeID: "UC1", Title: "YOASOBI"},
			},
			Categories: []*library.Category{
				{YoutubeID: "10", Title: "音楽", Assignable: true},
				{YoutubeID: "20", Title: "ゲーム", Assignable: true},
				{YoutubeID: "99", Title: "古いカテゴリ", Assignable: false},
			},
		}
	}

	tests := []struct {
		name      string
		arguments map[string]any
		want      string
		wantError string
	}{
		{
			name: "曲の情報を、チャンネルとカテゴリの名前と一緒に返す",
			arguments: map[string]any{
				"title": "夜に",
			},
			want: `{"total": 1, "offset": 0, "videos": [{
				"video_id": "song", "title": "夜に駆ける", "url": "https://www.youtube.com/watch?v=song&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り",
				"added_at": "2026-10-18T00:00:00+09:00", "published_at": "2019-12-15T00:00:00+09:00",
				"views": 1234, "status": "public", "channel": "YOASOBI", "channel_id": "UC1",
				"duration_seconds": 261, "tags": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"], "category": "音楽"
			}]}`,
		},
		{
			name: "わからない情報は省き、タグは空の配列にする",
			arguments: map[string]any{
				"title": "実況",
			},
			want: `{"total": 1, "offset": 0, "videos": [{
				"video_id": "game", "title": "実況", "url": "https://www.youtube.com/watch?v=game&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り",
				"added_at": "2026-10-17T23:00:00+09:00",
				"views": 0, "status": "public", "tags": [], "category": "ゲーム"
			}]}`,
		},
		{
			name: "カテゴリを名前で絞り込む",
			arguments: map[string]any{
				"category": "ゲーム",
			},
			want: `{"total": 1, "offset": 0, "videos": [{
				"video_id": "game", "title": "実況", "url": "https://www.youtube.com/watch?v=game&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り",
				"added_at": "2026-10-17T23:00:00+09:00",
				"views": 0, "status": "public", "tags": [], "category": "ゲーム"
			}]}`,
		},
		{
			name: "追加された日を JST で区切る",
			arguments: map[string]any{
				"added_from": "2026-10-18",
			},
			want: `{"total": 1, "offset": 0, "videos": [{
				"video_id": "song", "title": "夜に駆ける", "url": "https://www.youtube.com/watch?v=song&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り",
				"added_at": "2026-10-18T00:00:00+09:00", "published_at": "2019-12-15T00:00:00+09:00",
				"views": 1234, "status": "public", "channel": "YOASOBI", "channel_id": "UC1",
				"duration_seconds": 261, "tags": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"], "category": "音楽"
			}]}`,
		},
		{
			name: "見つからなければ空の配列を返す",
			arguments: map[string]any{
				"title": "見つからない",
			},
			want: `{"total": 0, "offset": 0, "videos": []}`,
		},
		{
			name: "知らないカテゴリは、設定できるカテゴリの名前を並べて断る",
			arguments: map[string]any{
				"category": "アニメ",
			},
			want:      `null`,
			wantError: `category は次のどれかを指定してください: ゲーム、音楽（指定された値: "アニメ"）`,
		},
		{
			name: "通知登録されていないプレイリストは断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
		{
			name: "日付の形式が違う",
			arguments: map[string]any{
				"added_until": "2026/10/18",
			},
			want:      `null`,
			wantError: `added_until は YYYY-MM-DD の形式で指定してください: "2026/10/18"`,
		},
		{
			name: "知らない並び順",
			arguments: map[string]any{
				"sort": "random",
			},
			want:      `null`,
			wantError: `sort は added_desc、added_asc、views_desc のいずれかを指定してください: "random"`,
		},
		{
			name: "件数が多すぎる",
			arguments: map[string]any{
				"limit": 101,
			},
			want:      `null`,
			wantError: "limit は 1〜100、offset は 0 以上で指定してください",
		},
		{
			name: "飛ばす件数が負",
			arguments: map[string]any{
				"offset": -1,
			},
			want:      `null`,
			wantError: "limit は 1〜100、offset は 0 以上で指定してください",
		},
		{
			name: "長さが負",
			arguments: map[string]any{
				"min_duration_seconds": -1,
			},
			want:      `null`,
			wantError: "min_duration_seconds と max_duration_seconds は 0 以上で指定してください",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := connect(t, &fake.SubscriptionRepository{Subscriptions: subscriptions}, newLibrary())

			got, gotError := call(t, session, "find_videos", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestPlaylistStats(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "お気に入り"),
	}

	tests := []struct {
		name      string
		arguments map[string]any
		want      string
		wantError string
	}{
		{
			name: "プレイリストの統計を返す",
			arguments: map[string]any{
				"playlist_id": "PL",
			},
			want: `{
				"playlist": {"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL"},
				"items": 1, "available": 1, "private": 0, "deleted": 0,
				"added_this_month": 1, "added_last_30_days": 1,
				"first_added_at": "2026-10-01T00:00:00+09:00", "last_added_at": "2026-10-01T00:00:00+09:00",
				"top_viewed": [{"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL", "views": 100}],
				"monthly": [
					{"month": "2026-05", "count": 0},
					{"month": "2026-06", "count": 0},
					{"month": "2026-07", "count": 0},
					{"month": "2026-08", "count": 0},
					{"month": "2026-09", "count": 0},
					{"month": "2026-10", "count": 1}
				]
			}`,
		},
		{
			name: "通知登録されていないプレイリストは断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{
				Items:  []*library.PlaylistItem{item("PL", "1", "a", date(2026, time.October, 1, 0))},
				Videos: []*library.Video{video("a", 100)},
			}
			session := connect(t, &fake.SubscriptionRepository{Subscriptions: subscriptions}, lib)

			got, gotError := call(t, session, "playlist_stats", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestPlaylistWrapped(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "お気に入り"),
	}

	tests := []struct {
		name      string
		arguments map[string]any
		want      string
		wantError string
	}{
		{
			name: "年を省略すると今年をまとめる",
			arguments: map[string]any{
				"playlist_id": "PL",
			},
			want: `{
				"playlist": {"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL"},
				"year": 2026, "added": 1, "previous_year_added": 0,
				"busiest_month": {"month": "2026-03", "count": 1},
				"first": {
					"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL",
					"playlist_id": "PL", "playlist_title": "お気に入り", "added_at": "2026-03-01T00:00:00+09:00",
					"views": 100, "status": "public", "tags": []
				},
				"last": {
					"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL",
					"playlist_id": "PL", "playlist_title": "お気に入り", "added_at": "2026-03-01T00:00:00+09:00",
					"views": 100, "status": "public", "tags": []
				},
				"top_viewed": [{"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL", "views": 100}]
			}`,
		},
		{
			name: "追加された曲がない年は、一番多い月と最初・最後の曲を省く",
			arguments: map[string]any{
				"playlist_id": "PL",
				"year":        2025,
			},
			want: `{
				"playlist": {"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL"},
				"year": 2025, "added": 0, "previous_year_added": 0,
				"top_viewed": []
			}`,
		},
		{
			name: "通知登録されていないプレイリストは断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{
				Items:  []*library.PlaylistItem{item("PL", "1", "a", date(2026, time.March, 1, 0))},
				Videos: []*library.Video{video("a", 100)},
			}
			session := connect(t, &fake.SubscriptionRepository{Subscriptions: subscriptions}, lib)

			got, gotError := call(t, session, "playlist_wrapped", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestRandomVideos(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "お気に入り"),
	}

	// 選ぶ順はランダムなので、候補は 1 曲だけにする
	tests := []struct {
		name      string
		arguments map[string]any
		want      string
		wantError string
	}{
		{
			name:      "観られる曲を選ぶ",
			arguments: map[string]any{},
			want: `{"videos": [{
				"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り", "added_at": "2026-03-01T00:00:00+09:00",
				"views": 100, "status": "public", "tags": []
			}]}`,
		},
		{
			name: "曲数が多すぎる",
			arguments: map[string]any{
				"count": 11,
			},
			want:      `null`,
			wantError: "count は 1〜10 で指定してください",
		},
		{
			name: "通知登録されていないプレイリストは断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{
				Items:  []*library.PlaylistItem{item("PL", "1", "a", date(2026, time.March, 1, 0))},
				Videos: []*library.Video{video("a", 100)},
			}
			session := connect(t, &fake.SubscriptionRepository{Subscriptions: subscriptions}, lib)

			got, gotError := call(t, session, "random_videos", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestVideoHistory(t *testing.T) {
	t.Parallel()

	subscriptions := []*subscription.Subscription{
		sub(1, "PL", "お気に入り"),
	}

	tests := []struct {
		name      string
		arguments map[string]any
		want      string
		wantError string
	}{
		{
			name: "直近 30 日の記録を古い順に返す",
			arguments: map[string]any{
				"video_id": "a",
			},
			want: `{
				"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a",
				"snapshots": [
					{"recorded_at": "2026-10-10T00:00:00+09:00", "title": "a", "status": "public", "views": 100, "likes": 10, "comments": 1},
					{"recorded_at": "2026-10-18T00:00:00+09:00", "title": "a", "status": "public", "views": 200, "likes": 20, "comments": 2}
				]
			}`,
		},
		{
			name: "日数を指定する",
			arguments: map[string]any{
				"video_id": "a",
				"days":     1,
			},
			want: `{
				"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a",
				"snapshots": [
					{"recorded_at": "2026-10-18T00:00:00+09:00", "title": "a", "status": "public", "views": 200, "likes": 20, "comments": 2}
				]
			}`,
		},
		{
			name: "通知登録されたプレイリストにない動画は断る",
			arguments: map[string]any{
				"video_id": "unknown",
			},
			want:      `null`,
			wantError: "登録されているプレイリストに、その動画は見つかりません。find_videos で video_id を確認してください。",
		},
		{
			name: "日数が多すぎる",
			arguments: map[string]any{
				"video_id": "a",
				"days":     366,
			},
			want:      `null`,
			wantError: "days は 1〜365 で指定してください",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{
				Items:  []*library.PlaylistItem{item("PL", "1", "a", date(2026, time.March, 1, 0))},
				Videos: []*library.Video{video("a", 200)},
				Snapshots: []*library.VideoSnapshot{
					{VideoYoutubeID: "a", RecordedAt: date(2026, time.September, 1, 0), Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 50, Likes: 5},
					{VideoYoutubeID: "a", RecordedAt: date(2026, time.October, 10, 0), Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 100, Likes: 10, Comments: 1},
					{VideoYoutubeID: "a", RecordedAt: date(2026, time.October, 18, 0), Title: "a", PrivacyStatus: library.PrivacyPublic, Views: 200, Likes: 20, Comments: 2},
				},
			}
			session := connect(t, &fake.SubscriptionRepository{Subscriptions: subscriptions}, lib)

			got, gotError := call(t, session, "video_history", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}
