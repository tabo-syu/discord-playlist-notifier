package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// まとめの中身の整形は view のゴールデンテストで確かめるので、ここでは中身のない
// プレイリストで、どの年をまとめるかだけを確かめる。
func TestWrapped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		data          *discordgo.ApplicationCommandInteractionData
		want          *command.Response
	}{
		{
			name: "年を省略すると今年をまとめる",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "お気に入り"),
				sub(2, "PL2", "作業用"),
			},
			data: subcommand("wrapped"),
			want: command.Text("**🎁 お気に入り の 2026 年まとめ**\n2026 年に追加された曲はありません。\n\n**🎁 作業用 の 2026 年まとめ**\n2026 年に追加された曲はありません。\n"),
		},
		{
			name: "指定した年をまとめる",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "お気に入り"),
			},
			data: subcommand("wrapped", intOption("year", 2024)),
			want: command.Text("**🎁 お気に入り の 2024 年まとめ**\n2024 年に追加された曲はありません。\n"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("wrapped"),
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := newCommand(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, &fake.LibraryRepository{}, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
