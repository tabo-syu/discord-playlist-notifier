package view

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// go test ./internal/presentation/view/ -update で期待する出力を書き直す
var update = flag.Bool("update", false, "update golden files")

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

func at(month time.Month, day, hour int) time.Time {
	return time.Date(2026, month, day, hour, 0, 0, 0, time.UTC)
}

func video(id, title string, views library.ViewCount) *library.Video {
	return &library.Video{YoutubeID: library.VideoID(id), Title: title, Views: views}
}

func added(v *library.Video, addedAt time.Time) *library.PlaylistVideo {
	return &library.PlaylistVideo{
		PlaylistID: "PLxxxx",
		Item:       &library.PlaylistItem{VideoYoutubeID: v.YoutubeID, AddedAt: addedAt},
		Video:      v,
	}
}

func snapshot(at time.Time, views library.ViewCount, likes, comments uint64) *library.VideoSnapshot {
	return &library.VideoSnapshot{RecordedAt: at, Views: views, Likes: likes, Comments: comments}
}

var (
	// 角括弧は Markdown のリンクを壊すので、置き換わることを確かめる
	song1 = video("aaaaaaaaaaa", "夜に駆ける [Official]", 1_234_567_890)
	song2 = video("bbbbbbbbbbb", "群青", 98_765)
	song3 = video("ccccccccccc", "怪物", 0)
)

func TestText(t *testing.T) {
	tests := map[string]string{
		"summary_empty": Summary("お気に入り", &library.Stats{}, jst),
		"summary_no_ranking": Summary("お気に入り", &library.Stats{
			Total: 3, Available: 1, Private: 1, Deleted: 1,
			AddedThisMonth: 0, AddedLast30Days: 2,
			// UTC では前の日なので、タイムゾーンを変換していることを確かめる
			FirstAddedAt: at(time.January, 1, 20),
			Monthly: []library.MonthCount{
				{Month: time.Date(2026, time.September, 1, 0, 0, 0, 0, jst), Count: 0},
				{Month: time.Date(2026, time.October, 1, 0, 0, 0, 0, jst), Count: 0},
			},
		}, jst),
		"summary": Summary("お気に入り", &library.Stats{
			Total: 1234, Available: 1200, Private: 30, Deleted: 4,
			AddedThisMonth: 5, AddedLast30Days: 12,
			FirstAddedAt: at(time.March, 4, 3),
			TopViewed:    []*library.Video{song1, song2, song3},
			Monthly: []library.MonthCount{
				{Month: time.Date(2026, time.August, 1, 0, 0, 0, 0, jst), Count: 1},
				{Month: time.Date(2026, time.September, 1, 0, 0, 0, 0, jst), Count: 7},
				{Month: time.Date(2026, time.October, 1, 0, 0, 0, 0, jst), Count: 0},
			},
		}, jst),

		"trending_not_enough": Trending(&library.Trending{From: at(time.October, 1, 0), To: at(time.October, 1, 0)}, jst),
		"trending_no_videos":  Trending(&library.Trending{From: at(time.September, 18, 0), To: at(time.October, 17, 18)}, jst),
		"trending": Trending(&library.Trending{
			From: at(time.September, 18, 0), To: at(time.October, 17, 18),
			Videos: []*library.TrendingVideo{
				{Video: song1, Increase: 12_345, Snapshots: []*library.VideoSnapshot{snapshot(at(time.September, 18, 0), 1_000, 0, 0), snapshot(at(time.October, 17, 18), 13_345, 0, 0)}},
				{Video: song2, Increase: 7, Snapshots: []*library.VideoSnapshot{snapshot(at(time.October, 17, 18), 98_765, 0, 0)}},
			},
		}, jst),

		"history_not_enough": History(&library.VideoHistory{Video: song1, Snapshots: []*library.VideoSnapshot{snapshot(at(time.October, 1, 0), 1, 1, 1)}}, 30),
		"history": History(&library.VideoHistory{Video: song2, Snapshots: []*library.VideoSnapshot{
			snapshot(at(time.October, 1, 0), 1_000, 500, 12),
			snapshot(at(time.October, 2, 0), 1_100, 600, 12),
			// コメントが減る場合
			snapshot(at(time.October, 3, 0), 2_345, 600, 9),
		}}, 7),

		"random_empty": Random(nil),
		"random":       Random([]*library.PlaylistVideo{added(song1, at(time.May, 1, 0)), added(song2, at(time.June, 1, 0))}),

		"wrapped_empty": Wrapped(&library.Wrapped{Year: 2026, PreviousYearAdded: 3}, jst, InMessage),
		"wrapped_no_ranking": Wrapped(&library.Wrapped{
			Year: 2026, Added: 1, PreviousYearAdded: 0,
			BusiestMonth: library.MonthCount{Month: time.Date(2026, time.February, 1, 0, 0, 0, 0, jst), Count: 1},
		}, jst, InMessage),
		"wrapped_message": wrapped(InMessage),
		"wrapped_embed":   wrapped(InEmbed),
	}

	for name, got := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name+".golden")
			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func wrapped(target Target) string {
	return Wrapped(&library.Wrapped{
		Year: 2026, Added: 1500, PreviousYearAdded: 1200,
		BusiestMonth: library.MonthCount{Month: time.Date(2026, time.December, 1, 0, 0, 0, 0, jst), Count: 230},
		TopViewed:    []*library.Video{song1, song2},
		// UTC では 12/31 なので、タイムゾーンを変換していることを確かめる
		First: added(song3, time.Date(2025, time.December, 31, 15, 30, 0, 0, time.UTC)),
		Last:  added(song1, at(time.December, 31, 14)),
	}, jst, target)
}
