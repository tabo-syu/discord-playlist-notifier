package playlist_notifier_test

import (
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
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

// newCommand は、ギルド「guild」（ID 1）だけがある状態のコマンドを、偽物のリポジトリと YouTube で組み立てる。
func newCommand(subscriptions *fake.SubscriptionRepository, lib *fake.LibraryRepository, yt *fake.YouTube) *playlist_notifier.PlaylistNotifier {
	guilds := &fake.GuildRepository{Guilds: []*guild.Guild{{Model: gorm.Model{ID: 1}, DiscordID: "guild"}}}
	subscriptions.Guilds = guilds

	return playlist_notifier.NewPlaylistNotifier(
		application.NewSubscriptionService(yt, subscriptions, guilds),
		application.NewLibraryService(yt, lib, jst, clock),
		jst,
		clock,
	)
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
		Model:         gorm.Model{ID: id, UpdatedAt: hoursAgo(6)},
		YoutubeID:     playlistID,
		SendChannelID: "ch",
		Title:         title,
		GuildID:       1,
	}
}

// item は保存したプレイリストのアイテムを返す。
func item(playlistID library.PlaylistID, videoID library.VideoID, addedAt time.Time) *library.PlaylistItem {
	return &library.PlaylistItem{PlaylistYoutubeID: playlistID, ItemID: library.ItemID(string(playlistID) + string(videoID)), VideoYoutubeID: videoID, AddedAt: addedAt}
}

// video は保存した公開の動画を返す。タイトルは ID と同じにする。
func video(id library.VideoID) *library.Video {
	return &library.Video{YoutubeID: id, Title: string(id), PrivacyStatus: library.PrivacyPublic}
}

// picking は sub(1, "PL", "PL") のピックの間隔を interval にしたものを返す。
func picking(interval subscription.PickInterval) *subscription.Subscription {
	return &subscription.Subscription{
		Model:         gorm.Model{ID: 1, UpdatedAt: hoursAgo(6)},
		YoutubeID:     "PL",
		SendChannelID: "ch",
		Title:         "PL",
		PickInterval:  interval,
		GuildID:       1,
	}
}
