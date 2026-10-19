package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestGuildRepositoryExists(t *testing.T) {
	tests := []struct {
		name      string
		stored    []any
		discordID guild.DiscordID
		want      bool
	}{
		{
			name:      "保存済み",
			stored:    []any{guildRow("guild")},
			discordID: "guild",
			want:      true,
		},
		{
			name:      "保存していない",
			stored:    []any{guildRow("other")},
			discordID: "guild",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, tt.stored...)

			got, err := persistence.NewGuildRepository(db).Exists(tt.discordID)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGuildRepositoryAdd(t *testing.T) {
	tests := []struct {
		name      string
		guild     *guild.Guild
		want      []*persistence.GuildRecord
		wantGuild *guild.Guild
	}{
		{
			name:  "ギルドを保存し、振った ID を書き戻す",
			guild: guild.New("guild"),
			want: []*persistence.GuildRecord{
				{Model: model(1), DiscordID: "guild"},
			},
			wantGuild: &guild.Guild{ID: 1, DiscordID: "guild"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)

			err := persistence.NewGuildRepository(db).Add(tt.guild)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[persistence.GuildRecord](t, db))
			assert.Equal(t, tt.wantGuild, tt.guild)
		})
	}
}

func TestGuildRepositoryFindByDiscordID(t *testing.T) {
	tests := []struct {
		name      string
		stored    []any
		discordID guild.DiscordID
		want      *guild.Guild
		wantErr   error
	}{
		{
			name:      "保存済みのギルドを返す",
			stored:    []any{guildRow("other"), guildRow("guild")},
			discordID: "guild",
			want:      &guild.Guild{ID: 2, DiscordID: "guild"},
		},
		{
			// GORM のエラーをそのまま返す
			name:      "保存していない",
			stored:    []any{guildRow("other")},
			discordID: "guild",
			want:      nil,
			wantErr:   gorm.ErrRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, tt.stored...)

			got, err := persistence.NewGuildRepository(db).FindByDiscordID(tt.discordID)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGuildRepositoryDelete(t *testing.T) {
	tests := []struct {
		name   string
		stored []any
		guild  *guild.Guild
		want   []*persistence.GuildRecord
	}{
		{
			name:   "指定したギルドを論理削除する",
			stored: []any{guildRow("guild"), guildRow("other")},
			guild:  &guild.Guild{ID: 1, DiscordID: "guild"},
			want: []*persistence.GuildRecord{
				{Model: deleted(1), DiscordID: "guild"},
				{Model: model(2), DiscordID: "other"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, tt.stored...)

			err := persistence.NewGuildRepository(db).Delete(tt.guild)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[persistence.GuildRecord](t, db))
		})
	}
}
