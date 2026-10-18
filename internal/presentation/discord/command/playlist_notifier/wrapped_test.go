package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// まとめの中身の整形は view のゴールデンテストで確かめるので、ここでは中身のない
// プレイリストで、どの年をまとめるかだけを確かめる。
func TestWrapped(t *testing.T) {
	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "年を省略すると今年をまとめる",
			data: subcommand("wrapped"),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL1", "お気に入り"), sub(2, "PL2", "作業用"))
				m.playlistHas("PL1")
				m.playlistHas("PL2")
			},
			want: command.Text("**🎁 お気に入り の 2026 年まとめ**\n2026 年に追加された曲はありません。\n\n**🎁 作業用 の 2026 年まとめ**\n2026 年に追加された曲はありません。\n"),
		},
		{
			name: "指定した年をまとめる",
			data: subcommand("wrapped", intOption("year", 2024)),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL1", "お気に入り"))
				m.playlistHas("PL1")
			},
			want: command.Text("**🎁 お気に入り の 2024 年まとめ**\n2024 年に追加された曲はありません。\n"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("wrapped"),
			setup: func(m *deps) {
				m.guildHas()
			},
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeps(t)
			tt.setup(m)

			got := newCommand(m).Handle(tt.data, "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
