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
	playlists     *mocks.MockPlaylistRepository
	videos        *mocks.MockVideoRepository
	channels      *mocks.MockChannelRepository
	categories    *mocks.MockCategoryRepository
	snapshots     *mocks.MockVideoSnapshotRepository
	source        *mocks.MockLibrarySource
	live          *mocks.MockLiveVideoSource
	notifier      *mocks.MockNotifier
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		guilds:        mocks.NewMockGuildRepository(t),
		subscriptions: mocks.NewMockSubscriptionRepository(t),
		playlists:     mocks.NewMockPlaylistRepository(t),
		videos:        mocks.NewMockVideoRepository(t),
		channels:      mocks.NewMockChannelRepository(t),
		categories:    mocks.NewMockCategoryRepository(t),
		snapshots:     mocks.NewMockVideoSnapshotRepository(t),
		source:        mocks.NewMockLibrarySource(t),
		live:          mocks.NewMockLiveVideoSource(t),
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
		ID: id, NotifiedUntil: notifiedAt,
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
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: status, Views: views, ViewMilestone: &milestone}
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

// playlist は items を持つプレイリストを返す。
func playlist(playlistID library.PlaylistID, items ...*library.PlaylistItem) *library.Playlist {
	return &library.Playlist{YoutubeID: playlistID, Items: items}
}

// expectPick は、プレイリストの唯一の動画「a」を s のピックとして投稿するときの呼び出しを期待する。
func (m *deps) expectPick(s *subscription.Subscription) {
	m.playlists.EXPECT().Find(s.YoutubeID).Return(playlist(s.YoutubeID,
		item(s.YoutubeID, "1", "a", hoursAgo(10)),
	), nil)
	m.videos.EXPECT().FindInPlaylists(s.YoutubeID).Return(map[library.VideoID]*library.Video{
		"a": video("a", library.PrivacyPublic, 0, 0),
	}, nil)
	m.live.EXPECT().FetchLiveVideos([]library.VideoID{"a"}).Return(map[library.VideoID]*application.LiveVideo{}, nil)
	m.notifier.EXPECT().PostPick(s, &library.PlaylistVideo{
		PlaylistID: s.YoutubeID,
		Item:       item(s.YoutubeID, "1", "a", hoursAgo(10)),
		Video:      video("a", library.PrivacyPublic, 0, 0),
	}, map[library.VideoID]*application.LiveVideo{}).Return(nil)
}

// synced は、items を持つプレイリスト playlistID を同期し、YouTube 上のタイトルが title
// だったイベントを返す。videos はそのアイテムの動画。
func synced(playlistID library.PlaylistID, title string, items []*library.PlaylistItem, videos map[library.VideoID]*library.Video) *library.PlaylistSynced {
	return &library.PlaylistSynced{
		Snapshot: &library.Snapshot{Playlist: playlist(playlistID, items...), Title: title, Videos: videos},
	}
}

// published は、テストの EventPublisher に発行されたイベントを種類ごとに記録する。
type published struct {
	hidden  []*library.VideoHidden
	synced  []*library.PlaylistSynced
	reached []*library.MilestoneReached
}

// record は、p に発行されたイベントを記録するハンドラを登録し、その記録を返す。
func record(p *application.EventPublisher) *published {
	got := &published{}

	application.Subscribe(p, func(events []*library.VideoHidden) { got.hidden = append(got.hidden, events...) })
	application.Subscribe(p, func(events []*library.PlaylistSynced) { got.synced = append(got.synced, events...) })
	application.Subscribe(p, func(events []*library.MilestoneReached) { got.reached = append(got.reached, events...) })

	return got
}

// newQueryService は mock で QueryService を組み立てる。
func newQueryService(m *deps) *application.QueryService {
	return application.NewQueryService(m.subscriptions, library.NewCatalog(m.playlists, m.videos), m.snapshots, m.categories, m.channels, jst, clock)
}

// newNotificationService は mock で NotificationService を組み立てる。cmd/bot/main.go と
// 同じく、新着の通知のハンドラに VideosAdded を購読させる。now は現在時刻を返す。
func newNotificationService(m *deps, now func() time.Time) *application.NotificationService {
	events := application.NewEventPublisher()
	application.Subscribe(events, newNotificationHandlers(m).NotifyNewVideos)

	return application.NewNotificationService(m.subscriptions, library.NewCatalog(m.playlists, m.videos), m.live, m.notifier, events, jst, now)
}

// newNotificationHandlers は mock で NotificationHandlers を組み立てる。
func newNotificationHandlers(m *deps) *application.NotificationHandlers {
	return application.NewNotificationHandlers(m.subscriptions, library.NewCatalog(m.playlists, m.videos), m.live, m.notifier)
}

// newLibraryService は mock で LibraryService を組み立てる。events は同期で起きたことを
// 発行する先で、now は現在時刻を返す。
func newLibraryService(m *deps, events *application.EventPublisher, now func() time.Time) *application.LibraryService {
	return application.NewLibraryService(
		m.source,
		m.subscriptions,
		m.playlists,
		m.videos,
		m.channels,
		m.categories,
		m.snapshots,
		library.NewCatalog(m.playlists, m.videos),
		events,
		jst,
		now,
	)
}
