package library

// ViewCount は動画の再生数。
type ViewCount uint64

// 通知する価値のある再生数
var ViewMilestones = []ViewCount{
	100_000,
	1_000_000,
	10_000_000,
	100_000_000,
}

// Milestone は再生数が達した最も高い節目を返す。なければ 0。
func (c ViewCount) Milestone() ViewCount {
	var reached ViewCount
	for _, m := range ViewMilestones {
		if c >= m {
			reached = m
		}
	}

	return reached
}
