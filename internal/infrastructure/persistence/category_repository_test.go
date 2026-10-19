package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestCategoryRepositorySave(t *testing.T) {
	tests := []struct {
		name       string
		categories []*library.Category
		want       map[library.CategoryID]*library.Category
	}{
		{
			name: "保存済みのカテゴリは名前と設定できるかを上書きし、新しいカテゴリは追加する",
			categories: []*library.Category{
				{YoutubeID: "10", Title: "音楽", Assignable: false},
				{YoutubeID: "20", Title: "ゲーム", Assignable: true},
			},
			want: map[library.CategoryID]*library.Category{
				"10": {YoutubeID: "10", Title: "音楽", Assignable: false},
				"20": {YoutubeID: "20", Title: "ゲーム", Assignable: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.CategoryRecord{YoutubeID: "10", Title: "Music", Assignable: true})
			repo := persistence.NewCategoryRepository(db)

			err := repo.Save(tt.categories)
			got, findErr := repo.FindAll()

			assert.NoError(t, err)
			assert.NoError(t, findErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
