package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// 集計した中身の整形は view のゴールデンテストで確かめるので、ここでは中身のない
// プレイリストで、どのプレイリストを表示するかだけを確かめる。
func TestSummary(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		data          *discordgo.ApplicationCommandInteractionData
		want          *command.Response
	}{
		{
			name: "登録した全てのプレイリストを表示する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "お気に入り"),
				sub(2, "PL2", "作業用"),
			},
			data: subcommand("summary"),
			want: command.Text("📊 **お気に入り**\nまだ曲の情報がありません。少し待ってからもう一度試してください。\n\n📊 **作業用**\nまだ曲の情報がありません。少し待ってからもう一度試してください。\n"),
		},
		{
			name: "指定したプレイリストだけを表示する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "お気に入り"),
				sub(2, "PL2", "作業用"),
			},
			data: subcommand("summary", stringOption("playlist-id", "PL2")),
			want: command.Text("📊 **作業用**\nまだ曲の情報がありません。少し待ってからもう一度試してください。\n"),
		},
		{
			name: "登録していないプレイリストを指定した",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "お気に入り"),
			},
			data: subcommand("summary", stringOption("playlist-id", "OTHER")),
			want: command.Text("通知登録されていないプレイリストです。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("summary"),
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCommand(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, &fake.LibraryRepository{}, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
