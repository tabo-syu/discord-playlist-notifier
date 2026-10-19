package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestPlaylistReconcile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		playlist    *library.Playlist
		fetched     []*library.FetchedItem
		want        *library.Playlist
		wantAdded   []*library.PlaylistItem
		wantRemoved []*library.PlaylistItem
	}{
		{
			name:      "新しいアイテムを追加する",
			playlist:  playlist(),
			fetched:   []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			want:      playlist(stored("1", "a")),
			wantAdded: []*library.PlaylistItem{stored("1", "a")},
		},
		{
			name:        "消えたアイテムを除く",
			playlist:    playlist(stored("1", "a")),
			want:        playlist(),
			wantRemoved: []*library.PlaylistItem{stored("1", "a")},
		},
		{
			name:     "非公開になったアイテムも残す",
			playlist: playlist(stored("1", "a")),
			fetched:  []*library.FetchedItem{fetched("1", "a", library.PrivacyPrivate)},
			want:     playlist(stored("1", "a")),
		},
		{
			name:      "同じ動画でもアイテムが違えば別に追加する",
			playlist:  playlist(),
			fetched:   []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic), fetched("2", "a", library.PrivacyPublic)},
			want:      playlist(stored("1", "a"), stored("2", "a")),
			wantAdded: []*library.PlaylistItem{stored("1", "a"), stored("2", "a")},
		},
		{
			name:     "アイテムを追加された時刻の順に並べる",
			playlist: playlist(storedAt("2", "b", hoursLater(2))),
			fetched: []*library.FetchedItem{
				fetchedAt("2", "b", hoursLater(2), library.PrivacyPublic),
				fetchedAt("1", "a", hoursLater(1), library.PrivacyPublic),
			},
			want:      playlist(storedAt("1", "a", hoursLater(1)), storedAt("2", "b", hoursLater(2))),
			wantAdded: []*library.PlaylistItem{storedAt("1", "a", hoursLater(1))},
		},
		{
			name:     "変化がなければ何もしない",
			playlist: playlist(stored("1", "a")),
			fetched:  []*library.FetchedItem{fetched("1", "a", library.PrivacyPublic)},
			want:     playlist(stored("1", "a")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			added, removed := tt.playlist.Reconcile(tt.fetched)

			assert.Equal(t, tt.want, tt.playlist)
			assert.Equal(t, tt.wantAdded, added)
			assert.Equal(t, tt.wantRemoved, removed)
		})
	}
}

func TestPlaylistContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		videoID library.VideoID
		want    bool
	}{
		{
			name:    "含む",
			videoID: "a",
			want:    true,
		},
		{
			name:    "含まない",
			videoID: "b",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := playlist(stored("1", "a")).Contains(tt.videoID)

			assert.Equal(t, tt.want, got)
		})
	}
}
