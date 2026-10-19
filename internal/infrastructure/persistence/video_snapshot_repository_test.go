package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestVideoSnapshotRepositoryFind(t *testing.T) {
	tests := []struct {
		name string
		want []*library.VideoSnapshot
	}{
		{
			name: "指定した時刻以降に記録した動画の状態を、古い順に返す",
			want: []*library.VideoSnapshot{
				snapshot("a", 48, 200),
				snapshot("a", 12, 300),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			repo := persistence.NewVideoSnapshotRepository(db)
			assert.NoError(t, repo.Add([]*library.VideoSnapshot{
				snapshot("a", 12, 300),
				snapshot("a", 72, 100),
				snapshot("a", 48, 200),
				snapshot("b", 12, 999),
			}))

			got, err := repo.Find("a", hoursAgo(48))

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVideoSnapshotRepositoryFindBoundary(t *testing.T) {
	tests := []struct {
		name         string
		ids          []library.VideoID
		wantEarliest map[library.VideoID]*library.VideoSnapshot
		wantLatest   map[library.VideoID]*library.VideoSnapshot
	}{
		{
			name: "動画ごとに、指定した時刻以降で最初と最後の状態を返す",
			ids:  []library.VideoID{"a", "b", "none"},
			wantEarliest: map[library.VideoID]*library.VideoSnapshot{
				"a": snapshot("a", 48, 200),
				"b": snapshot("b", 24, 10),
			},
			wantLatest: map[library.VideoID]*library.VideoSnapshot{
				"a": snapshot("a", 12, 300),
				"b": snapshot("b", 24, 10),
			},
		},
		{
			name:         "動画を指定しなければ空のマップを返す",
			ids:          nil,
			wantEarliest: map[library.VideoID]*library.VideoSnapshot{},
			wantLatest:   map[library.VideoID]*library.VideoSnapshot{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, snapshotRow("a", 72, 100), snapshotRow("a", 48, 200), snapshotRow("a", 12, 300), snapshotRow("b", 24, 10))
			repo := persistence.NewVideoSnapshotRepository(db)

			earliest, earliestErr := repo.FindEarliest(tt.ids, hoursAgo(48))
			latest, latestErr := repo.FindLatest(tt.ids, hoursAgo(48))

			assert.NoError(t, earliestErr)
			assert.NoError(t, latestErr)
			assert.Equal(t, tt.wantEarliest, earliest)
			assert.Equal(t, tt.wantLatest, latest)
		})
	}
}
