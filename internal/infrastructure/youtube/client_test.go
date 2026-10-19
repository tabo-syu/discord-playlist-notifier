package youtube_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/youtube"
)

func TestFindPlaylist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		responses    map[string][]response
		want         *library.PlaylistMeta
		wantErr      string
		wantRequests []string
	}{
		{
			name: "プレイリストの ID とタイトルを返す",
			responses: map[string][]response{
				"/playlists": {ok(`{"items": [{"id": "PL", "snippet": {"title": "お気に入り"}}]}`)},
			},
			want: &library.PlaylistMeta{YoutubeID: "PL", Title: "お気に入り"},
			wantRequests: []string{
				"/playlists id=PL maxResults=1 part=id,snippet",
			},
		},
		{
			name: "存在しないか非公開のプレイリストは ErrPlaylistNotFound",
			responses: map[string][]response{
				"/playlists": {ok(`{"items": []}`)},
			},
			want:    nil,
			wantErr: "playlist not found",
			wantRequests: []string{
				"/playlists id=PL maxResults=1 part=id,snippet",
			},
		},
		{
			name: "API のエラー",
			responses: map[string][]response{
				"/playlists": {failure()},
			},
			want:    nil,
			wantErr: "failed to fetch playlists: googleapi: Error 500: backend error",
			wantRequests: []string{
				"/playlists id=PL maxResults=1 part=id,snippet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, err := youtube.NewClient(yt).FindPlaylist("PL")

			assert.Equal(t, tt.wantErr, message(err))
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}

func TestFetchPlaylistMetas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ids          []library.PlaylistID
		responses    map[string][]response
		want         map[library.PlaylistID]*library.PlaylistMeta
		wantRequests []string
	}{
		{
			name: "タイトルとアイテム数を返し、存在しないプレイリストは除く",
			ids:  []library.PlaylistID{"PL1", "PL2", "GONE"},
			responses: map[string][]response{
				"/playlists": {ok(`{"items": [
					{"id": "PL1", "snippet": {"title": "お気に入り"}, "contentDetails": {"itemCount": 12}},
					{"id": "PL2"}
				]}`)},
			},
			want: map[library.PlaylistID]*library.PlaylistMeta{
				"PL1": {YoutubeID: "PL1", Title: "お気に入り", ItemCount: 12},
				// snippet と contentDetails が欠けていてもゼロ値で読む
				"PL2": {YoutubeID: "PL2"},
			},
			wantRequests: []string{
				"/playlists id=PL1,PL2,GONE maxResults=3 part=id,snippet,contentDetails",
			},
		},
		{
			name: "50 件ごとに分けて取得する",
			ids:  numbered[library.PlaylistID]("PL", 51),
			responses: map[string][]response{
				"/playlists": {
					ok(`{"items": [{"id": "PL01", "contentDetails": {"itemCount": 1}}]}`),
					ok(`{"items": [{"id": "PL51", "contentDetails": {"itemCount": 51}}]}`),
				},
			},
			want: map[library.PlaylistID]*library.PlaylistMeta{
				"PL01": {YoutubeID: "PL01", ItemCount: 1},
				"PL51": {YoutubeID: "PL51", ItemCount: 51},
			},
			wantRequests: []string{
				"/playlists id=" + joined(numbered[library.PlaylistID]("PL", 50)) + " maxResults=50 part=id,snippet,contentDetails",
				"/playlists id=PL51 maxResults=1 part=id,snippet,contentDetails",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, err := youtube.NewClient(yt).FetchPlaylistMetas(tt.ids...)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}

func TestFetchPlaylistItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		responses    map[string][]response
		want         []*library.FetchedItem
		wantErr      string
		wantRequests []string
	}{
		{
			name: "全てのページを取得する",
			responses: map[string][]response{
				"/playlistItems": {
					ok(`{"nextPageToken": "NEXT", "items": [
						{"id": "1", "snippet": {"title": "a", "publishedAt": "2026-10-18T03:04:05Z", "resourceId": {"videoId": "a"}}, "status": {"privacyStatus": "public"}}
					]}`),
					ok(`{"items": [
						{"id": "2", "snippet": {"title": "b", "publishedAt": "2026-10-18T06:00:00Z", "resourceId": {"videoId": "b"}}, "status": {"privacyStatus": "public"}}
					]}`),
				},
			},
			want: []*library.FetchedItem{
				{ItemID: "1", VideoID: "a", AddedAt: time.Date(2026, time.October, 18, 3, 4, 5, 0, time.UTC), PrivacyStatus: library.PrivacyPublic},
				{ItemID: "2", VideoID: "b", AddedAt: time.Date(2026, time.October, 18, 6, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyPublic},
			},
			wantRequests: []string{
				"/playlistItems maxResults=50 part=snippet,status playlistId=PL",
				"/playlistItems maxResults=50 pageToken=NEXT part=snippet,status playlistId=PL",
			},
		},
		{
			name: "アイテムの公開状態を読む",
			responses: map[string][]response{
				"/playlistItems": {ok(`{"items": [
					{"id": "1", "snippet": {"title": "a", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "a"}}, "status": {"privacyStatus": "unlisted"}},
					{"id": "2", "snippet": {"title": "b", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "b"}}, "status": {"privacyStatus": "private"}},
					{"id": "3", "snippet": {"title": "c", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "c"}}, "status": {"privacyStatus": "privacyStatusUnspecified"}},
					{"id": "4", "snippet": {"title": "Deleted video", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "d"}}, "status": {"privacyStatus": "public"}},
					{"id": "5", "snippet": {"title": "e", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "e"}}},
					{"id": "6", "snippet": {"title": "f", "publishedAt": "2026-10-18T00:00:00Z", "resourceId": {"videoId": "f"}}, "status": {"privacyStatus": "unknown"}}
				]}`)},
			},
			want: []*library.FetchedItem{
				{ItemID: "1", VideoID: "a", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyUnlisted},
				{ItemID: "2", VideoID: "b", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyPrivate},
				// 削除された動画は privacyStatusUnspecified か、タイトル「Deleted video」で返る
				{ItemID: "3", VideoID: "c", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyDeleted},
				{ItemID: "4", VideoID: "d", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyDeleted},
				// status が欠けているか知らない値なら公開とみなす
				{ItemID: "5", VideoID: "e", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyPublic},
				{ItemID: "6", VideoID: "f", AddedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC), PrivacyStatus: library.PrivacyPublic},
			},
			wantRequests: []string{
				"/playlistItems maxResults=50 part=snippet,status playlistId=PL",
			},
		},
		{
			name: "動画を指さないアイテムは除く",
			responses: map[string][]response{
				"/playlistItems": {ok(`{"items": [
					{"id": "1", "snippet": {"title": "a", "publishedAt": "2026-10-18T00:00:00Z"}},
					{"id": "2"}
				]}`)},
			},
			want: nil,
			wantRequests: []string{
				"/playlistItems maxResults=50 part=snippet,status playlistId=PL",
			},
		},
		{
			name: "追加日時が読めない",
			responses: map[string][]response{
				"/playlistItems": {ok(`{"items": [
					{"id": "1", "snippet": {"title": "a", "publishedAt": "yesterday", "resourceId": {"videoId": "a"}}}
				]}`)},
			},
			want:    nil,
			wantErr: `failed to parse playlist item publish time: parsing time "yesterday" as "2006-01-02T15:04:05Z": cannot parse "yesterday" as "2006"`,
			wantRequests: []string{
				"/playlistItems maxResults=50 part=snippet,status playlistId=PL",
			},
		},
		{
			name: "2 ページ目で API のエラー",
			responses: map[string][]response{
				"/playlistItems": {
					ok(`{"nextPageToken": "NEXT", "items": []}`),
					failure(),
				},
			},
			want:    nil,
			wantErr: "failed to fetch playlist items: googleapi: Error 500: backend error",
			wantRequests: []string{
				"/playlistItems maxResults=50 part=snippet,status playlistId=PL",
				"/playlistItems maxResults=50 pageToken=NEXT part=snippet,status playlistId=PL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, err := youtube.NewClient(yt).FetchPlaylistItems("PL")

			assert.Equal(t, tt.wantErr, message(err))
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}

func TestFetchVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ids          []library.VideoID
		responses    map[string][]response
		want         []*library.Video
		wantChannels []*library.Channel
		wantRequests []string
	}{
		{
			name: "動画の詳細と、投稿したチャンネルを返す",
			ids:  []library.VideoID{"a", "b"},
			responses: map[string][]response{
				"/videos": {ok(`{"items": [
					{
						"id": "a",
						"snippet": {"title": "夜に駆ける", "publishedAt": "2019-12-15T00:00:00Z", "channelId": "UC1", "channelTitle": "YOASOBI", "tags": ["J-Pop"], "categoryId": "10", "liveBroadcastContent": "none"},
						"contentDetails": {"duration": "PT4M21S"},
						"statistics": {"viewCount": "1234", "likeCount": "56", "commentCount": "7"},
						"status": {"privacyStatus": "unlisted"}
					},
					{
						"id": "b",
						"snippet": {"title": "群青", "publishedAt": "2020-09-01T00:00:00Z", "channelId": "UC1", "channelTitle": "YOASOBI"},
						"contentDetails": {"duration": "P1DT2H"}
					}
				]}`)},
			},
			want: []*library.Video{
				{YoutubeID: "a", Title: "夜に駆ける", Views: 1234, Likes: 56, Comments: 7, PrivacyStatus: library.PrivacyUnlisted, PublishedAt: time.Date(2019, time.December, 15, 0, 0, 0, 0, time.UTC), Duration: 4*time.Minute + 21*time.Second, Tags: []string{"J-Pop"}, CategoryID: "10", ChannelYoutubeID: "UC1"},
				// statistics と status が欠けていれば、再生数は 0 で公開とみなす
				{YoutubeID: "b", Title: "群青", PrivacyStatus: library.PrivacyPublic, PublishedAt: time.Date(2020, time.September, 1, 0, 0, 0, 0, time.UTC), Duration: 26 * time.Hour, ChannelYoutubeID: "UC1"},
			},
			// 同じチャンネルは 1 つにまとめる
			wantChannels: []*library.Channel{
				{YoutubeID: "UC1", Title: "YOASOBI"},
			},
			wantRequests: []string{
				"/videos id=a,b maxResults=2 part=id,snippet,contentDetails,statistics,status",
			},
		},
		{
			name: "ライブ配信中・配信予定の動画の長さは 0 にする",
			ids:  []library.VideoID{"live"},
			responses: map[string][]response{
				"/videos": {ok(`{"items": [
					{"id": "live", "snippet": {"title": "配信", "publishedAt": "2026-10-18T00:00:00Z", "liveBroadcastContent": "live"}, "contentDetails": {"duration": "PT2H"}}
				]}`)},
			},
			want: []*library.Video{
				{YoutubeID: "live", Title: "配信", PrivacyStatus: library.PrivacyPublic, PublishedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)},
			},
			wantChannels: nil,
			wantRequests: []string{
				"/videos id=live maxResults=1 part=id,snippet,contentDetails,statistics,status",
			},
		},
		{
			name: "読めない長さは 0 にする",
			ids:  []library.VideoID{"a"},
			responses: map[string][]response{
				"/videos": {ok(`{"items": [
					{"id": "a", "snippet": {"title": "a", "publishedAt": "2026-10-18T00:00:00Z"}, "contentDetails": {"duration": "4 minutes"}}
				]}`)},
			},
			want: []*library.Video{
				{YoutubeID: "a", Title: "a", PrivacyStatus: library.PrivacyPublic, PublishedAt: time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)},
			},
			wantChannels: nil,
			wantRequests: []string{
				"/videos id=a maxResults=1 part=id,snippet,contentDetails,statistics,status",
			},
		},
		{
			name: "50 件ごとに分けて取得する",
			ids:  numbered[library.VideoID]("v", 51),
			responses: map[string][]response{
				"/videos": {
					ok(`{"items": []}`),
					ok(`{"items": []}`),
				},
			},
			want:         nil,
			wantChannels: nil,
			wantRequests: []string{
				"/videos id=" + joined(numbered[library.VideoID]("v", 50)) + " maxResults=50 part=id,snippet,contentDetails,statistics,status",
				"/videos id=v51 maxResults=1 part=id,snippet,contentDetails,statistics,status",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, gotChannels, err := youtube.NewClient(yt).FetchVideos(tt.ids)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantChannels, gotChannels)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}

func TestFetchCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		responses    map[string][]response
		want         []*library.Category
		wantRequests []string
	}{
		{
			name: "日本の地域のカテゴリを日本語の名前で返す",
			responses: map[string][]response{
				"/videoCategories": {ok(`{"items": [
					{"id": "10", "snippet": {"title": "音楽", "assignable": true}},
					{"id": "18", "snippet": {"title": "ショート ムービー", "assignable": false}},
					{"id": "99"}
				]}`)},
			},
			want: []*library.Category{
				{YoutubeID: "10", Title: "音楽", Assignable: true},
				{YoutubeID: "18", Title: "ショート ムービー", Assignable: false},
			},
			wantRequests: []string{
				"/videoCategories hl=ja part=id,snippet regionCode=JP",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, err := youtube.NewClient(yt).FetchCategories()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}

func TestFetchLiveVideos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ids          []library.VideoID
		responses    map[string][]response
		want         map[library.VideoID]*library.LiveVideo
		wantRequests []string
	}{
		{
			name: "動画の現在の再生数と、チャンネルの名前とアイコンを返す",
			ids:  []library.VideoID{"a", "b", "c"},
			responses: map[string][]response{
				"/videos": {ok(`{"items": [
					{"id": "a", "snippet": {"channelId": "UC1", "channelTitle": "YOASOBI"}, "statistics": {"viewCount": "100"}},
					{"id": "b", "snippet": {"channelId": "UC1", "channelTitle": "YOASOBI"}, "statistics": {"viewCount": "200"}},
					{"id": "c", "snippet": {"channelId": "UC2", "channelTitle": "アイコンなし"}}
				]}`)},
				"/channels": {ok(`{"items": [
					{"id": "UC1", "snippet": {"thumbnails": {"default": {"url": "https://example.com/uc1.jpg"}}}},
					{"id": "UC2", "snippet": {}}
				]}`)},
			},
			want: map[library.VideoID]*library.LiveVideo{
				"a": {ChannelName: "YOASOBI", ChannelIcon: "https://example.com/uc1.jpg", Views: 100},
				"b": {ChannelName: "YOASOBI", ChannelIcon: "https://example.com/uc1.jpg", Views: 200},
				"c": {ChannelName: "アイコンなし", ChannelIcon: "", Views: 0},
			},
			// 同じチャンネルは 1 回だけ取得する
			wantRequests: []string{
				"/videos id=a,b,c maxResults=3 part=id,snippet,statistics",
				"/channels id=UC1,UC2 maxResults=2 part=id,snippet",
			},
		},
		{
			name: "同じ動画を 2 回指定しても、1 回だけ取得する",
			ids:  []library.VideoID{"a", "a"},
			responses: map[string][]response{
				"/videos": {ok(`{"items": [
					{"id": "a", "snippet": {"channelId": "UC1", "channelTitle": "YOASOBI"}, "statistics": {"viewCount": "100"}}
				]}`)},
				"/channels": {ok(`{"items": [
					{"id": "UC1", "snippet": {}}
				]}`)},
			},
			want: map[library.VideoID]*library.LiveVideo{
				"a": {ChannelName: "YOASOBI", ChannelIcon: "", Views: 100},
			},
			wantRequests: []string{
				"/videos id=a maxResults=1 part=id,snippet,statistics",
				"/channels id=UC1 maxResults=1 part=id,snippet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yt, api := serve(t, tt.responses)

			got, err := youtube.NewClient(yt).FetchLiveVideos(tt.ids)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequests, api.requests)
		})
	}
}
