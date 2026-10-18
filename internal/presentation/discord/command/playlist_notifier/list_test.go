package playlist_notifier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestList(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		want          *command.Response
	}{
		{
			name: "登録したプレイリストの URL を並べる",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL1", "PL1"),
				sub(2, "PL2", "PL2"),
			},
			want: command.Text("通知登録されているプレイリスト一覧\nhttps://www.youtube.com/playlist?list=PL1\nhttps://www.youtube.com/playlist?list=PL2\n"),
		},
		{
			name: "登録したプレイリストがない",
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCommand(&fake.SubscriptionRepository{Subscriptions: tt.subscriptions}, &fake.LibraryRepository{}, &fake.YouTube{})

			got := c.Handle(subcommand("list"), "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
