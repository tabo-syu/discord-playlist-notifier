package repository

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	cases := map[string]time.Duration{
		"PT4M13S":   4*time.Minute + 13*time.Second,
		"PT1H":      time.Hour,
		"PT1H2M3S":  time.Hour + 2*time.Minute + 3*time.Second,
		"P1DT2H":    26 * time.Hour,
		"PT45S":     45 * time.Second,
		"P0D":       0,
		"":          0,
		"not-a-dur": 0,
	}
	for in, want := range cases {
		if got := ParseDuration(in); got != want {
			t.Errorf("ParseDuration(%q) = %v, want %v", in, got, want)
		}
	}
}
