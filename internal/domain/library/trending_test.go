package library_test

import (
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestComputeTrending(t *testing.T) {
	base := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	at := func(hours int) time.Time { return base.Add(time.Duration(hours) * time.Hour) }
	snap := func(hours int, views library.ViewCount) *library.VideoSnapshot {
		return &library.VideoSnapshot{RecordedAt: at(hours), Views: views}
	}

	videos := map[library.VideoID]*library.Video{
		"top":       available("top", 0),
		"tie-a":     available("tie-a", 0),
		"tie-b":     available("tie-b", 0),
		"private":   hidden("private", library.PrivacyPrivate),
		"once":      available("once", 0),
		"decreased": available("decreased", 0),
		"flat":      available("flat", 0),
	}
	earliest := map[library.VideoID]*library.VideoSnapshot{
		"top":       snap(6, 100),
		"tie-a":     snap(6, 0),
		"tie-b":     snap(6, 500),
		"private":   snap(0, 0),
		"once":      snap(6, 10),
		"decreased": snap(6, 1_000),
		"flat":      snap(6, 1_000),
		// プレイリストにもう無い動画
		"gone": snap(6, 0),
	}
	latest := map[library.VideoID]*library.VideoSnapshot{
		"top":     snap(12, 3_100),
		"tie-a":   snap(12, 1_000),
		"tie-b":   snap(18, 1_500),
		"private": snap(12, 1_000_000),
		// 1 回しか記録されていない
		"once":      snap(6, 10),
		"decreased": snap(12, 900),
		"flat":      snap(12, 1_000),
		"gone":      snap(24, 1_000_000),
	}

	got := library.ComputeTrending(videos, earliest, latest, 3)

	// 期間は順位に入らない動画も含めて決める
	if !got.From.Equal(at(0)) || !got.To.Equal(at(24)) {
		t.Errorf("From, To = %v, %v, want %v, %v", got.From, got.To, at(0), at(24))
	}
	var ids []library.VideoID
	for _, v := range got.Videos {
		ids = append(ids, v.Video.YoutubeID)
	}
	// 伸びが同じなら ID の順
	assertIDs(t, "Videos", ids, []library.VideoID{"top", "tie-a", "tie-b"})
	if got.Videos[0].Increase != 3_000 {
		t.Errorf("Increase = %d, want 3000", got.Videos[0].Increase)
	}

	if got := library.ComputeTrending(videos, earliest, latest, 1); len(got.Videos) != 1 {
		t.Errorf("ComputeTrending(n=1) returned %d videos, want 1", len(got.Videos))
	}
}

func TestComputeTrendingEmpty(t *testing.T) {
	got := library.ComputeTrending(nil, nil, nil, 3)

	if !got.From.IsZero() || !got.To.IsZero() || len(got.Videos) != 0 {
		t.Errorf("ComputeTrending() = %+v, want zero values", got)
	}
}
