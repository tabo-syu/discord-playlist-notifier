package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		subscriptions     []*subscription.Subscription
		youtube           *fake.YouTube
		data              *discordgo.ApplicationCommandInteractionData
		want              *command.Response
		wantSubscriptions []*subscription.Subscription
	}{
		{
			name: "プレイリストを登録する",
			youtube: &fake.YouTube{
				Playlists: []*fake.Playlist{
					{ID: "PL", Title: "タイトル"},
				},
			},
			data: subcommand("add", stringOption("playlist-id", "PL")),
			want: command.Text("通知登録しました！\nhttps://www.youtube.com/playlist?list=PL"),
			wantSubscriptions: []*subscription.Subscription{
				{ID: 1, YoutubeID: "PL", SendChannelID: "ch", Title: "タイトル", GuildID: 1},
			},
		},
		{
			name: "登録済みのプレイリスト",
			subscriptions: []*subscription.Subscription{
				sub(1, "PL", "タイトル"),
			},
			youtube: &fake.YouTube{
				Playlists: []*fake.Playlist{
					{ID: "PL", Title: "タイトル"},
				},
			},
			data: subcommand("add", stringOption("playlist-id", "PL")),
			want: command.Text("既に通知登録されているプレイリストです。"),
			wantSubscriptions: []*subscription.Subscription{
				sub(1, "PL", "タイトル"),
			},
		},
		{
			name:    "YouTube にないプレイリスト",
			youtube: &fake.YouTube{},
			data:    subcommand("add", stringOption("playlist-id", "PL")),
			want:    command.Text("該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"),
		},
		{
			name: "YouTube API が失敗した",
			youtube: &fake.YouTube{
				FindFails: true,
			},
			data: subcommand("add", stringOption("playlist-id", "PL")),
			want: command.Text("YouTube API のサービス状況を確認してください。"),
		},
		{
			name:    "プレイリスト ID がない",
			youtube: &fake.YouTube{},
			data:    subcommand("add"),
			want:    command.Text("Error: Playlist ID is required."),
		},
		{
			name:    "プレイリスト ID が空",
			youtube: &fake.YouTube{},
			data:    subcommand("add", stringOption("playlist-id", "")),
			want:    command.Text("Error: Playlist ID cannot be empty."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			subscriptions := &fake.SubscriptionRepository{Subscriptions: tt.subscriptions}
			c := newCommand(subscriptions, &fake.LibraryRepository{}, tt.youtube)

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantSubscriptions, subscriptions.Subscriptions)
		})
	}
}
