package library

import "time"

const WRAPPED_TOP = 5

// Wrapped summarizes what was added to a playlist in one year.
type Wrapped struct {
	Year              int
	Added             int
	PreviousYearAdded int
	BusiestMonth      MonthCount
	// The most viewed available videos added in the year
	TopViewed []*Video
	First     *PlaylistVideo
	Last      *PlaylistVideo
}

// ComputeWrapped aggregates the items added in the given year (in loc).
// contents must be ordered by the time they were added, as Contents returns them.
func ComputeWrapped(contents []*PlaylistVideo, year int, loc *time.Location) *Wrapped {
	w := &Wrapped{Year: year}
	monthly := map[time.Month]int{}
	viewed := map[VideoID]*Video{}
	for _, c := range contents {
		addedAt := c.Item.AddedAt.In(loc)
		if addedAt.Year() == year-1 {
			w.PreviousYearAdded++
		}
		if addedAt.Year() != year {
			continue
		}

		w.Added++
		monthly[addedAt.Month()]++

		v := c.Video
		if !v.Available() {
			continue
		}
		viewed[v.YoutubeID] = v
		if w.First == nil {
			w.First = c
		}
		w.Last = c
	}

	for m := time.January; m <= time.December; m++ {
		if monthly[m] > w.BusiestMonth.Count {
			w.BusiestMonth = MonthCount{Month: time.Date(year, m, 1, 0, 0, 0, 0, loc), Count: monthly[m]}
		}
	}
	w.TopViewed = rankViewed(viewed, WRAPPED_TOP)

	return w
}
