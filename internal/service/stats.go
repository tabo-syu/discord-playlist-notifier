package service

import (
	"sort"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

const (
	STATS_TOP_VIEWED = 3
	STATS_MONTHS     = 6
)

type MonthCount struct {
	Month time.Time
	Count int
}

type PlaylistStats struct {
	Total           int
	Available       int
	Private         int
	Deleted         int
	AddedLast30Days int
	AddedThisMonth  int
	FirstAddedAt    time.Time
	TopViewed       []*domain.Video
	// The last STATS_MONTHS months including the current one, oldest first
	Monthly []MonthCount
}

// ComputeStats aggregates the contents of a playlist. Every item is counted,
// so a video added twice counts twice, except in TopViewed.
func ComputeStats(contents []*PlaylistVideo, now time.Time, loc *time.Location) *PlaylistStats {
	stats := &PlaylistStats{}
	now = now.In(loc)
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	firstMonth := thisMonth.AddDate(0, -(STATS_MONTHS - 1), 0)

	monthly := map[time.Time]int{}
	viewed := map[string]*domain.Video{}
	for _, c := range contents {
		stats.Total++
		addedAt := c.Item.AddedAt.In(loc)
		if stats.FirstAddedAt.IsZero() || addedAt.Before(stats.FirstAddedAt) {
			stats.FirstAddedAt = addedAt
		}
		if now.Sub(addedAt) <= 30*24*time.Hour {
			stats.AddedLast30Days++
		}
		if !addedAt.Before(thisMonth) {
			stats.AddedThisMonth++
		}
		if !addedAt.Before(firstMonth) {
			monthly[time.Date(addedAt.Year(), addedAt.Month(), 1, 0, 0, 0, 0, loc)]++
		}

		switch v := c.Video; {
		case v.Available():
			stats.Available++
			viewed[v.YoutubeID] = v
		case v.PrivacyStatus == domain.PrivacyPrivate:
			stats.Private++
		case v.PrivacyStatus == domain.PrivacyDeleted:
			stats.Deleted++
		}
	}

	for _, v := range viewed {
		stats.TopViewed = append(stats.TopViewed, v)
	}
	sort.Slice(stats.TopViewed, func(i, j int) bool {
		a, b := stats.TopViewed[i], stats.TopViewed[j]
		if a.Views != b.Views {
			return a.Views > b.Views
		}
		return a.YoutubeID < b.YoutubeID
	})
	if len(stats.TopViewed) > STATS_TOP_VIEWED {
		stats.TopViewed = stats.TopViewed[:STATS_TOP_VIEWED]
	}

	for m := firstMonth; !m.After(thisMonth); m = m.AddDate(0, 1, 0) {
		stats.Monthly = append(stats.Monthly, MonthCount{Month: m, Count: monthly[m]})
	}

	return stats
}
