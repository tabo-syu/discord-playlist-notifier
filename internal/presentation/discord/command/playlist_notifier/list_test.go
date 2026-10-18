package playlist_notifier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "登録したプレイリストの URL を並べる",
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL1", "PL1"), sub(2, "PL2", "PL2"))
			},
			want: command.Text("通知登録されているプレイリスト一覧\nhttps://www.youtube.com/playlist?list=PL1\nhttps://www.youtube.com/playlist?list=PL2\n"),
		},
		{
			name: "登録したプレイリストがない",
			setup: func(m *deps) {
				m.guildHas()
			},
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name: "DB の読み込みに失敗した",
			setup: func(m *deps) {
				m.subscriptions.EXPECT().FindByGuild(mockGuild).Return(nil, errMock)
			},
			want: command.Text("エラー！システムに問題があります！"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newDeps(t)
			tt.setup(m)

			got := newCommand(m).Handle(subcommand("list"), "guild", "ch")

			assert.Equal(t, tt.want, got)
		})
	}
}
