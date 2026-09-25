package service

import (
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

func TestRandomVideosPicksDistinctAvailableVideos(t *testing.T) {
	s, yt, _, _ := newTestService()
	yt.metas["PL1"].ItemCount = 4
	yt.items["PL1"] = []*repository.PlaylistItemInfo{
		item("i1", "v1", 0, domain.PrivacyPublic),
		item("i2", "v2", 1, domain.PrivacyPrivate),
		item("i3", "v3", 2, domain.PrivacyPublic),
		item("i4", "v1", 3, domain.PrivacyPublic),
	}
	if _, err := s.Sync([]string{"PL1"}); err != nil {
		t.Fatal(err)
	}

	for range 20 {
		picked, err := s.RandomVideos([]string{"PL1"}, 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(picked) != 2 {
			t.Fatalf("expected v1 and v3 once each, got %d videos", len(picked))
		}
		seen := map[string]bool{}
		for _, p := range picked {
			if p.PlaylistID != "PL1" || !p.Video.Available() || seen[p.Video.YoutubeID] {
				t.Fatalf("unexpected pick: %+v", p)
			}
			seen[p.Video.YoutubeID] = true
		}
	}

	picked, _ := s.RandomVideos([]string{"PL1"}, 1)
	if len(picked) != 1 {
		t.Errorf("expected 1 video, got %d", len(picked))
	}
	if none, _ := s.RandomVideos([]string{"UNKNOWN"}, 1); len(none) != 0 {
		t.Errorf("unknown playlists should give nothing, got %+v", none)
	}
}
