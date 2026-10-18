package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

func TestLibraryRepositoryFindItems(t *testing.T) {
	tests := []struct {
		name        string
		stored      []any
		playlistIDs []library.PlaylistID
		want        []*library.PlaylistItem
	}{
		{
			name: "指定したプレイリストのアイテムを、追加された順に返す",
			stored: []any{
				newItem("PL1", "1", "a", hoursAgo(2)),
				newItem("PL2", "2", "b", hoursAgo(1)),
				newItem("PL1", "3", "c", hoursAgo(3)),
				newItem("OTHER", "4", "d", hoursAgo(4)),
			},
			playlistIDs: []library.PlaylistID{"PL1", "PL2"},
			want: []*library.PlaylistItem{
				storedItem(3, "PL1", "3", "c", hoursAgo(3)),
				storedItem(1, "PL1", "1", "a", hoursAgo(2)),
				storedItem(2, "PL2", "2", "b", hoursAgo(1)),
			},
		},
		{
			name:        "アイテムがなければ空のスライスを返す",
			playlistIDs: []library.PlaylistID{"PL1"},
			want:        []*library.PlaylistItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, tt.stored...)

			got, err := persistence.NewLibraryRepository(db).FindItems(tt.playlistIDs...)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryFindVideos(t *testing.T) {
	tests := []struct {
		name string
		ids  []library.VideoID
		want map[library.VideoID]*library.Video
	}{
		{
			name: "指定した動画を返す",
			ids:  []library.VideoID{"a", "c", "unknown"},
			want: map[library.VideoID]*library.Video{
				"a": storedVideo(1, "a", 100),
				"c": storedVideo(3, "c", 300),
			},
		},
		{
			name: "動画を指定しなければ空のマップを返す",
			ids:  nil,
			want: map[library.VideoID]*library.Video{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, newVideo("a", 100), newVideo("b", 200), newVideo("c", 300))

			got, err := persistence.NewLibraryRepository(db).FindVideos(tt.ids)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryFindItemsByVideos(t *testing.T) {
	tests := []struct {
		name string
		ids  []library.VideoID
		want []*library.PlaylistItem
	}{
		{
			name: "指定した動画を含む全てのプレイリストのアイテムを返す",
			ids:  []library.VideoID{"a"},
			want: []*library.PlaylistItem{
				storedItem(1, "PL1", "1", "a", hoursAgo(2)),
				storedItem(3, "PL2", "3", "a", hoursAgo(1)),
			},
		},
		{
			name: "動画を指定しなければ空のスライスを返す",
			ids:  nil,
			want: []*library.PlaylistItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, newItem("PL1", "1", "a", hoursAgo(2)), newItem("PL1", "2", "b", hoursAgo(2)), newItem("PL2", "3", "a", hoursAgo(1)))

			got, err := persistence.NewLibraryRepository(db).FindItemsByVideos(tt.ids)

			// 並び順は決めていない
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryFindVideosInPlaylists(t *testing.T) {
	tests := []struct {
		name        string
		playlistIDs []library.PlaylistID
		want        map[library.VideoID]*library.Video
	}{
		{
			name:        "指定したプレイリストのアイテムの動画を返す",
			playlistIDs: []library.PlaylistID{"PL1", "PL2"},
			want: map[library.VideoID]*library.Video{
				"a": storedVideo(1, "a", 100),
				"b": storedVideo(2, "b", 200),
			},
		},
		{
			name:        "アイテムのないプレイリストは空のマップを返す",
			playlistIDs: []library.PlaylistID{"EMPTY"},
			want:        map[library.VideoID]*library.Video{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				newVideo("a", 100), newVideo("b", 200), newVideo("c", 300), newVideo("orphan", 400),
				newItem("PL1", "1", "a", hoursAgo(1)), newItem("PL2", "2", "b", hoursAgo(1)), newItem("OTHER", "3", "c", hoursAgo(1)),
			)

			got, err := persistence.NewLibraryRepository(db).FindVideosInPlaylists(tt.playlistIDs...)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryFindListedVideos(t *testing.T) {
	tests := []struct {
		name string
		want map[library.VideoID]*library.Video
	}{
		{
			name: "どれかのプレイリストにある動画だけを返す",
			want: map[library.VideoID]*library.Video{
				"a": storedVideo(1, "a", 100),
				"b": storedVideo(2, "b", 200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				newVideo("a", 100), newVideo("b", 200), newVideo("orphan", 300),
				newItem("PL1", "1", "a", hoursAgo(1)), newItem("PL2", "2", "b", hoursAgo(1)), newItem("PL2", "3", "a", hoursAgo(1)),
			)

			got, err := persistence.NewLibraryRepository(db).FindListedVideos()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryApplyItems(t *testing.T) {
	tests := []struct {
		name       string
		added      []*library.PlaylistItem
		removed    []*library.PlaylistItem
		videos     []*library.Video
		wantItems  []*library.PlaylistItem
		wantVideos []*library.Video
	}{
		{
			name: "アイテムを足して除き、動画を保存する",
			added: []*library.PlaylistItem{
				newItem("PL", "3", "c", hoursAgo(1)),
			},
			removed: []*library.PlaylistItem{
				storedItem(2, "PL", "2", "b", hoursAgo(2)),
			},
			videos: []*library.Video{
				storedVideo(1, "a", 150),
				newVideo("c", 300),
			},
			wantItems: []*library.PlaylistItem{
				storedItem(1, "PL", "1", "a", hoursAgo(3)),
				storedItem(3, "PL", "3", "c", hoursAgo(1)),
			},
			wantVideos: []*library.Video{
				storedVideo(1, "a", 150),
				storedVideo(2, "b", 200),
				storedVideo(3, "c", 300),
			},
		},
		{
			name: "何も変わらなければ何もしない",
			wantItems: []*library.PlaylistItem{
				storedItem(1, "PL", "1", "a", hoursAgo(3)),
				storedItem(2, "PL", "2", "b", hoursAgo(2)),
			},
			wantVideos: []*library.Video{
				storedVideo(1, "a", 100),
				storedVideo(2, "b", 200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				newItem("PL", "1", "a", hoursAgo(3)), newItem("PL", "2", "b", hoursAgo(2)),
				newVideo("a", 100), newVideo("b", 200),
			)

			err := persistence.NewLibraryRepository(db).ApplyItems("PL", tt.added, tt.removed, tt.videos)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantItems, rows[library.PlaylistItem](t, db))
			assert.Equal(t, tt.wantVideos, rows[library.Video](t, db))
		})
	}
}

func TestLibraryRepositorySaveVideos(t *testing.T) {
	milestone := library.ViewCount(100_000)

	tests := []struct {
		name   string
		videos []*library.Video
		want   []*library.Video
	}{
		{
			name: "保存済みの動画は全てのカラムを上書きし、新しい動画は追加する",
			videos: []*library.Video{
				{YoutubeID: "a", Title: "新しいタイトル", PrivacyStatus: library.PrivacyUnlisted, Views: 120_000, ViewMilestone: &milestone, Tags: []string{"tag"}, CategoryID: "10", ChannelYoutubeID: "UC1"},
				newVideo("b", 200),
			},
			want: []*library.Video{
				{ID: 1, CreatedAt: now, UpdatedAt: now, YoutubeID: "a", Title: "新しいタイトル", PrivacyStatus: library.PrivacyUnlisted, Views: 120_000, ViewMilestone: &milestone, Tags: []string{"tag"}, CategoryID: "10", ChannelYoutubeID: "UC1"},
				storedVideo(3, "b", 200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, newVideo("a", 100))

			err := persistence.NewLibraryRepository(db).SaveVideos(tt.videos)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[library.Video](t, db))
		})
	}
}

func TestLibraryRepositoryFindSnapshots(t *testing.T) {
	tests := []struct {
		name string
		want []*library.VideoSnapshot
	}{
		{
			name: "指定した時刻以降に記録した動画の状態を、古い順に返す",
			want: []*library.VideoSnapshot{
				storedSnapshot(3, "a", 48, 200),
				storedSnapshot(1, "a", 12, 300),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			repo := persistence.NewLibraryRepository(db)
			assert.NoError(t, repo.AddSnapshots([]*library.VideoSnapshot{
				newSnapshot("a", 12, 300),
				newSnapshot("a", 72, 100),
				newSnapshot("a", 48, 200),
				newSnapshot("b", 12, 999),
			}))

			got, err := repo.FindSnapshots("a", hoursAgo(48))

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositoryFindBoundarySnapshots(t *testing.T) {
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
				"a": storedSnapshot(2, "a", 48, 200),
				"b": storedSnapshot(4, "b", 24, 10),
			},
			wantLatest: map[library.VideoID]*library.VideoSnapshot{
				"a": storedSnapshot(3, "a", 12, 300),
				"b": storedSnapshot(4, "b", 24, 10),
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
			seed(t, db, newSnapshot("a", 72, 100), newSnapshot("a", 48, 200), newSnapshot("a", 12, 300), newSnapshot("b", 24, 10))
			repo := persistence.NewLibraryRepository(db)

			earliest, earliestErr := repo.FindEarliestSnapshots(tt.ids, hoursAgo(48))
			latest, latestErr := repo.FindLatestSnapshots(tt.ids, hoursAgo(48))

			assert.NoError(t, earliestErr)
			assert.NoError(t, latestErr)
			assert.Equal(t, tt.wantEarliest, earliest)
			assert.Equal(t, tt.wantLatest, latest)
		})
	}
}

func TestLibraryRepositorySaveChannels(t *testing.T) {
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
				"UC1": {ID: 1, CreatedAt: now, UpdatedAt: now, YoutubeID: "UC1", Title: "新しい名前"},
				"UC2": {ID: 3, CreatedAt: now, UpdatedAt: now, YoutubeID: "UC2", Title: "UC2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &library.Channel{YoutubeID: "UC1", Title: "古い名前"})
			repo := persistence.NewLibraryRepository(db)

			err := repo.SaveChannels(tt.channels)
			got, findErr := repo.FindChannels([]library.ChannelID{"UC1", "UC2"})

			assert.NoError(t, err)
			assert.NoError(t, findErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLibraryRepositorySaveCategories(t *testing.T) {
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
				"10": {ID: 1, CreatedAt: now, UpdatedAt: now, YoutubeID: "10", Title: "音楽", Assignable: false},
				"20": {ID: 3, CreatedAt: now, UpdatedAt: now, YoutubeID: "20", Title: "ゲーム", Assignable: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &library.Category{YoutubeID: "10", Title: "Music", Assignable: true})
			repo := persistence.NewLibraryRepository(db)

			err := repo.SaveCategories(tt.categories)
			got, findErr := repo.FindCategories()

			assert.NoError(t, err)
			assert.NoError(t, findErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
