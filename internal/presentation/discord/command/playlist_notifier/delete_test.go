package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		data              *discordgo.ApplicationCommandInteractionData
		want              *command.Response
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "プレイリストの登録を解除する",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
				sub(2, "PL2", "PL2"),
			},
			data: subcommand("delete", stringOption("playlist-id", "PL1")),
			want: command.Text("指定されたプレイリストを削除しました。"),
			wantSubscriptions: []*subscription.Subscription{
				sub(2, "PL2", "PL2"),
			},
		},
		{
			name: "登録していないプレイリスト",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
			},
			data: subcommand("delete", stringOption("playlist-id", "PL2")),
			want: command.Text("通知登録されていないプレイリストです。"),
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
			},
		},
		{
			name: "プレイリスト ID がない",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
			},
			data: subcommand("delete"),
			want: command.Text("Error: Playlist ID is required."),
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
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
