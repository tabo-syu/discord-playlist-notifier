package playlist_notifier_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command/playlist_notifier"
)

func TestAdd(t *testing.T) {
	t.Parallel()

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
				m.source.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.guilds.EXPECT().FindByDiscordID(mockGuild).Return(&guild.Guild{ID: 1, DiscordID: "guild"}, nil)
				m.subscriptions.EXPECT().Add(&subscription.Subscription{
					YoutubeID:     "PL",
					SendChannelID: "ch",
					Title:         "タイトル",
					GuildID:       1,
					NotifiedUntil: now,
				}).Return(nil)
			},
			want: command.Text("通知登録しました！\nhttps://www.youtube.com/playlist?list=PL"),
		},
		{
			name: "登録済みのプレイリスト",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.source.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(&library.PlaylistMeta{YoutubeID: "PL", Title: "タイトル"}, nil)
				m.guilds.EXPECT().FindByDiscordID(mockGuild).Return(&guild.Guild{ID: 1, DiscordID: "guild"}, nil)
				m.subscriptions.EXPECT().Add(&subscription.Subscription{
					YoutubeID:     "PL",
					SendChannelID: "ch",
					Title:         "タイトル",
					GuildID:       1,
					NotifiedUntil: now,
				}).Return(subscription.ErrAlreadySubscribed)
			},
			want: command.Text("既に通知登録されているプレイリストです。"),
		},
		{
			name: "YouTube にないプレイリスト",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.source.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, library.ErrPlaylistNotFound)
			},
			want: command.Text("該当するプレイリストが見つかりませんでした...\n非公開のプレイリストではありませんか？"),
		},
		{
			name: "YouTube API が失敗した",
			data: subcommand("add", stringOption("playlist-id", "PL")),
			setup: func(m *deps) {
				m.source.EXPECT().FindPlaylist(library.PlaylistID("PL")).Return(nil, errMock)
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
		{
			name:  "プレイリスト ID の形式が正しくなければ、何も呼ばずに断る",
			data:  subcommand("add", stringOption("playlist-id", "https://www.youtube.com/playlist?list=PL")),
			setup: func(m *deps) {},
			want:  command.Text(playlist_notifier.INVALID_PLAYLIST_ID_MESSAGE),
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
