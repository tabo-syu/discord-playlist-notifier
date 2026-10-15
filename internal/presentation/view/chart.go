package view

import (
	"fmt"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

const (
	CHART_WIDTH  = 960
	CHART_HEIGHT = 480
	// The bundled font has no Japanese glyphs, so the labels are in English
	CHART_TIME_FORMAT = "01/02"
	// Used when the chart spans less than CHART_SHORT_SPAN, where many ticks
	// fall on the same day
	CHART_SHORT_TIME_FORMAT = "01/02 15:04"
	CHART_SHORT_SPAN        = 5 * 24 * time.Hour
)

// timeFormatter formats an axis value as a date in loc, with the time when
// the chart spans from from to to is short.
func timeFormatter(from, to time.Time, loc *time.Location) chart.ValueFormatter {
	layout := CHART_TIME_FORMAT
	if to.Sub(from) < CHART_SHORT_SPAN {
		layout = CHART_SHORT_TIME_FORMAT
	}

	return func(v any) string {
		if f, ok := v.(float64); ok {
			return chart.TimeFromFloat64(f).In(loc).Format(layout)
		}
		return ""
	}
}

// compactNumber formats an axis value, e.g. 1.23M or 45.6K.
func compactNumber(v any) string {
	f, ok := v.(float64)
	if !ok {
		return ""
	}
	switch {
	case f >= 1_000_000:
		return fmt.Sprintf("%.2fM", f/1_000_000)
	case f >= 10_000:
		return fmt.Sprintf("%.1fK", f/1_000)
	default:
		return fmt.Sprintf("%.0f", f)
	}
}
