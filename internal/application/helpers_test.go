package application_test

import (
	"errors"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
)

// TestMain はテストの出力を読みやすくするため、サービスのログを捨てる。
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

// errMock は mock に失敗させるときに返すエラー。
var errMock = errors.New("mock error")

// deps はサービスが使う mock をまとめる。各テストケースの setup で、呼ばれるはずの
// メソッドと、その戻り値を指定する。指定した呼び出しがなければテストは失敗する。
type deps struct {
	guilds        *mocks.MockGuildRepository
	subscriptions *mocks.MockSubscriptionRepository
	library       *mocks.MockLibraryRepository
	youtube       *mocks.MockYouTube
	notifier      *mocks.MockNotifier
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		guilds:        mocks.NewMockGuildRepository(t),
		subscriptions: mocks.NewMockSubscriptionRepository(t),
		library:       mocks.NewMockLibraryRepository(t),
		youtube:       mocks.NewMockYouTube(t),
		notifier:      mocks.NewMockNotifier(t),
	}
}

// sameElements は、順序を問わず want と同じ要素を持つ引数に一致する。マップから
// 組み立てるので順序の決まらない引数に使う。
func sameElements[T any](want []T) any {
	return mock.MatchedBy(func(got []T) bool {
		return assert.ElementsMatch(discard{}, want, got)
	})
}

// discard は失敗の報告を捨てる assert.TestingT。
type discard struct{}

func (discard) Errorf(string, ...any) {}

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// now はテストの現在時刻。
var now = time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)

func clock() time.Time {
	return now
}

// hoursAgo は now の hours 時間前の時刻を返す。
func hoursAgo(hours int) time.Time {
	return now.Add(-time.Duration(hours) * time.Hour)
}

// lastNotified は前回の通知の時刻。
var lastNotified = hoursAgo(6)

// guildOf は ID が id のギルドを返す。
func guildOf(id uint, discordID guild.DiscordID) *guild.Guild {
	return &guild.Guild{ID: id, DiscordID: discordID}
}

// sub はギルド 1 の通知設定を返す。notifiedAt は新着動画の通知の基準時刻。
func sub(id uint, playlistID library.PlaylistID, channelID subscription.ChannelID, title string, notifiedAt time.Time) *subscription.Subscription {
	return &subscription.Subscription{
		ID: id, UpdatedAt: notifiedAt,
		YoutubeID:     playlistID,
		SendChannelID: channelID,
		Title:         title,
		GuildID:       1,
	}
}

// fetched は YouTube 上のプレイリストのアイテムを返す。
func fetched(itemID library.ItemID, videoID library.VideoID, addedAt time.Time, status library.PrivacyStatus) *library.FetchedItem {
	return &library.FetchedItem{ItemID: itemID, VideoID: videoID, AddedAt: addedAt, PrivacyStatus: status}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, itemID library.ItemID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: itemID, VideoYoutubeID: videoID, AddedAt: addedAt}
}

// video は保存した動画を返す。タイトルは ID と同じにし、再生数の節目は milestone まで記録済みにする。
func video(id library.VideoID, status library.PrivacyStatus, views, milestone library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: status, Views: views, ViewMilestone: &milestone}
}

// picking はピックの間隔を interval にした、ギルド 1 の通知設定を返す。
func picking(id uint, playlistID library.PlaylistID, interval subscription.PickInterval) *subscription.Subscription {
	s := sub(id, playlistID, "ch", string(playlistID), lastNotified)
	s.ChangePickInterval(interval)

	return s
}

// date は JST の日時を返す。
func date(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, jst)
}

// expectUnchangedSync は、保存済みの中身から何も変わっていないプレイリストを初めて同期する
// ときの呼び出しを期待する。YouTube は items と同じアイテムを、videos の公開状態で返す。
// FetchPlaylistMetas は全てのプレイリストをまとめて取得するので、ここでは期待しない。
func (m *deps) expectUnchangedSync(playlistID library.PlaylistID, items []*library.PlaylistItem, videos map[library.VideoID]*library.Video) {
	fetchedItems := make([]*library.FetchedItem, 0, len(items))
	for _, item := range items {
		fetchedItems = append(fetchedItems, fetched(item.ItemID, item.VideoYoutubeID, item.AddedAt, videos[item.VideoYoutubeID].PrivacyStatus))
	}

	m.library.EXPECT().FindItems(playlistID).Return(items, nil).Times(2)
	m.youtube.EXPECT().FetchPlaylistItems(playlistID).Return(fetchedItems, nil)
	m.library.EXPECT().FindVideosInPlaylists(playlistID).Return(videos, nil)
	m.library.EXPECT().FindVideos([]library.VideoID(nil)).Return(map[library.VideoID]*library.Video{}, nil)
	m.library.EXPECT().ApplyItems(playlistID, []*library.PlaylistItem(nil), []*library.PlaylistItem(nil), []*library.Video(nil)).Return(nil)
}

// meta は YouTube から取得したプレイリストの情報を返す。
func meta(playlistID library.PlaylistID, title string, itemCount int64) *library.PlaylistMeta {
	return &library.PlaylistMeta{YoutubeID: playlistID, Title: title, ItemCount: itemCount}
}

// expectPick は、プレイリストの唯一の曲「a」を s のピックとして投稿するときの呼び出しを期待する。
func (m *deps) expectPick(s *subscription.Subscription) {
	m.library.EXPECT().FindItems(s.YoutubeID).Return([]*library.PlaylistItem{
		item(s.YoutubeID, "1", "a", hoursAgo(10)),
	}, nil)
	m.library.EXPECT().FindVideosInPlaylists(s.YoutubeID).Return(map[library.VideoID]*library.Video{
		"a": video("a", library.PrivacyPublic, 0, 0),
	}, nil)
	m.youtube.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(map[library.VideoID]*library.LiveVideo{}, nil)
	m.notifier.EXPECT().PostPick(s, &library.PlaylistVideo{
		PlaylistID: s.YoutubeID,
		Item:       item(s.YoutubeID, "1", "a", hoursAgo(10)),
		Video:      video("a", library.PrivacyPublic, 0, 0),
	}, map[library.VideoID]*library.LiveVideo{}).Return(nil)
}

// newQueryService は mock で QueryService を組み立てる。
func newQueryService(m *deps) *application.QueryService {
	return application.NewQueryService(m.subscriptions, application.NewLibraryService(m.youtube, m.library, jst, clock), m.library)
}
