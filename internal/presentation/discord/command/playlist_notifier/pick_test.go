package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestPick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		data              *discordgo.ApplicationCommandInteractionData
		want              *command.Response
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "毎日投稿する",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "daily")),
			want: command.Text("毎日 12:00 に「今日の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickDaily),
			},
		},
		{
			name: "毎週投稿する",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "weekly")),
			want: command.Text("毎週月曜の 12:00 に「今週の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickWeekly),
			},
		},
		{
			name: "毎月投稿する",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "monthly")),
			want: command.Text("毎月 1 日の 12:00 に「今月の一曲」を投稿します！\nhttps://www.youtube.com/playlist?list=PL"),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickMonthly),
			},
		},
		{
			name: "投稿をやめる",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickDaily),
			},
			data: subcommand("pick", stringOption("playlist-id", "PL"), stringOption("interval", "off")),
			want: command.Text("一曲の投稿をやめました。\nhttps://www.youtube.com/playlist?list=PL"),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
		},
		{
			name: "登録していないプレイリスト",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
			data: subcommand("pick", stringOption("playlist-id", "OTHER"), stringOption("interval", "daily")),
			want: command.Text("通知登録されていないプレイリストです。"),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
		},
		{
			name: "間隔がない",
			subscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
			data: subcommand("pick", stringOption("playlist-id", "PL")),
			want: command.Text("Error: interval is required."),
			wantSubscriptions: []*subscription.Subscription{
				picking(subscription.PickNone),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			subscriptions := &fake.SubscriptionRepository{Subscriptions: tt.subscriptions}
			c := newCommand(subscriptions, &fake.LibraryRepository{}, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}
