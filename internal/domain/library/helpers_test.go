package library_test

import (
	"slices"
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

func available(id string, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: library.VideoID(id), Title: id, Views: views, PrivacyStatus: library.PrivacyPublic}
}

func hidden(id string, status library.PrivacyStatus) *library.Video {
	return &library.Video{YoutubeID: library.VideoID(id), Title: id, PrivacyStatus: status}
}

func added(v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: "PLxxxx",
		Item:       &library.PlaylistItem{VideoYoutubeID: v.YoutubeID, AddedAt: addedAt},
		Video:      v,
	}
}

func videoIDs(videos []*library.Video) []library.VideoID {
	ids := []library.VideoID{}
	for _, v := range videos {
		ids = append(ids, v.YoutubeID)
	}

	return ids
}

func assertIDs[T comparable](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
