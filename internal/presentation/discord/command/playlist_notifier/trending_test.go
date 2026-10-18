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
func TestTrending(t *testing.T) {
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
				sub(1, "PL", "お気に入り"),
			},
			data: subcommand("trending"),
			want: &command.Response{
				Content: "🔥 **お気に入り** の急上昇\nまだ記録が足りません。再生数は 6 時間ごとに記録しています。\n",
			},
		},
		{
			name: "期間内に伸びた曲がない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "お気に入り"),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "a", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("a"),
			},
			storedSnapshot: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(48), Views: 100},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100},
			},
			data: subcommand("trending"),
			want: &command.Response{
				Content: "🔥 **お気に入り** の急上昇\n10/16〜10/17 の伸び\nこの期間に再生数が伸びた曲はありません。\n",
			},
		},
		{
			name: "指定した日数より前の記録は使わない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "お気に入り"),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "a", hoursAgo(100)),
			},
			storedVideos: []*library.Video{
				video("a"),
			},
			storedSnapshot: []*library.VideoSnapshot{
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(48), Views: 100},
				{VideoYoutubeID: "a", RecordedAt: hoursAgo(12), Views: 200},
			},
			data: subcommand("trending", intOption("days", 1)),
			want: &command.Response{
				Content: "🔥 **お気に入り** の急上昇\nまだ記録が足りません。再生数は 6 時間ごとに記録しています。\n",
			},
		},
		{
			name: "登録していないプレイリストを指定した",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "お気に入り"),
			},
			data: subcommand("trending", stringOption("playlist-id", "OTHER")),
			want: command.Text("通知登録されていないプレイリストです。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("trending"),
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name: "日数が 0",
			data: subcommand("trending", intOption("days", 0)),
			want: command.Text("Error: days must be between 1 and 365."),
		},
		{
			name: "日数が 366",
			data: subcommand("trending", intOption("days", 366)),
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
