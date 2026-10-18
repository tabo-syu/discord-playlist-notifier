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

func TestRandom(t *testing.T) {
	t.Parallel()

	// 選ぶ順はランダムなので、候補が 1 曲だけの中身を使う
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		storedItems   []*library.PlaylistItem
		storedVideos  []*library.Video
		data          *discordgo.ApplicationCommandInteractionData
		want          *command.Response
	}{
		{
			name: "登録したプレイリストから選ぶ",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "PL"),
			},
			storedItems: []*library.PlaylistItem{
				item("PL", "aaaaaaaaaaa", hoursAgo(1)),
			},
			storedVideos: []*library.Video{
				video("aaaaaaaaaaa"),
			},
			data: subcommand("random", intOption("count", 3)),
			want: command.Text("🎲 ランダムに選びました！\n\n**aaaaaaaaaaa**\nhttps://www.youtube.com/watch?v=aaaaaaaaaaa&list=PL\n"),
		},
		{
			name: "選べる曲がない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "PL"),
			},
			data: subcommand("random"),
			want: command.Text("選べる曲がまだありません。少し待ってからもう一度試してください。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("random"),
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name: "曲数が 0",
			data: subcommand("random", intOption("count", 0)),
			want: command.Text("Error: count must be between 1 and 5."),
		},
		{
			name: "曲数が 6",
			data: subcommand("random", intOption("count", 6)),
			want: command.Text("Error: count must be between 1 and 5."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lib := &fake.LibraryRepository{Items: tt.storedItems, Videos: tt.storedVideos}
			c := newCommand(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, lib, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
