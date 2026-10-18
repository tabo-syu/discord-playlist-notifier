package view

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const (
	CHART_WIDTH  = 960
	CHART_HEIGHT = 520
	// 目盛りのおおよその数
	CHART_TICKS = 5
	// 凡例とタイトルに出す曲名の長さ（文字数）
	CHART_LEGEND_TITLE_LENGTH = 18
	CHART_TITLE_LENGTH        = 40
	CHART_TIME_FORMAT         = "1/2"
	// グラフの期間が CHART_SHORT_SPAN より短く、多くの目盛りが同じ日に
	// 重なるときに使う
	CHART_SHORT_TIME_FORMAT = "1/2 15時"
	CHART_SHORT_SPAN        = 5 * 24 * time.Hour
)

// グラフの中の文字に使う丸ゴシック（M PLUS Rounded 1c、OFL。fonts/OFL.txt）。
// go-chart に同梱のフォントには日本語がない。
//
//go:embed fonts/MPLUSRounded1c-Medium.ttf
var chartFontTTF []byte

var chartFont = mustParseFont(chartFontTTF)

// 夜空の色。線の色は順位の順に使う。
var (
	chartBackground = drawing.ColorFromHex("141C33")
	chartCanvas     = drawing.ColorFromHex("18213D")
	chartText       = drawing.ColorFromHex("E8EEFF")
	chartGrid       = drawing.ColorFromHex("2A3760")
	chartLines      = []drawing.Color{
		drawing.ColorFromHex("FFD36E"),
		drawing.ColorFromHex("FF9ECD"),
		drawing.ColorFromHex("8FE3FF"),
	}
)

func mustParseFont(ttf []byte) *truetype.Font {
	font, err := truetype.Parse(ttf)
	if err != nil {
		panic(fmt.Sprintf("bundled chart font is broken: %v", err))
	}
	return font
}

// newChart は夜空の見た目のグラフの土台を返す。系列と凡例は呼び出し側で足す。
func newChart(title string, from, to time.Time, loc *time.Location) chart.Chart {
	axis := chart.Style{FontColor: chartText, Font: chartFont, FontSize: 11, StrokeColor: chartGrid}
	return chart.Chart{
		Title:      title,
		TitleStyle: chart.Style{FontColor: chartText, Font: chartFont, FontSize: 18},
		Font:       chartFont,
		Width:      CHART_WIDTH,
		Height:     CHART_HEIGHT,
		Background: chart.Style{FillColor: chartBackground, Padding: chart.Box{Top: 115, Left: 30, Right: 30, Bottom: 30}},
		Canvas:     chart.Style{FillColor: chartCanvas},
		XAxis:      chart.XAxis{Style: axis, ValueFormatter: timeFormatter(from, to, loc)},
		YAxis:      chart.YAxis{Style: axis, GridMajorStyle: chart.Style{StrokeColor: chartGrid, StrokeWidth: 1.5}},
	}
}

// starLine は 1 本の線を、下に重ねる光と最後の点の丸と一緒に返す。光には
// 名前を付けないので、凡例には出ない。
func starLine(name string, color drawing.Color, times []time.Time, values []float64, yAxis chart.YAxisType) []chart.Series {
	last := len(times) - 1
	halo := chart.TimeSeries{XValues: times, YValues: values, YAxis: yAxis,
		Style: chart.Style{StrokeColor: color.WithAlpha(60), StrokeWidth: 10}}
	line := chart.TimeSeries{Name: name, XValues: times, YValues: values, YAxis: yAxis,
		Style: chart.Style{StrokeColor: color, StrokeWidth: 4, DotColor: color,
			DotWidthProvider: func(_, _ chart.Range, index int, _, _ float64) float64 {
				if index == last {
					return 9
				}
				return 0
			}}}

	return []chart.Series{halo, line}
}

// endLabel は線の最後の点に付ける値札を返す。
func endLabel(color drawing.Color, at time.Time, value float64, label string) chart.Value2 {
	return chart.Value2{XValue: chart.TimeToFloat64(at), YValue: value, Label: label,
		Style: chart.Style{FillColor: color, StrokeColor: color, FontColor: chartBackground, Font: chartFont, FontSize: 11}}
}

