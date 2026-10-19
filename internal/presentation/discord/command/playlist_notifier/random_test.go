package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestRandom(t *testing.T) {
	t.Parallel()

	// 選ぶ順はランダムなので、候補が 1 本だけの中身を使う
	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "登録したプレイリストから選ぶ",
			data: subcommand("random", intOption("count", 3)),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "PL"))
				m.playlistHas("PL", added("PL", video("aaaaaaaaaaa"), hoursAgo(1)))
			},
			want: command.Text("🎲 ランダムに選びました！\n\n**aaaaaaaaaaa**\nhttps://www.youtube.com/watch?v=aaaaaaaaaaa&list=PL\n"),
		},
		{
			name: "選べる動画がない",
			data: subcommand("random"),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "PL"))
				m.playlistHas("PL")
			},
			want: command.Text("選べる動画がまだありません。少し待ってからもう一度試してください。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("random"),
			setup: func(m *deps) {
				m.guildHas()
			},
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name:  "本数が 0",
			data:  subcommand("random", intOption("count", 0)),
			setup: func(m *deps) {},
			want:  command.Text("Error: count must be between 1 and 5."),
		},
		{
			name:  "本数が 6",
			data:  subcommand("random", intOption("count", 6)),
			setup: func(m *deps) {},
			want:  command.Text("Error: count must be between 1 and 5."),
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
