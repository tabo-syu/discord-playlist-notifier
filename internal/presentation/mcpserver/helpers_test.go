package mcpserver_test

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/mcpserver"
)

// TestMain はテストの出力を読みやすくするため、ログを捨てる。
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// now はテストの現在時刻（JST では 2026-10-18 21:00）。
var now = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

func clock() time.Time {
	return now
}

// date は JST の日時を返す。
func date(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, jst)
}

// deps は QueryService が使う mock をまとめる。各テストケースの setup で、呼ばれる
// はずのメソッドと、その戻り値を指定する。
type deps struct {
	subscriptions *mocks.MockSubscriptionRepository
	playlists     *mocks.MockPlaylistRepository
	videos        *mocks.MockVideoRepository
	channels      *mocks.MockChannelRepository
	categories    *mocks.MockCategoryRepository
	snapshots     *mocks.MockVideoSnapshotRepository
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		subscriptions: mocks.NewMockSubscriptionRepository(t),
		playlists:     mocks.NewMockPlaylistRepository(t),
		videos:        mocks.NewMockVideoRepository(t),
		channels:      mocks.NewMockChannelRepository(t),
		categories:    mocks.NewMockCategoryRepository(t),
		snapshots:     mocks.NewMockVideoSnapshotRepository(t),
	}
}

// connect は mock で組み立てた MCP サーバーを HTTP で立て、クライアントとして接続する。
func connect(t *testing.T, m *deps) *mcp.ClientSession {
	t.Helper()

	library := application.NewLibraryService(mocks.NewMockYouTube(t), m.playlists, m.videos, m.channels, m.categories, m.snapshots, jst, clock)
	query := application.NewQueryService(m.subscriptions, library, m.categories, m.channels)
	ts := httptest.NewServer(mcpserver.NewServer("", query, jst, clock).Handler())
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL + mcpserver.ENDPOINT_PATH}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	return session
}

// watched は、通知登録されたプレイリストが subscriptions であることを期待する。何度呼ばれてもよい。
func (m *deps) watched(subscriptions ...*subscription.Subscription) {
	m.subscriptions.EXPECT().FindAll().Return(subscriptions, nil)
}

// playlistHas は、プレイリストの保存済みの中身が contents であることを期待する。
func (m *deps) playlistHas(playlistID library.PlaylistID, contents ...*library.PlaylistVideo) {
	items := make([]*library.PlaylistItem, 0, len(contents))
	videos := map[library.VideoID]*library.Video{}

	for _, c := range contents {
		items = append(items, c.Item)
		videos[c.Video.YoutubeID] = c.Video
	}

	m.playlists.EXPECT().Find(playlistID).Return(playlist(playlistID, items...), nil)
	m.videos.EXPECT().FindInPlaylists(playlistID).Return(videos, nil)
}

// names は、動画を出力するときに引くチャンネルとカテゴリを期待する。channelIDs は
// 出力する動画のチャンネル（わからなければ空文字列）を、出力する順に並べたもの。
func (m *deps) names(channelIDs []library.ChannelID, channels map[library.ChannelID]*library.Channel) {
	m.channels.EXPECT().Find(channelIDs).Return(channels, nil)
	m.categories.EXPECT().FindAll().Return(categories(), nil)
}

// categories は保存済みのカテゴリを返す。
func categories() map[library.CategoryID]*library.Category {
	return map[library.CategoryID]*library.Category{
		"10": {YoutubeID: "10", Title: "音楽", Assignable: true},
		"20": {YoutubeID: "20", Title: "ゲーム", Assignable: true},
		"99": {YoutubeID: "99", Title: "古いカテゴリ", Assignable: false},
	}
}

// added はプレイリストに addedAt に追加された動画を返す。
func added(playlistID library.PlaylistID, v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{PlaylistID: playlistID, Item: item(playlistID, "1", v.YoutubeID, addedAt), Video: v}
}

// call はツールを呼び、構造化された出力の JSON と、ツールが返したエラーの文言を返す。
// エラーのときの出力は null、成功したときのエラーの文言は空になる。
func call(t *testing.T, session *mcp.ClientSession, name string, arguments map[string]any) (string, string) {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	require.NoError(t, err)

	output, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)

	if !res.IsError {
		return string(output), ""
	}

	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	return string(output), text.Text
}

// sub はギルド 1 の通知設定を返す。
func sub(id uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	return &subscription.Subscription{ID: id, YoutubeID: playlistID, SendChannelID: "ch", Title: title, GuildID: 1}
}

// playlist は items を持つプレイリストを返す。
func playlist(playlistID library.PlaylistID, items ...*library.PlaylistItem) *library.Playlist {
	return &library.Playlist{YoutubeID: playlistID, Items: items}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// video は保存した公開の動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: library.PrivacyPublic, Views: views}
}
