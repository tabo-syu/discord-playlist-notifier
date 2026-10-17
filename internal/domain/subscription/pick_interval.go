package subscription

import "time"

// PickInterval はプレイリストのランダムな動画を投稿する頻度。
type PickInterval string

const (
	PickNone    PickInterval = ""
	PickDaily   PickInterval = "daily"
	PickWeekly  PickInterval = "weekly"
	PickMonthly PickInterval = "monthly"
)

// DueOn はその間隔のピックを t の日に投稿するかを返す。
// 毎週のピックは月曜日、毎月のピックは 1 日に投稿する。
func (i PickInterval) DueOn(t time.Time) bool {
	switch i {
	case PickDaily:
		return true
	case PickWeekly:
		return t.Weekday() == time.Monday
	case PickMonthly:
		return t.Day() == 1
	default:
		return false
	}
}
