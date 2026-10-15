package view

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/wcharczuk/go-chart/v2"
)

// TrendingChart draws the views each video gained since its first recorded
// state as a PNG image, one line per video. The lines are named #1, #2, ...
// in the order of the videos, which should match the ranking in the message.
func TrendingChart(videos []*library.TrendingVideo, loc *time.Location) ([]byte, error) {
	var series []chart.Series
	var from, to time.Time
	for i, v := range videos {
		if len(v.Snapshots) < 2 {
			return nil, fmt.Errorf("at least 2 snapshots are needed, got %d for %s", len(v.Snapshots), v.Video.YoutubeID)
		}

		var times []time.Time
		var gained []float64
		for _, s := range v.Snapshots {
			times = append(times, s.RecordedAt)
			// Views can go down when YouTube removes invalid ones
			gained = append(gained, float64(s.Views)-float64(v.Snapshots[0].Views))
		}
		if first := v.Snapshots[0].RecordedAt; from.IsZero() || first.Before(from) {
			from = first
		}
		if last := v.Snapshots[len(v.Snapshots)-1].RecordedAt; last.After(to) {
			to = last
		}
		series = append(series, chart.TimeSeries{Name: fmt.Sprintf("#%d", i+1), XValues: times, YValues: gained})
	}
	if len(series) == 0 {
		return nil, fmt.Errorf("no videos to draw")
	}

	graph := chart.Chart{
		Width:  CHART_WIDTH,
		Height: CHART_HEIGHT,
		Background: chart.Style{
			Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20},
		},
		XAxis: chart.XAxis{ValueFormatter: timeFormatter(from, to, loc)},
		YAxis: chart.YAxis{
			ValueFormatter: func(v any) string { return "+" + compactNumber(v) },
		},
		Series: series,
	}
	graph.Elements = []chart.Renderable{chart.LegendThin(&graph)}

	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
