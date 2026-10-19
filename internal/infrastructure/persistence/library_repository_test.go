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
				itemRow("PL1", "1", "a", hoursAgo(2)),
				itemRow("PL2", "2", "b", hoursAgo(1)),
				itemRow("PL1", "3", "c", hoursAgo(3)),
				itemRow("OTHER", "4", "d", hoursAgo(4)),
			},
			playlistIDs: []library.PlaylistID{"PL1", "PL2"},
			want: []*library.PlaylistItem{
				item("PL1", "3", "c", hoursAgo(3)),
				item("PL1", "1", "a", hoursAgo(2)),
				item("PL2", "2", "b", hoursAgo(1)),
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
				"a": video("a", 100),
				"c": video("c", 300),
			},
		},
		{
			name: "行の全てのカラムを動画に写す",
			ids:  []library.VideoID{"full"},
			want: map[library.VideoID]*library.Video{
				"full": fullVideo(),
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
			seed(t, db, videoRow("a", 100), videoRow("b", 200), videoRow("c", 300), fullVideoRow())

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
				item("PL1", "1", "a", hoursAgo(2)),
				item("PL2", "3", "a", hoursAgo(1)),
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
			seed(t, db, itemRow("PL1", "1", "a", hoursAgo(2)), itemRow("PL1", "2", "b", hoursAgo(2)), itemRow("PL2", "3", "a", hoursAgo(1)))

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
				"a": video("a", 100),
				"b": video("b", 200),
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
				videoRow("a", 100), videoRow("b", 200), videoRow("c", 300), videoRow("orphan", 400),
				itemRow("PL1", "1", "a", hoursAgo(1)), itemRow("PL2", "2", "b", hoursAgo(1)), itemRow("OTHER", "3", "c", hoursAgo(1)),
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
				"a": video("a", 100),
				"b": video("b", 200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				videoRow("a", 100), videoRow("b", 200), videoRow("orphan", 300),
				itemRow("PL1", "1", "a", hoursAgo(1)), itemRow("PL2", "2", "b", hoursAgo(1)), itemRow("PL2", "3", "a", hoursAgo(1)),
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
		wantItems  []*persistence.PlaylistItemRecord
		wantVideos []*persistence.VideoRecord
	}{
		{
			// 外れたアイテムは YouTube のアイテム ID で消す
			name: "アイテムを足して除き、動画を保存する",
			added: []*library.PlaylistItem{
				item("PL", "3", "c", hoursAgo(1)),
			},
			removed: []*library.PlaylistItem{
				item("PL", "2", "b", hoursAgo(2)),
			},
			videos: []*library.Video{
				video("a", 150),
				video("c", 300),
			},
			wantItems: []*persistence.PlaylistItemRecord{
				storedItemRow(1, itemRow("PL", "1", "a", hoursAgo(3))),
				storedItemRow(3, itemRow("PL", "3", "c", hoursAgo(1))),
			},
			// 競合した a の INSERT も ID を 1 つ使うので、c の ID は 4 になる
			wantVideos: []*persistence.VideoRecord{
				storedVideoRow(1, videoRow("a", 150)),
				storedVideoRow(2, videoRow("b", 200)),
				storedVideoRow(4, videoRow("c", 300)),
			},
		},
		{
			name: "何も変わらなければ何もしない",
			wantItems: []*persistence.PlaylistItemRecord{
				storedItemRow(1, itemRow("PL", "1", "a", hoursAgo(3))),
				storedItemRow(2, itemRow("PL", "2", "b", hoursAgo(2))),
			},
			wantVideos: []*persistence.VideoRecord{
				storedVideoRow(1, videoRow("a", 100)),
				storedVideoRow(2, videoRow("b", 200)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db,
				itemRow("PL", "1", "a", hoursAgo(3)), itemRow("PL", "2", "b", hoursAgo(2)),
				videoRow("a", 100), videoRow("b", 200),
			)

			err := persistence.NewLibraryRepository(db).ApplyItems("PL", tt.added, tt.removed, tt.videos)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantItems, rows[persistence.PlaylistItemRecord](t, db))
			assert.Equal(t, tt.wantVideos, rows[persistence.VideoRecord](t, db))
		})
	}
}

func TestLibraryRepositorySaveVideos(t *testing.T) {
	tests := []struct {
		name   string
		videos []*library.Video
		want   []*persistence.VideoRecord
	}{
		{
			// id と created_at は保存済みの行のまま、updated_at は保存した時刻にする。
			// 保存済みの行は詳細を取得していないので、details_fetched も上書きされることを確かめる
			name: "保存済みの動画は全てのカラムを上書きし、新しい動画は追加する",
			videos: []*library.Video{
				fullVideo(),
				video("b", 200),
			},
			want: []*persistence.VideoRecord{
				storedVideoRow(1, fullVideoRow()),
				storedVideoRow(3, videoRow("b", 200)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.VideoRecord{YoutubeID: "full", PrivacyStatus: "public"})

			err := persistence.NewLibraryRepository(db).SaveVideos(tt.videos)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, rows[persistence.VideoRecord](t, db))
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
				snapshot("a", 48, 200),
				snapshot("a", 12, 300),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			repo := persistence.NewLibraryRepository(db)
			assert.NoError(t, repo.AddSnapshots([]*library.VideoSnapshot{
				snapshot("a", 12, 300),
				snapshot("a", 72, 100),
				snapshot("a", 48, 200),
				snapshot("b", 12, 999),
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
				"UC1": {YoutubeID: "UC1", Title: "新しい名前"},
				"UC2": {YoutubeID: "UC2", Title: "UC2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.ChannelRecord{YoutubeID: "UC1", Title: "古い名前"})
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
				"10": {YoutubeID: "10", Title: "音楽", Assignable: false},
				"20": {YoutubeID: "20", Title: "ゲーム", Assignable: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, &persistence.CategoryRecord{YoutubeID: "10", Title: "Music", Assignable: true})
			repo := persistence.NewLibraryRepository(db)

			err := repo.SaveCategories(tt.categories)
			got, findErr := repo.FindCategories()

			assert.NoError(t, err)
			assert.NoError(t, findErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
