package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name  string
		data  *discordgo.ApplicationCommandInteractionData
		setup func(m *deps)
		want  *command.Response
	}{
		{
			name: "プレイリストを YouTube のタイトルで登録する",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.subscriptions.EXPECT().Exists(mockGuild, library.PlaylistID("PL")).Return(false, nil)
				m.guilds.EXPECT().FindByDiscordID(mockGuild).Return(&guild.Guild{Model: gorm.Model{ID: 1}, DiscordID: "guild"}, nil)
				m.subscriptions.EXPECT().Add(&subscription.Subscription{
					YoutubeID:     "PL",
					SendChannelID: "ch",
					Title:         "タイトル",
					GuildID:       1,
				}).Return(nil)
			},
			want: command.Text("通知登録しました！\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "登録済みのプレイリスト",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.subscriptions.EXPECT().Exists(mockGuild, library.PlaylistID("PL")).Return(true, nil)
			},
			want: command.Text("既に通知登録されているプレイリストです。"),
		},
		{
			name: "YouTube にないプレイリスト",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, domain.ErrYouTubePlaylistNotFound)
			},
			want: command.Text("該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"),
		},
		{
			name: "YouTube API が失敗した",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.youtube.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, errMock)
			},
			want: command.Text("YouTube API のサービス状況を確認してください。"),
		},
		{
			name:  "プレイリスト ID がない",
			data:  subcommand("add"),
			setup: func(m *deps) {},
			want:  command.Text("Error: Playlist ID is required."),
		},
		{
			name:  "プレイリスト ID が空",
			data:  subcommand("add", stringOption("playlist-id", "")),
			setup: func(m *deps) {},
			want:  command.Text("Error: Playlist ID cannot be empty."),
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
