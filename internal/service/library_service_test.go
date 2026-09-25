package service

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

var base = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

func item(id, video string, minutes int, privacy string) *repository.PlaylistItemInfo {
	return &repository.PlaylistItemInfo{ItemID: id, VideoID: video, AddedAt: base.Add(time.Duration(minutes) * time.Minute), PrivacyStatus: privacy}
}

func video(id string, views uint64) *domain.YouTubeVideo {
	return &domain.YouTubeVideo{YoutubeID: id, Title: "title " + id, ChannelName: "ch", Views: views, PrivacyStatus: domain.PrivacyPublic}
}

func newTestService() (*LibraryService, *fakeYouTube, *fakeLibrary, *time.Time) {
	yt := &fakeYouTube{
		metas: map[string]*repository.PlaylistMeta{"PL1": {YoutubeID: "PL1", Title: "list", ItemCount: 2}},
		items: map[string][]*repository.PlaylistItemInfo{"PL1": {
			item("i1", "v1", 0, domain.PrivacyPublic),
			item("i2", "v2", 1, domain.PrivacyPrivate),
		}},
		videos: map[string]*domain.YouTubeVideo{"v1": video("v1", 10), "v3": video("v3", 30)},
	}
	lib := newFakeLibrary()
	s := NewLibraryService(yt, lib)
	now := base.Add(24 * time.Hour)
	s.now = func() time.Time { return now }

	return s, yt, lib, &now
}

func TestSyncStoresItemsAndOnlyFetchesAvailableVideos(t *testing.T) {
	s, yt, lib, _ := newTestService()

	snapshots, err := s.Sync([]string{"PL1", "PL1"})
	if err != nil {
		t.Fatal(err)
	}

	snap := snapshots["PL1"]
	if len(snap.Items) != 2 || snap.Title != "list" {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
	if len(yt.videoCalls) != 1 || len(yt.videoCalls[0]) != 1 || yt.videoCalls[0][0] != "v1" {
		t.Fatalf("expected details to be fetched only for v1, got %v", yt.videoCalls)
	}
	if !lib.videos["v1"].Available() {
		t.Errorf("v1 should be available: %+v", lib.videos["v1"])
	}
	if v := lib.videos["v2"]; v == nil || v.PrivacyStatus != domain.PrivacyPrivate || v.Available() {
		t.Errorf("v2 should be stored as private without details: %+v", v)
	}
}

func TestSyncSkipsFetchingItemsWhileCountIsUnchanged(t *testing.T) {
	s, yt, _, now := newTestService()
	if _, err := s.Sync([]string{"PL1"}); err != nil {
		t.Fatal(err)
	}

	*now = now.Add(30 * time.Minute)
	snapshots, err := s.Sync([]string{"PL1"})
	if err != nil {
		t.Fatal(err)
	}
	if yt.itemCalls != 1 {
		t.Errorf("items should not be fetched again, got %d calls", yt.itemCalls)
	}
	if len(snapshots["PL1"].Items) != 2 {
		t.Errorf("snapshot should come from the database: %+v", snapshots["PL1"])
	}

	*now = now.Add(FULL_SYNC_INTERVAL)
	if _, err := s.Sync([]string{"PL1"}); err != nil {
		t.Fatal(err)
	}
	if yt.itemCalls != 2 {
		t.Errorf("items should be fetched after the full sync interval, got %d calls", yt.itemCalls)
	}
}

func TestSyncAppliesAdditionsRemovalsAndPrivacyChanges(t *testing.T) {
	s, yt, lib, now := newTestService()
	if _, err := s.Sync([]string{"PL1"}); err != nil {
		t.Fatal(err)
	}

	// v1 becomes private, i2 is removed and v3 is added
	yt.metas["PL1"].ItemCount = 3
	yt.items["PL1"] = []*repository.PlaylistItemInfo{
		item("i1", "v1", 0, domain.PrivacyPrivate),
		item("i3", "v3", 2, domain.PrivacyPublic),
	}
	*now = now.Add(time.Minute)

	snapshots, err := s.Sync([]string{"PL1"})
	if err != nil {
		t.Fatal(err)
	}

	items := snapshots["PL1"].Items
	if len(items) != 2 || items[0].ItemID != "i1" || items[1].ItemID != "i3" {
		t.Fatalf("unexpected items: %+v %+v", items[0], items[1])
	}
	v1 := lib.videos["v1"]
	if v1.PrivacyStatus != domain.PrivacyPrivate || v1.Title != "title v1" {
		t.Errorf("v1 should be private and keep its details: %+v", v1)
	}
	if !lib.videos["v3"].Available() {
		t.Errorf("v3 should be available: %+v", lib.videos["v3"])
	}
}

func TestSyncMarksMissingPlaylistsAsDeleted(t *testing.T) {
	s, _, _, _ := newTestService()

	snapshots, err := s.Sync([]string{"PL1", "GONE"})
	if err != nil {
		t.Fatal(err)
	}
	if !snapshots["GONE"].Deleted || snapshots["PL1"].Deleted {
		t.Errorf("unexpected snapshots: %+v", snapshots)
	}
}

func TestRefreshVideosUpdatesDetailsButKeepsIconAndPrivacy(t *testing.T) {
	s, yt, lib, _ := newTestService()
	if _, err := s.Sync([]string{"PL1"}); err != nil {
		t.Fatal(err)
	}
	lib.videos["v1"].ChannelIcon = "icon"
	yt.videos["v1"] = video("v1", 999)

	if err := s.RefreshVideos(); err != nil {
		t.Fatal(err)
	}

	v1 := lib.videos["v1"]
	if v1.Views != 999 || v1.ChannelIcon != "icon" || v1.PrivacyStatus != domain.PrivacyPublic {
		t.Errorf("unexpected v1 after refresh: %+v", v1)
	}
	last := yt.videoCalls[len(yt.videoCalls)-1]
	if len(last) != 1 || last[0] != "v1" {
		t.Errorf("only available videos should be refreshed, got %v", last)
	}
}
