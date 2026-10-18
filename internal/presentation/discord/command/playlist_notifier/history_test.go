package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// グラフの画像は比べられないので、グラフを添付しない場合だけを確かめる。
func TestHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subscriptions  []*subscription.Subscription
		storedItems    []*library.PlaylistItem
		storedVideos   []*library.Video
		storedSnapshot []*library.VideoSnapshot
		data           *discordgo.ApplicationCommandInteractionData
		want           *command.Response
	}{
		{
			name: "記録が足りない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "PL"),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "aaaaaaaaaaa", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("aaaaaaaaaaa"),
			},
			storedSnapshot: []*library.VideoSnapshot{
				{VideoYoutubeID: "aaaaaaaaaaa", RecordedAt: hoursAgo(24), Views: 100},
			},
			data: subcommand("history", stringOption("video", "https://youtu.be/aaaaaaaaaaa")),
			want: command.Text("📈 **aaaaaaaaaaa** の推移（直近 30 日）\nまだ記録が足りません。再生数などは 6 時間ごとに記録しています。\nhttps://www.youtube.com/watch?v=aaaaaaaaaaa"),
		},
		{
			name: "登録したプレイリストにない動画",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "PL"),
			},
			storedItems: []*library.PlaylistItem{
				item("OTHER", "aaaaaaaaaaa", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("aaaaaaaaaaa"),
			},
			data: subcommand("history", stringOption("video", "aaaaaaaaaaa")),
			want: command.Text("通知登録されているプレイリストに、その動画は見つかりませんでした。"),
		},
		{
			name: "動画の URL でも ID でもない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "PL"),
			},
			data: subcommand("history", stringOption("video", "https://example.com/")),
			want: command.Text("動画の URL か ID を入力してください。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("history", stringOption("video", "aaaaaaaaaaa")),
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name: "動画がない",
			data: subcommand("history"),
			want: command.Text("Error: Video is required."),
		},
		{
			name: "日数が 366",
			data: subcommand("history", stringOption("video", "aaaaaaaaaaa"), intOption("days", 366)),
			want: command.Text("Error: days must be between 1 and 365."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos, Snapshots: tt.storedSnapshot}
			c := newCommand(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, lib, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
