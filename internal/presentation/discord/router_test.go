package discord_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/mocks"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// newCommand は name という名前で、呼ばれると name を本文にして返すコマンドの mock を返す。
func newCommand(t *testing.T, name string) *mocks.MockCommand {
	t.Helper()

	c := mocks.NewMockCommand(t)
	c.EXPECT().GetCommand().Return(&discordgo.ApplicationCommand{Name: name})
	c.EXPECT().Handle(&discordgo.ApplicationCommandInteractionData{Name: name}, "guild", "ch").Return(command.Text(name)).Maybe()

	return c
}

func TestRouterRoute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		commandName string
		want        *command.Response
	}{
		{
			name:        "コマンド名に対応するコマンドで処理する",
			commandName: "b",
			want:        command.Text("b"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := discord.NewRouter([]command.Command{newCommand(t, "a"), newCommand(t, "b")})

			got := r.Route(tt.commandName)(&discordgo.ApplicationCommandInteractionData{Name: tt.commandName}, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRouterRouteUnknown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "知らないコマンド名には何も返さない",
			commandName: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := discord.NewRouter([]command.Command{newCommand(t, "a")})

			got := r.Route(tt.commandName)

			assert.Nil(t, got)
		})
	}
}
