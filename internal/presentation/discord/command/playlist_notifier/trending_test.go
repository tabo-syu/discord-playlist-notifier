package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// グラフの画像は比べられないので、グラフを添付しない場合だけを確かめる。
func TestTrending(t *testing.T) {
	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "記録が足りない",
			data: subcommand("trending"),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "お気に入り"))
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{}, nil)
				m.library.EXPECT().FindEarliestSnapshots([]library.VideoID{}, daysAgo(30)).Return(map[library.VideoID]*library.VideoSnapshot{}, nil)
				m.library.EXPECT().FindLatestSnapshots([]library.VideoID{}, daysAgo(30)).Return(map[library.VideoID]*library.VideoSnapshot{}, nil)
			},
			want: &command.Response{
				Content: "🔥 **お気に入り** の急上昇\nまだ記録が足りません。再生数は 6 時間ごとに記録しています。\n",
			},
		},
		{
			name: "指定した日数の間に伸びた曲がない",
			data: subcommand("trending", intOption("days", 7)),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "お気に入り"))
				m.library.EXPECT().FindVideosInPlaylists(library.PlaylistID("PL")).Return(map[library.VideoID]*library.Video{
					"a": video("a"),
				}, nil)
				m.library.EXPECT().FindEarliestSnapshots([]library.VideoID{"a"}, daysAgo(7)).Return(map[library.VideoID]*library.VideoSnapshot{
					"a": {VideoYoutubeID: "a", RecordedAt: hoursAgo(48), Views: 100},
				}, nil)
				m.library.EXPECT().FindLatestSnapshots([]library.VideoID{"a"}, daysAgo(7)).Return(map[library.VideoID]*library.VideoSnapshot{
					"a": {VideoYoutubeID: "a", RecordedAt: hoursAgo(24), Views: 100},
				}, nil)
			},
			want: &command.Response{
				Content: "🔥 **お気に入り** の急上昇\n10/16〜10/17 の伸び\nこの期間に再生数が伸びた曲はありません。\n",
			},
		},
		{
			name: "登録していないプレイリストを指定した",
			data: subcommand("trending", stringOption("playlist-id", "OTHER")),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "お気に入り"))
			},
			want: command.Text("通知登録されていないプレイリストです。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("trending"),
			setup: func(m *deps) {
				m.guildHas()
			},
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name:  "日数が 0",
			data:  subcommand("trending", intOption("days", 0)),
			setup: func(m *deps) {},
			want:  command.Text("Error: days must be between 1 and 365."),
		},
		{
			name:  "日数が 366",
			data:  subcommand("trending", intOption("days", 366)),
			setup: func(m *deps) {},
			want:  command.Text("Error: days must be between 1 and 365."),
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