// legend は名前のある系列を「線の色 名前」の形で、タイトルの下に枠なしで
// 横一列に並べる。値札の分だけ描画域が狭くなるので、画像の幅で中央に寄せる。
// go-chart の LegendThin は枠の色の指定を無視するので使わない。
func legend(series []chart.Series) chart.Renderable {
	return func(r chart.Renderer, canvas chart.Box, _ chart.Style) {
		r.SetFont(chartFont)
		r.SetFontSize(12)
		const swatch, gap, between = 22, 6, 28
		var named []chart.TimeSeries
		var widths []int
		total := 0
		for _, s := range series {
			ts, ok := s.(chart.TimeSeries)
			if !ok || ts.Name == "" {
				continue
			}
			w := r.MeasureText(ts.Name).Width()
			named = append(named, ts)
			widths = append(widths, w)
			total += swatch + gap + w
		}
		total += between * max(len(named)-1, 0)

		x := max((CHART_WIDTH-total)/2, 10)
		y := canvas.Top - 30
		for i, ts := range named {
			r.SetStrokeColor(ts.Style.StrokeColor)
			r.SetStrokeWidth(4)
			r.MoveTo(x, y-4)
			r.LineTo(x+swatch, y-4)
			r.Stroke()
			r.SetFontColor(chartText)
			r.Text(ts.Name, x+swatch+gap, y)
			x += swatch + gap + widths[i] + between
		}
	}
}

// niceTicks は 0 から maxValue までを、1・2・2.5・5 × 10 のべき乗の切りの
// いい間隔で区切った目盛りを返す。label が空を返す目盛りは線だけになる。
func niceTicks(maxValue float64, label func(float64) string) []chart.Tick {
	raw := max(maxValue, 1) / CHART_TICKS
	pow := math.Pow(10, math.Floor(math.Log10(raw)))
	step := 10 * pow
	for _, m := range []float64{1, 2, 2.5, 5} {
		if raw <= m*pow {
			step = m * pow
			break
		}
	}

	var ticks []chart.Tick
	for v := 0.0; v < maxValue+step; v += step {
		ticks = append(ticks, chart.Tick{Value: v, Label: label(v)})
	}
	return ticks
}

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

// ShortCount は回数を万・億の単位で短く書く。例: 2.3億、1,079万、9,999
func ShortCount[T ~int | ~uint64 | ~float64](n T) string {
	f := float64(n)
	switch {
	case f >= 100_000_000:
		return strings.TrimSuffix(fmt.Sprintf("%.1f", f/100_000_000), ".0") + "億"
	case f >= 10_000:
		return Number(int(f/10_000)) + "万"
	default:
		return Number(int(f))
	}
}

// axisFormatter は lo から hi までの軸の値を、目盛りどうしが同じ表記に
// ならない桁数で、万・億の単位で書く。例: 6億9,442万 ではなく 69,442万
func axisFormatter(lo, hi float64) chart.ValueFormatter {
	unit, suffix := 1.0, ""
	switch {
	case hi-lo >= 100_000_000:
		unit, suffix = 100_000_000, "億"
	case hi >= 10_000:
		unit, suffix = 10_000, "万"
	}
	step := (hi - lo) / CHART_TICKS / unit
	decimals := 0
	if step > 0 && step < 1 {
		decimals = min(int(math.Ceil(-math.Log10(step))), 2)
	}

	return func(v any) string {
		f, _ := v.(float64)
		// 丸めてから整数部に桁区切りを付ける（509.996 を 510.00 にするため）
		whole, frac, _ := strings.Cut(fmt.Sprintf("%.*f", decimals, f/unit), ".")
		n, _ := strconv.Atoi(whole)
		if frac == "" {
			return Number(n) + suffix
		}
		return Number(n) + "." + frac + suffix
	}
}

// truncate は曲名を n 文字までに切り詰める。
func truncate(title string, n int) string {
	runes := []rune(title)
	if len(runes) <= n {
		return title
	}
	return string(runes[:n-1]) + "…"
}
