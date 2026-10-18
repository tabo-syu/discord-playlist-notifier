package view

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/wcharczuk/go-chart/v2"
)

// 値札（「+1,079万 ・ 計2.3億回」）が収まるように右に空ける幅
const TRENDING_LABEL_SPACE = 150

// TrendingChart は各動画が最初に記録された状態から伸ばした再生数を、動画
// ごとに 1 本の線で PNG 画像として描く。凡例は動画の順に「1. 曲名」で、
// メッセージの順位と一致させること。線の端には伸びと今の再生数を書く。
func TrendingChart(title string, videos []*library.TrendingVideo, loc *time.Location) ([]byte, error) {
	var series []chart.Series
	var labels []chart.Value2
	var from, to time.Time
	maxGained := 0.0
	for i, v := range videos {
		if len(v.Snapshots) < 2 {
			return nil, fmt.Errorf("at least 2 snapshots are needed, got %d for %s", len(v.Snapshots), v.Video.YoutubeID)
		}

		var times []time.Time
		var gained []float64
		for _, s := range v.Snapshots {
			times = append(times, s.RecordedAt)
			// YouTube が不正な再生を取り除くと、再生数が減ることがある
			gained = append(gained, float64(s.Views)-float64(v.Snapshots[0].Views))
			maxGained = max(maxGained, gained[len(gained)-1])
		}
		if from.IsZero() || times[0].Before(from) {
			from = times[0]
		}
		last := v.Snapshots[len(v.Snapshots)-1]
		if last.RecordedAt.After(to) {
			to = last.RecordedAt
		}

		color := chartLines[i%len(chartLines)]
		series = append(series, starLine(fmt.Sprintf("%d. %s", i+1, truncate(v.Video.Title, CHART_LEGEND_TITLE_LENGTH)), color, times, gained, chart.YAxisPrimary)...)
		labels = append(labels, endLabel(color, last.RecordedAt, gained[len(gained)-1],
			fmt.Sprintf("+%s ・ 計%s回", ShortCount(gained[len(gained)-1]), ShortCount(last.Views))))
	}
	if len(series) == 0 {
		return nil, fmt.Errorf("no videos to draw")
	}

	graph := newChart(title, from, to, loc)
	graph.Background.Padding.Right = TRENDING_LABEL_SPACE
	// 値札に値を書くので、軸は目盛りの線だけにする
	ticks := niceTicks(maxGained, func(float64) string { return "" })
	graph.YAxis.Ticks = ticks
	graph.YAxis.Range = &chart.ContinuousRange{Min: 0, Max: ticks[len(ticks)-1].Value}
	graph.Series = append(series, chart.AnnotationSeries{Annotations: labels})
	graph.Elements = []chart.Renderable{legend(series)}

	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
