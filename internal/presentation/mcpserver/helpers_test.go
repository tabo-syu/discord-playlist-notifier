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
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
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

// connect は偽物のリポジトリで MCP サーバーを HTTP で立て、クライアントとして接続する。
func connect(t *testing.T, subscriptions *fake.SubscriptionRepository, lib *fake.LibraryRepository) *mcp.ClientSession {
	t.Helper()

	query := application.NewQueryService(subscriptions, application.NewLibraryService(&fake.YouTube{}, lib, jst, clock), lib)
	ts := httptest.NewServer(mcpserver.NewServer("", query, jst, clock).Handler())
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL + mcpserver.ENDPOINT_PATH}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	return session
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

	return string(output), res.Content[0].(*mcp.TextContent).Text
}

// sub はギルド 1 の通知設定を返す。
func sub(id uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	return &subscription.Subscription{ID: id, YoutubeID: playlistID, SendChannelID: "ch", Title: title, GuildID: 1}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// video は保存した公開の動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: library.PrivacyPublic, Views: views}
}
