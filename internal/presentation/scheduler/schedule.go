package scheduler

import (
	"log"
	"runtime/debug"
)

// Local time (in the scheduler location) to post the picks. Weekly picks are
// posted on Mondays and monthly picks on the 1st.
const PICK_TIME = "12:00"

// Cron expression (in the scheduler location) to post the yearly summary: Dec 31 21:00
const WRAPPED_CRON = "0 21 31 12 *"

// guard keeps a panic in one run of a job from bringing the whole bot down.
func guard(name string, job func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in %s: %v\n%s", name, r, debug.Stack())
			}
		}()

		job()
	}
}
