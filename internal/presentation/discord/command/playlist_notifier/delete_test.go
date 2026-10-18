package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "指定したプレイリストだけ登録を解除する",
			data: subcommand("delete", stringOption("playlist-id", "PL1")),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().Exists(mockGuild, library.PlaylistID("PL1")).Return(true, nil)
				m.guildHas(sub(1, "PL1", "PL1"), sub(2, "PL2", "PL2"))
				m.subscriptions.EXPECT().DeleteAll([]*subscription.Subscription{sub(1, "PL1", "PL1")}).Return(nil)
			},
			want: command.Text("指定されたプレイリストを削除しました。"),
		},
		{
			name: "登録していないプレイリスト",
			data: subcommand("delete", stringOption("playlist-id", "PL2")),
			setup: func(m *deps) {
				m.subscriptions.EXPECT().Exists(mockGuild, library.PlaylistID("PL2")).Return(false, nil)
			},
			want: command.Text("通知登録されていないプレイリストです。"),
		},
		{
			name:  "プレイリスト ID がない",
			data:  subcommand("delete"),
			setup: func(m *deps) {},
			want:  command.Text("Error: Playlist ID is required."),
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
