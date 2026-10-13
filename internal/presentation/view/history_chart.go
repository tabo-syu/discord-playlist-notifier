package view

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/wcharczuk/go-chart/v2"
)

const (
	CHART_WIDTH  = 960
	CHART_HEIGHT = 480
	// The bundled font has no Japanese glyphs, so the labels are in English
	CHART_TIME_FORMAT = "01/02"
)

// HistoryChart draws the views (right axis) and likes (left axis) of the
// snapshots as a PNG image. At least 2 snapshots are needed.
func HistoryChart(snapshots []*library.VideoSnapshot, loc *time.Location) ([]byte, error) {
	if len(snapshots) < 2 {
		return nil, fmt.Errorf("at least 2 snapshots are needed, got %d", len(snapshots))
	}

	var times []time.Time
	var views, likes []float64
	for _, s := range snapshots {
		times = append(times, s.RecordedAt)
		views = append(views, float64(s.Views))
		likes = append(likes, float64(s.Likes))
	}

	formatTime := func(v any) string {
		if f, ok := v.(float64); ok {
			return chart.TimeFromFloat64(f).In(loc).Format(CHART_TIME_FORMAT)
		}
		return ""
	}

	graph := chart.Chart{
		Width:  CHART_WIDTH,
		Height: CHART_HEIGHT,
		Background: chart.Style{
			Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20},
		},
		XAxis: chart.XAxis{ValueFormatter: formatTime},
		YAxis: chart.YAxis{
			ValueFormatter: compactNumber,
			Range:          paddedRange(views),
		},
		YAxisSecondary: chart.YAxis{
			ValueFormatter: compactNumber,
			Range:          paddedRange(likes),
		},
		Series: []chart.Series{
			chart.TimeSeries{Name: "Views", XValues: times, YValues: views},
			chart.TimeSeries{Name: "Likes", YAxis: chart.YAxisSecondary, XValues: times, YValues: likes},
		},
	}
	graph.Elements = []chart.Renderable{chart.LegendThin(&graph)}

	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// paddedRange leaves room above and below the values, and keeps the range
// from being empty when every value is the same.
func paddedRange(values []float64) *chart.ContinuousRange {
	lo, hi := values[0], values[0]
	for _, v := range values {
		lo, hi = min(lo, v), max(hi, v)
	}
	pad := (hi - lo) * 0.1
	if pad == 0 {
		pad = max(hi*0.01, 1)
	}

	return &chart.ContinuousRange{Min: max(lo-pad, 0), Max: hi + pad}
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
