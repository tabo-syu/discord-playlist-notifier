package service

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

func TestGetDiffFromLatest(t *testing.T) {
	available := video("v1", 10)
	later := video("v2", 20)
	private := &domain.YouTubeVideo{YoutubeID: "v3", Title: "was public", PrivacyStatus: domain.PrivacyPrivate}

	snapshots := map[string]*PlaylistSnapshot{
		"PL1": {
			YoutubeID: "PL1",
			Title:     "new title",
			Items: []*domain.PlaylistItem{
				{ItemID: "old", VideoYoutubeID: "v1", AddedAt: base.Add(-time.Minute)},
				{ItemID: "i2", VideoYoutubeID: "v2", AddedAt: base.Add(time.Minute)},
				{ItemID: "i1", VideoYoutubeID: "v1", AddedAt: base},
				{ItemID: "i3", VideoYoutubeID: "v3", AddedAt: base.Add(2 * time.Minute)},
			},
			Videos: map[string]*domain.YouTubeVideo{"v1": available, "v2": later, "v3": private},
		},
		"GONE": {YoutubeID: "GONE", Deleted: true},
	}

	registered := &domain.Playlist{YoutubeID: "PL1", Title: "old title"}
	registered.UpdatedAt = base
	gone := &domain.Playlist{YoutubeID: "GONE"}
	failed := &domain.Playlist{YoutubeID: "FAILED"}

	diffs := (&PlaylistService{}).GetDiffFromLatest([]*domain.Playlist{registered, gone, failed}, snapshots)

	if len(diffs) != 1 || diffs[0] != registered {
		t.Fatalf("expected only PL1 to be updated, got %+v", diffs)
	}
	if registered.Title != "new title" {
		t.Errorf("title should be updated, got %q", registered.Title)
	}
	videos := registered.Videos
	if len(videos) != 2 || videos[0].YoutubeID != "v1" || videos[1].YoutubeID != "v2" {
		t.Fatalf("expected v1 then v2, got %+v", videos)
	}
	if !videos[0].PublishedAt.Equal(base) || videos[1].Views != 20 {
		t.Errorf("unexpected video fields: %+v", videos)
	}
}
