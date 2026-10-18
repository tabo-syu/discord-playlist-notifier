package library_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestSnapshotContains(t *testing.T) {
	s := &library.Snapshot{Items: []*library.PlaylistItem{{VideoYoutubeID: "a"}, {VideoYoutubeID: "b"}}}
	if !s.Contains("b") {
		t.Error(`Contains("b") = false, want true`)
	}
	if s.Contains("c") {
		t.Error(`Contains("c") = true, want false`)
	}
}

func TestPickRandom(t *testing.T) {
	now := time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)
	a, b, c := available("a", 1), available("b", 2), available("c", 3)
	contents := []*library.PlaylistVideo{
		added(a, now),
		added(b, now),
		// 2 回追加された動画も 1 回しか選ばない
		added(a, now.Add(time.Hour)),
		added(c, now),
		added(hidden("private", library.PrivacyPrivate), now),
		added(&library.Video{YoutubeID: "no-details", PrivacyStatus: library.PrivacyPublic}, now),
	}
	candidates := map[library.VideoID]bool{"a": true, "b": true, "c": true}

	for _, n := range []int{0, 2, 3, 10} {
		picked := library.PickRandom(contents, n)

		want := min(n, len(candidates))
		if len(picked) != want {
			t.Errorf("PickRandom(n=%d) returned %d videos, want %d", n, len(picked), want)
		}
		seen := map[library.VideoID]bool{}
		for _, p := range picked {
			id := p.Video.YoutubeID
			if !candidates[id] {
				t.Errorf("PickRandom(n=%d) picked %q, which is not available", n, id)
			}
			if seen[id] {
				t.Errorf("PickRandom(n=%d) picked %q twice", n, id)
			}
			seen[id] = true
		}
	}
}
