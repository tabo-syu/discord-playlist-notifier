package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestPick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "毎日投稿する",
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "daily")),
			setup: func(m *deps) {
				m.guildHas(picking(subscription.PickNone))
				m.subscriptions.EXPECT().UpdatePickInterval(picking(subscription.PickDaily)).Return(nil)
			},
			want: command.Text("毎日 12:00 に「今日の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "毎週投稿する",
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "weekly")),
			setup: func(m *deps) {
				m.guildHas(picking(subscription.PickNone))
				m.subscriptions.EXPECT().UpdatePickInterval(picking(subscription.PickWeekly)).Return(nil)
			},
			want: command.Text("毎週月曜の 12:00 に「今週の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "毎月投稿する",
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "monthly")),
			setup: func(m *deps) {
				m.guildHas(picking(subscription.PickNone))
				m.subscriptions.EXPECT().UpdatePickInterval(picking(subscription.PickMonthly)).Return(nil)
			},
			want: command.Text("毎月 1 日の 12:00 に「今月の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "投稿をやめる",
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "off")),
			setup: func(m *deps) {
				m.guildHas(picking(subscription.PickDaily))
				m.subscriptions.EXPECT().UpdatePickInterval(picking(subscription.PickNone)).Return(nil)
			},
			want: command.Text("一曲の投稿をやめました。\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "登録していないプレイリスト",
			data: subcommand("pick", stringOption("playlist-id", "OTHER"), stringOption("interval", "daily")),
			setup: func(m *deps) {
				m.guildHas(picking(subscription.PickNone))
			},
			want: command.Text("通知登録されていないプレイリストです。"),
		},
		{
			name:  "間隔がない",
			data:  subcommand("pick", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {},
			want:  command.Text("Error: interval is required."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)

			got := newCommand(m).Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
