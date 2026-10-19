package playlist_notifier_test

import (
	"errors"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command/playlist_notifier"
)

// TestMain はテストの出力を読みやすくするため、ログを捨てる。
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

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

// errMock は mock に失敗させるときに返すエラー。
var errMock = errors.New("mock error")

// deps はコマンドが使うサービスの、さらに先にある mock をまとめる。各テストケースの
// setup で、呼ばれるはずのメソッドと、その戻り値を指定する。
type deps struct {
	guilds        *mocks.MockGuildRepository
	subscriptions *mocks.MockSubscriptionRepository
	playlists     *mocks.MockPlaylistRepository
	videos        *mocks.MockVideoRepository
	channels      *mocks.MockChannelRepository
	categories    *mocks.MockCategoryRepository
	snapshots     *mocks.MockVideoSnapshotRepository
	youtube       *mocks.MockYouTube
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
		youtube:       mocks.NewMockYouTube(t),
	}
}

// newCommand は mock で組み立てた本物のサービスでコマンドを作る。
func newCommand(m *deps) *playlist_notifier.PlaylistNotifier {
	return playlist_notifier.NewPlaylistNotifier(
		application.NewSubscriptionService(m.youtube, m.subscriptions, m.guilds, clock),
		application.NewLibraryService(m.youtube, m.playlists, m.videos, m.channels, m.categories, m.snapshots, jst, clock),
		jst,
		clock,
	)
}

// guildHas は、ギルド「guild」の通知設定が subscriptions であることを期待する。
func (m *deps) guildHas(subscriptions ...*subscription.Subscription) {
	if len(subscriptions) == 0 {
		m.subscriptions.EXPECT().FindByGuild(mockGuild).Return(nil, subscription.ErrNotSubscribed)

		return
	}

	m.subscriptions.EXPECT().FindByGuild(mockGuild).Return(subscriptions, nil)
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

// subcommand はサブコマンド name を options 付きで呼んだときの入力を返す。
func subcommand(name string, options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionData {
	return &discordgo.ApplicationCommandInteractionData{
		Name: "playlist-notifier",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: name, Options: options},
		},
	}
}

// stringOption は文字列のオプションを返す。
func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Type: discordgo.ApplicationCommandOptionString, Name: name, Value: value}
}

// intOption は整数のオプションを返す。Discord は整数も float64 で送る。
func intOption(name string, value int) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Type: discordgo.ApplicationCommandOptionInteger, Name: name, Value: float64(value)}
}

// sub はギルド 1 の通知設定を返す。
func sub(id uint, playlistID library.PlaylistID, title string) *subscription.Subscription {
	return &subscription.Subscription{
		ID: id, NotifiedUntil: hoursAgo(6),
		YoutubeID:     playlistID,
		SendChannelID: "ch",
		Title:         title,
		GuildID:       1,
	}
}

// playlist は items を持つプレイリストを返す。
func playlist(playlistID library.PlaylistID, items ...*library.PlaylistItem) *library.Playlist {
	return &library.Playlist{YoutubeID: playlistID, Items: items}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: library.ItemID(string(playlistID) + string(videoID)), VideoYoutubeID: videoID, AddedAt: addedAt}
}

// added はプレイリストに addedAt に追加された動画を返す。
func added(playlistID library.PlaylistID, v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{PlaylistID: playlistID, Item: item(playlistID, v.YoutubeID, addedAt), Video: v}
}

// video は保存した公開の動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID) *library.Video {
	return &library.Video{YoutubeID: id, DetailsFetched: true, Title: string(id), PrivacyStatus: library.PrivacyPublic}
}

// picking は sub(1, "PL", "PL") のピックの間隔を interval にしたものを返す。
func picking(interval subscription.PickInterval) *subscription.Subscription {
	return &subscription.Subscription{
		ID: 1, NotifiedUntil: hoursAgo(6),
		YoutubeID:     "PL",
		SendChannelID: "ch",
		Title:         "PL",
		PickInterval:  interval,
		GuildID:       1,
	}
}

// mockGuild はテストのギルドの ID。
var mockGuild = guild.DiscordID("guild")

// daysAgo は now の days 日前の時刻を返す。
func daysAgo(days int) time.Time {
	return now.AddDate(0, 0, -days)
}
