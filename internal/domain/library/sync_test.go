package library_test

import (
	"slices"
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestReconcile(t *testing.T) {
	addedAt := time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)
	stored := []*library.PlaylistItem{
		{ItemID: "item-kept", VideoYoutubeID: "kept"},
		{ItemID: "item-hidden", VideoYoutubeID: "hidden"},
		{ItemID: "item-removed", VideoYoutubeID: "removed"},
	}
	fetched := []*library.FetchedItem{
		{ItemID: "item-kept", VideoID: "kept", PrivacyStatus: library.PrivacyPublic},
		{ItemID: "item-hidden", VideoID: "hidden", PrivacyStatus: library.PrivacyPrivate},
		// 同じ動画が 2 回追加された
		{ItemID: "item-new-1", VideoID: "new", AddedAt: addedAt, PrivacyStatus: library.PrivacyPublic},
		{ItemID: "item-new-2", VideoID: "new", AddedAt: addedAt.Add(time.Hour), PrivacyStatus: library.PrivacyPublic},
		{ItemID: "item-deleted", VideoID: "deleted", AddedAt: addedAt, PrivacyStatus: library.PrivacyDeleted},
	}
	videos := map[library.VideoID]*library.Video{
		"kept":   available("kept", 10),
		"hidden": available("hidden", 20),
	}

	s := library.Reconcile("PLxxxx", stored, fetched, videos)

	var addedIDs []library.ItemID
	for _, item := range s.Added {
		addedIDs = append(addedIDs, item.ItemID)
		if item.PlaylistYoutubeID != "PLxxxx" {
			t.Errorf("Added %s: PlaylistYoutubeID = %q, want %q", item.ItemID, item.PlaylistYoutubeID, "PLxxxx")
		}
	}
	assertIDs(t, "Added", addedIDs, []library.ItemID{"item-new-1", "item-new-2", "item-deleted"})
	if got := s.Added[1]; got.VideoYoutubeID != "new" || !got.AddedAt.Equal(addedAt.Add(time.Hour)) {
		t.Errorf("Added[1] = %+v, want the video and time of the fetched item", got)
	}

	var removedIDs []library.ItemID
	for _, item := range s.Removed {
		removedIDs = append(removedIDs, item.ItemID)
	}
	assertIDs(t, "Removed", removedIDs, []library.ItemID{"item-removed"})

	// 詳細を持っていた動画が見られなくなったときだけ報告する
	assertIDs(t, "Hidden", videoIDs(s.Hidden), []library.VideoID{"hidden"})
	// 視聴できない動画の詳細は取れないので取得しない。2 回追加された動画も 1 回だけ
	assertIDs(t, "NeedDetails", s.NeedDetails, []library.VideoID{"new"})

	// 初めて見る動画は渡したマップに追加する
	for _, id := range []library.VideoID{"kept", "hidden", "new", "deleted"} {
		if _, ok := videos[id]; !ok {
			t.Errorf("Videos is missing %q", id)
		}
	}
	if videos["deleted"].PrivacyStatus != library.PrivacyDeleted {
		t.Errorf("PrivacyStatus of a new video = %q, want %q", videos["deleted"].PrivacyStatus, library.PrivacyDeleted)
	}

	changed := videoIDs(s.ChangedVideos())
	slices.Sort(changed)
	assertIDs(t, "ChangedVideos", changed, []library.VideoID{"deleted", "hidden", "new"})

	s.ApplyDetails([]*library.Video{{YoutubeID: "new", Title: "新曲", Views: 250_000}})
	v := videos["new"]
	if v.Title != "新曲" || v.Views != 250_000 {
		t.Errorf("details were not applied: %+v", v)
	}
	if v.PrivacyStatus != library.PrivacyPublic {
		t.Errorf("PrivacyStatus = %q, want %q", v.PrivacyStatus, library.PrivacyPublic)
	}
	// 詳細を取得した時点の再生数は通知しない
	if v.ViewMilestone == nil || *v.ViewMilestone != 100_000 {
		t.Errorf("ViewMilestone = %v, want 100000", v.ViewMilestone)
	}
}

func TestReconcileNoChanges(t *testing.T) {
	stored := []*library.PlaylistItem{{ItemID: "item", VideoYoutubeID: "video"}}
	fetched := []*library.FetchedItem{{ItemID: "item", VideoID: "video", PrivacyStatus: library.PrivacyPublic}}
	videos := map[library.VideoID]*library.Video{"video": available("video", 10)}

	s := library.Reconcile("PLxxxx", stored, fetched, videos)

	if len(s.Added) != 0 || len(s.Removed) != 0 || len(s.Hidden) != 0 || len(s.NeedDetails) != 0 || len(s.ChangedVideos()) != 0 {
		t.Errorf("Reconcile() = %+v, want no changes", s)
	}
}
