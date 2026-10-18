package library

import "time"

const WRAPPED_TOP = 5

// Wrapped はある 1 年にプレイリストへ追加されたもののまとめ。
type Wrapped struct {
	Year              int
	Added             int
	PreviousYearAdded int
	BusiestMonth      MonthCount
	// その年に追加された視聴可能な動画のうち、再生数の多いもの
	TopViewed []*Video
	First     *PlaylistVideo
	Last      *PlaylistVideo
}

// ComputeWrapped は指定した年（loc での年）に追加されたアイテムを集計する。
// contents は Contents が返すとおり、追加された時刻の順に並んでいること。
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
