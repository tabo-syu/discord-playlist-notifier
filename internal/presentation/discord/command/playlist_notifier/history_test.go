package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

// グラフの画像は比べられないので、グラフを添付しない場合だけを確かめる。
func TestHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "記録が足りない",
			data: subcommand("history", stringOption("video", "https://youtu.be/aaaaaaaaaaa")),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "PL"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"aaaaaaaaaaa"}).Return([]*library.Playlist{
					playlist("PL", item("PL", "aaaaaaaaaaa", hoursAgo(100))),
				}, nil)
				m.videos.EXPECT().Find([]library.VideoID{"aaaaaaaaaaa"}).Return(map[library.VideoID]*library.Video{
					"aaaaaaaaaaa": video("aaaaaaaaaaa"),
				}, nil)
				m.snapshots.EXPECT().Find(library.VideoID("aaaaaaaaaaa"), daysAgo(30)).Return([]*library.VideoSnapshot{
					{VideoYoutubeID: "aaaaaaaaaaa", RecordedAt: hoursAgo(24), Views: 100},
				}, nil)
			},
			want: command.Text("📈 **aaaaaaaaaaa** の推移（直近 30 日）\nまだ記録が足りません。再生数などは 6 時間ごとに記録しています。\nhttps://www.youtube.com/watch?v=aaaaaaaaaaa"),
		},
		{
			name: "登録したプレイリストにない動画",
			data: subcommand("history", stringOption("video", "aaaaaaaaaaa"), intOption("days", 7)),
			setup: func(m *deps) {
				m.guildHas(sub(1, "PL", "PL"))
				m.playlists.EXPECT().FindContaining([]library.VideoID{"aaaaaaaaaaa"}).Return([]*library.Playlist{
					playlist("OTHER", item("OTHER", "aaaaaaaaaaa", hoursAgo(100))),
				}, nil)
			},
			want: command.Text("通知登録されているプレイリストに、その動画は見つかりませんでした。"),
		},
		{
			name:  "動画の URL でも ID でもない",
			data:  subcommand("history", stringOption("video", "https://example.com/")),
			setup: func(m *deps) {},
			want:  command.Text("動画の URL か ID を入力してください。"),
		},
		{
			name: "登録したプレイリストがない",
			data: subcommand("history", stringOption("video", "aaaaaaaaaaa")),
			setup: func(m *deps) {
				m.guildHas()
			},
			want: command.Text("通知登録されているプレイリストが存在しません。"),
		},
		{
			name:  "動画がない",
			data:  subcommand("history"),
			setup: func(m *deps) {},
			want:  command.Text("Error: Video is required."),
		},
		{
			name:  "日数が 366",
			data:  subcommand("history", stringOption("video", "aaaaaaaaaaa"), intOption("days", 366)),
			setup: func(m *deps) {},
			want:  command.Text("Error: days must be between 1 and 365."),
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
