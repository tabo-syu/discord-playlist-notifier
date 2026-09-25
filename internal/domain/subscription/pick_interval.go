package subscription

import "time"

// PickInterval is how often a random video of the playlist is posted.
type PickInterval string

const (
	PickNone    PickInterval = ""
	PickDaily   PickInterval = "daily"
	PickWeekly  PickInterval = "weekly"
	PickMonthly PickInterval = "monthly"
)

// DueOn reports whether a pick with the interval is posted on the day of t.
// Weekly picks are posted on Mondays and monthly picks on the 1st.
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
