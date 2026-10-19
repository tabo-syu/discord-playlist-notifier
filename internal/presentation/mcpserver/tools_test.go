package mcpserver_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestListPlaylists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  string
	}{
		{
			name: "プレイリストを動画数と一緒に返す",
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL",
					added("PL", video("a", 100), date(2026, time.October, 1, 0)),
					added("PL", &library.Video{YoutubeID: "b", DetailsFetched: true, Title: "b", PrivacyStatus: library.PrivacyDeleted}, date(2026, time.October, 2, 0)),
				)
			},
			want: `{"playlists": [{
				"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL",
				"items": 2, "available": 1, "private": 0, "deleted": 1,
				"first_added_at": "2026-10-01T00:00:00+09:00", "last_added_at": "2026-10-02T00:00:00+09:00"
			}]}`,
		},
		{
			name: "通知設定がなければ空の配列を返す",
			setup: func(m *deps) {
				m.watched()
			},
			want: `{"playlists": []}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "list_playlists", nil)

			assert.JSONEq(t, tt.want, got)
			assert.Empty(t, gotError)
		})
	}
}

func TestFindVideos(t *testing.T) {
	t.Parallel()

	// 詳細を全て持つ動画。タグは 12 個あり、返すのは最初の 10 個だけ。
	// UTC では 10/17 の 15 時、JST では 10/18 の 0 時に追加された。JST で返すことを確かめる
	song := &library.PlaylistVideo{
		PlaylistID: "PL",
		Item:       item("PL", "1", "song", time.Date(2026, time.October, 17, 15, 0, 0, 0, time.UTC)),
		Video: &library.Video{
			YoutubeID:        "song",
			DetailsFetched:   true,
			Title:            "夜に駆ける",
			PrivacyStatus:    library.PrivacyPublic,
			Views:            1234,
			PublishedAt:      date(2019, time.December, 15, 0),
			Duration:         261 * time.Second,
			Tags:             []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
			CategoryID:       "10",
			ChannelYoutubeID: "UC1",
		},
	}
	// 詳細のほとんどがわからない動画
	game := &library.PlaylistVideo{
		PlaylistID: "PL",
		Item:       item("PL", "2", "game", date(2026, time.October, 17, 23)),
		Video: &library.Video{
			YoutubeID:      "game",
			DetailsFetched: true,
			Title:          "実況",
			PrivacyStatus:  library.PrivacyPublic,
			CategoryID:     "20",
		},
	}
	channels := map[library.ChannelID]*library.Channel{
		"UC1": {YoutubeID: "UC1", Title: "YOASOBI"},
	}
	songOut := `{
		"video_id": "song", "title": "夜に駆ける", "url": "https://www.youtube.com/watch?v=song&list=PL",
		"playlist_id": "PL", "playlist_title": "お気に入り",
		"added_at": "2026-10-18T00:00:00+09:00", "published_at": "2019-12-15T00:00:00+09:00",
		"views": 1234, "status": "public", "channel": "YOASOBI", "channel_id": "UC1",
		"duration_seconds": 261, "tags": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"], "category": "音楽"
	}`
	gameOut := `{
		"video_id": "game", "title": "実況", "url": "https://www.youtube.com/watch?v=game&list=PL",
		"playlist_id": "PL", "playlist_title": "お気に入り",
		"added_at": "2026-10-17T23:00:00+09:00",
		"views": 0, "status": "public", "tags": [], "category": "ゲーム"
	}`

	tests := []struct {
		name      string
		arguments map[string]any
		setup     func(m *deps)
		want      string
		wantError string
	}{
		{
			name: "動画の情報を、チャンネルとカテゴリの名前と一緒に返す",
			arguments: map[string]any{
				"title": "夜に",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", song, game)
				m.names([]library.ChannelID{"UC1"}, channels)
			},
			want: `{"total": 1, "offset": 0, "videos": [` + songOut + `]}`,
		},
		{
			name: "わからない情報は省き、タグは空の配列にする",
			arguments: map[string]any{
				"title": "実況",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", song, game)
				m.names([]library.ChannelID{""}, map[library.ChannelID]*library.Channel{})
			},
			want: `{"total": 1, "offset": 0, "videos": [` + gameOut + `]}`,
		},
		{
			name: "カテゴリを名前で絞り込む",
			arguments: map[string]any{
				"category": "ゲーム",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", song, game)
				m.names([]library.ChannelID{""}, map[library.ChannelID]*library.Channel{})
			},
			want: `{"total": 1, "offset": 0, "videos": [` + gameOut + `]}`,
		},
		{
			name: "追加された日を JST で区切る",
			arguments: map[string]any{
				"added_from": "2026-10-18",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", song, game)
				m.names([]library.ChannelID{"UC1"}, channels)
			},
			want: `{"total": 1, "offset": 0, "videos": [` + songOut + `]}`,
		},
		{
			name: "見つからなければ空の配列を返す",
			arguments: map[string]any{
				"title": "見つからない",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", song, game)
				m.names([]library.ChannelID(nil), map[library.ChannelID]*library.Channel{})
			},
			want: `{"total": 0, "offset": 0, "videos": []}`,
		},
		{
			name: "知らないカテゴリは、設定できるカテゴリの名前を並べて断る",
			arguments: map[string]any{
				"category": "アニメ",
			},
			setup: func(m *deps) {
				m.categories.EXPECT().FindAll().Return(categories(), nil)
			},
			want:      `null`,
			wantError: `category は次のどれかを指定してください: ゲーム、音楽（指定された値: "アニメ"）`,
		},
		{
			name: "通知登録されていないプレイリストは、中身を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
		{
			name: "日付の形式が違う",
			arguments: map[string]any{
				"added_until": "2026/10/18",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: `added_until は YYYY-MM-DD の形式で指定してください: "2026/10/18"`,
		},
		{
			name: "プレイリストの ID の形式が正しくなければ、DB を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "PL ID",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "playlist_id の形式が正しくありません。list_playlists で ID を確認してください。",
		},
		{
			name: "チャンネルの ID の形式が正しくなければ、DB を読まずに断る",
			arguments: map[string]any{
				"channel_id": "@YOASOBI",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "channel_id の形式が正しくありません。find_videos の結果の channel_id を指定してください。",
		},
		{
			name: "知らない並び順",
			arguments: map[string]any{
				"sort": "random",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: `sort は added_desc、added_asc、views_desc のいずれかを指定してください: "random"`,
		},
		{
			name: "件数が多すぎる",
			arguments: map[string]any{
				"limit": 101,
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "limit は 1〜100、offset は 0 以上で指定してください",
		},
		{
			name: "飛ばす件数が負",
			arguments: map[string]any{
				"offset": -1,
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "limit は 1〜100、offset は 0 以上で指定してください",
		},
		{
			name: "長さが負",
			arguments: map[string]any{
				"min_duration_seconds": -1,
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "min_duration_seconds と max_duration_seconds は 0 以上で指定してください",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "find_videos", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestPlaylistStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments map[string]any
		setup     func(m *deps)
		want      string
		wantError string
	}{
		{
			name: "プレイリストの統計を返す",
			arguments: map[string]any{
				"playlist_id": "PL",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", added("PL", video("a", 100), date(2026, time.October, 1, 0)))
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
			name: "通知登録されていないプレイリストは、中身を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
		{
			name: "プレイリストの ID の形式が正しくなければ、DB を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "https://www.youtube.com/playlist?list=PL",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "playlist_id の形式が正しくありません。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "playlist_stats", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestPlaylistWrapped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments map[string]any
		setup     func(m *deps)
		want      string
		wantError string
	}{
		{
			name: "年を省略すると今年をまとめる",
			arguments: map[string]any{
				"playlist_id": "PL",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", added("PL", video("a", 100), date(2026, time.March, 1, 0)))
				// 最初と最後の動画のチャンネル
				m.names([]library.ChannelID{"", ""}, map[library.ChannelID]*library.Channel{})
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
			name: "追加された動画がない年は、一番多い月と最初・最後の動画を省く",
			arguments: map[string]any{
				"playlist_id": "PL",
				"year":        2025,
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", added("PL", video("a", 100), date(2026, time.March, 1, 0)))
			},
			want: `{
				"playlist": {"id": "PL", "title": "お気に入り", "url": "https://www.youtube.com/playlist?list=PL"},
				"year": 2025, "added": 0, "previous_year_added": 0,
				"top_viewed": []
			}`,
		},
		{
			name: "通知登録されていないプレイリストは、中身を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
		{
			name: "プレイリストの ID の形式が正しくなければ、DB を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "https://www.youtube.com/playlist?list=PL",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "playlist_id の形式が正しくありません。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "playlist_wrapped", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestRandomVideos(t *testing.T) {
	t.Parallel()

	// 選ぶ順はランダムなので、候補は 1 本だけにする
	tests := []struct {
		name      string
		arguments map[string]any
		setup     func(m *deps)
		want      string
		wantError string
	}{
		{
			name:      "観られる動画を選ぶ",
			arguments: map[string]any{},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlistHas("PL", added("PL", video("a", 100), date(2026, time.March, 1, 0)))
				m.names([]library.ChannelID{""}, map[library.ChannelID]*library.Channel{})
			},
			want: `{"videos": [{
				"video_id": "a", "title": "a", "url": "https://www.youtube.com/watch?v=a&list=PL",
				"playlist_id": "PL", "playlist_title": "お気に入り", "added_at": "2026-03-01T00:00:00+09:00",
				"views": 100, "status": "public", "tags": []
			}]}`,
		},
		{
			name: "本数が多すぎる",
			arguments: map[string]any{
				"count": 11,
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "count は 1〜10 で指定してください",
		},
		{
			name: "通知登録されていないプレイリストは、中身を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "UNWATCHED",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
			},
			want:      `null`,
			wantError: "登録されていないプレイリストです。list_playlists で ID を確認してください。",
		},
		{
			name: "プレイリストの ID の形式が正しくなければ、DB を読まずに断る",
			arguments: map[string]any{
				"playlist_id": "https://www.youtube.com/playlist?list=PL",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "playlist_id の形式が正しくありません。list_playlists で ID を確認してください。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "random_videos", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

func TestVideoHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments map[string]any
		setup     func(m *deps)
		want      string
		wantError string
	}{
		{
			name: "直近 30 日の記録を古い順に返す",
			arguments: map[string]any{
				"video_id": "dQw4w9WgXcQ",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"dQw4w9WgXcQ"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "dQw4w9WgXcQ", date(2026, time.March, 1, 0))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"dQw4w9WgXcQ"}).Return(map[library.VideoID]*library.Video{
					"dQw4w9WgXcQ": video("dQw4w9WgXcQ", 200),
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("dQw4w9WgXcQ"), now.AddDate(0, 0, -30)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "dQw4w9WgXcQ", RecordedAt: date(2026, time.October, 10, 0), Title: "dQw4w9WgXcQ", PrivacyStatus: library.PrivacyPublic, Views: 100, Likes: 10, Comments: 1},
					{VideoYoutubeID: "dQw4w9WgXcQ", RecordedAt: date(2026, time.October, 18, 0), Title: "dQw4w9WgXcQ", PrivacyStatus: library.PrivacyPublic, Views: 200, Likes: 20, Comments: 2},
				}, nil)
			},
			want: `{
				"video_id": "dQw4w9WgXcQ", "title": "dQw4w9WgXcQ", "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				"snapshots": [
					{"recorded_at": "2026-10-10T00:00:00+09:00", "title": "dQw4w9WgXcQ", "status": "public", "views": 100, "likes": 10, "comments": 1},
					{"recorded_at": "2026-10-18T00:00:00+09:00", "title": "dQw4w9WgXcQ", "status": "public", "views": 200, "likes": 20, "comments": 2}
				]
			}`,
		},
		{
			name: "日数を指定する",
			arguments: map[string]any{
				"video_id": "dQw4w9WgXcQ",
				"days":     1,
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"dQw4w9WgXcQ"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "dQw4w9WgXcQ", date(2026, time.March, 1, 0))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"dQw4w9WgXcQ"}).Return(map[library.VideoID]*library.Video{
					"dQw4w9WgXcQ": video("dQw4w9WgXcQ", 200),
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("dQw4w9WgXcQ"), now.AddDate(0, 0, -1)).Return([]*library.VideoSnapshot{}, nil)
			},
			want: `{
				"video_id": "dQw4w9WgXcQ", "title": "dQw4w9WgXcQ", "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				"snapshots": []
			}`,
		},
		{
			name: "通知登録されたプレイリストにない動画は断る",
			arguments: map[string]any{
				"video_id": "notWatched0",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"notWatched0"}).Return([]*library.Playlist{}, nil)
			},
			want:      `null`,
			wantError: "登録されているプレイリストに、その動画は見つかりません。find_videos で video_id を確認してください。",
		},
		{
			name: "動画の URL から ID を読み取る",
			arguments: map[string]any{
				"video_id": "https://youtu.be/dQw4w9WgXcQ?si=abc",
			},
			setup: func(m *deps) {
				m.watched(sub(1, "PL", "お気に入り"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"dQw4w9WgXcQ"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "1", "dQw4w9WgXcQ", date(2026, time.March, 1, 0))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"dQw4w9WgXcQ"}).Return(map[library.VideoID]*library.Video{
					"dQw4w9WgXcQ": video("dQw4w9WgXcQ", 200),
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("dQw4w9WgXcQ"), now.AddDate(0, 0, -30)).Return([]*library.VideoSnapshot{}, nil)
			},
			want: `{
				"video_id": "dQw4w9WgXcQ", "title": "dQw4w9WgXcQ", "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				"snapshots": []
			}`,
		},
		{
			name: "動画の ID とも URL とも読めない値は、DB を読まずに断る",
			arguments: map[string]any{
				"video_id": "https://www.youtube.com/playlist?list=PL",
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "video_id には動画の ID か URL を指定してください。",
		},
		{
			name: "日数が多すぎる",
			arguments: map[string]any{
				"video_id": "dQw4w9WgXcQ",
				"days":     366,
			},
			setup:     func(m *deps) {},
			want:      `null`,
			wantError: "days は 1〜365 で指定してください",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)
			session := connect(t, m)

			got, gotError := call(t, session, "video_history", tt.arguments)

			assert.JSONEq(t, tt.want, got)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}
