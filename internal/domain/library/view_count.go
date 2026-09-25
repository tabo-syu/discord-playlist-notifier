package library

// ViewCount is the number of views of a video.
type ViewCount uint64

// View counts that are worth a notification
var ViewMilestones = []ViewCount{
	100_000,
	500_000,
	1_000_000,
	5_000_000,
	10_000_000,
	50_000_000,
	100_000_000,
}

// Milestone returns the highest milestone the views have reached, or 0.
func (c ViewCount) Milestone() ViewCount {
	var reached ViewCount
	for _, m := range ViewMilestones {
		if c >= m {
			reached = m
		}
	}

	return reached
}
