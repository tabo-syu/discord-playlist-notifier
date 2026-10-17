package view

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/wcharczuk/go-chart/v2"
)

// HistoryChart はスナップショットの再生数（右軸）と高評価数（左軸）を
// PNG 画像として描く。スナップショットは 2 つ以上必要。
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

	graph := chart.Chart{
		Width:  CHART_WIDTH,
		Height: CHART_HEIGHT,
		Background: chart.Style{
			Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20},
		},
		XAxis: chart.XAxis{ValueFormatter: timeFormatter(times[0], times[len(times)-1], loc)},
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

// paddedRange は値の上下に余白を取り、全ての値が同じでも範囲が
// 空にならないようにする。
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
