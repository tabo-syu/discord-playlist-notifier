package view

import (
	"fmt"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

const (
	CHART_WIDTH  = 960
	CHART_HEIGHT = 480
	// 同梱のフォントには日本語のグリフがないので、ラベルは英語にする
	CHART_TIME_FORMAT = "01/02"
	// グラフの期間が CHART_SHORT_SPAN より短く、多くの目盛りが同じ日に
	// 重なるときに使う
	CHART_SHORT_TIME_FORMAT = "01/02 15:04"
	CHART_SHORT_SPAN        = 5 * 24 * time.Hour
)

// timeFormatter は軸の値を loc での日付として整形する。グラフの期間
// （from から to）が短いときは時刻も付ける。
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

// compactNumber は軸の値を整形する。例: 1.23M、45.6K
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
