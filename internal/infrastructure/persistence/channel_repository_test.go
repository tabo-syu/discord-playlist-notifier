package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestChannelRepositorySave(t *testing.T) {
	tests := []struct {
		name     string
		channels []*library.Channel
		want     map[library.ChannelID]*library.Channel
	}{
		{
			name: "保存済みのチャンネルは名前を上書きし、新しいチャンネルは追加する",
			channels: []*library.Channel{
				{YoutubeID: "UC1", Title: "新しい名前"},
				{YoutubeID: "UC2", Title: "UC2"},
			},
			want: map[library.ChannelID]*library.Channel{
				"UC1": {YoutubeID: "UC1", Title: "新しい名前"},
				"UC2": {YoutubeID: "UC2", Title: "UC2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.ChannelRecord{YoutubeID: "UC1", Title: "古い名前"})
			repo := persistence.NewChannelRepository(db)

			err := repo.Save(tt.channels)
			got, findErr := repo.Find([]library.ChannelID{"UC1", "UC2"})

			assert.NoError(t, err)
			assert.NoError(t, findErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
