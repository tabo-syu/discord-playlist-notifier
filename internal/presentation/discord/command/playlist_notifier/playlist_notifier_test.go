package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/fake"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name string
		data *discordgo.ApplicationCommandInteractionData
		want *command.Response
	}{
		{
			name: "サブコマンドがない",
			data: &discordgo.ApplicationCommandInteractionData{Name: "playlist-notifier"},
			want: command.Text("Error: No subcommand provided. Please use one of the available subcommands."),
		},
		{
			name: "最初のオプションがサブコマンドではない",
			data: &discordgo.ApplicationCommandInteractionData{
				Name: "playlist-notifier",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					stringOption("playlist-id", "PL"),
				},
			},
			want: command.Text("Error: Invalid command format. Please use one of the available subcommands."),
		},
		{
			name: "知らないサブコマンド",
			data: subcommand("unknown"),
			want: command.Text("Error: Unknown subcommand. Please use one of the available subcommands."),
		},
		{
			name: "source はリポジトリの URL を返す",
			data: subcommand("source"),
			want: command.Text("https://github.com/tabo-syu/discord-playlist-notifier"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCommand(&fake.SubscriptionRepository{}, &fake.LibraryRepository{}, &fake.YouTube{})

			got := c.Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
