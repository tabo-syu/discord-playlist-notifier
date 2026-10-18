package view

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const (
	// グラフの下の表の 1 行の高さ
	TRENDING_ROW_HEIGHT = 30
	// 表に出す曲名の長さ（文字数）
	TRENDING_TABLE_TITLE_LENGTH = 36
)

// trendingLine は 1 曲分の線と、値札・表に書く値。
type trendingLine struct {
	video  *library.Video
	color  drawing.Color
	times  []time.Time
	gained []float64
	total  library.ViewCount
}

// TrendingChart は各動画が最初に記録された状態から伸ばした再生数を、動画
// ごとに 1 本の線で PNG 画像として描く。線の端に伸びの値札を付け、グラフの
// 下に順位・曲名・伸び・今の再生数の表を描く。表の順はメッセージの順位と
// 一致させること。
func TrendingChart(title string, videos []*library.TrendingVideo, loc *time.Location) ([]byte, error) {
	var lines []*trendingLine
	var from, to time.Time
	maxGained := 0.0
	for i, v := range videos {
		if len(v.Snapshots) < 2 {
			return nil, fmt.Errorf("at least 2 snapshots are needed, got %d for %s", len(v.Snapshots), v.Video.YoutubeID)
		}

		l := &trendingLine{video: v.Video, color: chartLines[i%len(chartLines)], total: v.Snapshots[len(v.Snapshots)-1].Views}
		for _, s := range v.Snapshots {
			l.times = append(l.times, s.RecordedAt)
			// YouTube が不正な再生を取り除くと、再生数が減ることがある
			l.gained = append(l.gained, float64(s.Views)-float64(v.Snapshots[0].Views))
			maxGained = max(maxGained, l.gained[len(l.gained)-1])
		}
		if from.IsZero() || l.times[0].Before(from) {
			from = l.times[0]
		}
		if at := l.times[len(l.times)-1]; at.After(to) {
			to = at
		}
		lines = append(lines, l)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no videos to draw")
	}

	graph := newChart(title, from, to, loc)
	tableHeight := 60 + TRENDING_ROW_HEIGHT*len(lines)
	graph.Height = CHART_HEIGHT + tableHeight - 20
	// 表が凡例を兼ねるので、上の凡例の分は空けない
	graph.Background.Padding.Top = 70
	graph.Background.Padding.Bottom = tableHeight
	// 値札と表に値を書くので、軸は目盛りの線だけにする
	ticks := niceTicks(maxGained, func(float64) string { return "" })
	top := ticks[len(ticks)-1].Value
	graph.YAxis.Ticks = ticks
	graph.YAxis.Range = &chart.ContinuousRange{Min: 0, Max: top}
	for _, l := range lines {
		graph.Series = append(graph.Series, starLine("", l.color, l.times, l.gained, chart.YAxisPrimary)...)
	}
	graph.Elements = []chart.Renderable{trendingTags(lines, from, to, top), trendingTable(lines)}

	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// trendingTags は各線の最後の点の左上（グラフの内側）に「+1,078万」の値札を
// 描く。右に値札の余白を取らずに済むように内側に置く。上に収まらない値札は
// グラフの上端に寄せ、重なりそうな値札は下にずらす。go-chart の値札は点の右に
// しか置けないので自前で描く。
func trendingTags(lines []*trendingLine, from, to time.Time, top float64) chart.Renderable {
	return func(r chart.Renderer, canvas chart.Box, _ chart.Style) {
		const height, gap, padX = 22, 4, 8
		type tag struct {
			color   drawing.Color
			text    string
			x, y, w int
		}

		r.SetFont(chartFont)
		r.SetFontSize(12)
		span := to.Sub(from).Seconds()
		var tags []tag
		for _, l := range lines {
			at, gained := l.times[len(l.times)-1], l.gained[len(l.gained)-1]
			x := canvas.Left + int(float64(canvas.Width())*at.Sub(from).Seconds()/span)
			y := canvas.Bottom - int(float64(canvas.Height())*gained/top)
			text := "+" + ShortCount(gained)
			w := r.MeasureText(text).Width() + padX*2
			t := tag{l.color, text, x - 14 - w, y - height - 8, w}
			// 線は左下から点に向かってくるので、下に回すと線にかぶる。上端に寄せる
			t.y = max(t.y, canvas.Top+2)
			tags = append(tags, t)
		}
		sort.Slice(tags, func(i, j int) bool { return tags[i].y < tags[j].y })
		for i := 1; i < len(tags); i++ {
			tags[i].y = max(tags[i].y, tags[i-1].y+height+gap)
		}

		for _, t := range tags {
			chart.Draw.Box(r, chart.Box{Left: t.x, Top: t.y, Right: t.x + t.w, Bottom: t.y + height},
				chart.Style{FillColor: t.color, StrokeColor: t.color})
			// Draw.Box がフォントを書き換えるので、毎回設定し直す
			r.SetFont(chartFont)
			r.SetFontSize(12)
			r.SetFontColor(chartBackground)
			r.Text(t.text, t.x+padX, t.y+16)
		}
	}
}

// trendingTable はグラフの下に、線の色・順位・曲名・この期間の伸び・今の
// 再生数の表を描く。凡例も兼ねる。
func trendingTable(lines []*trendingLine) chart.Renderable {
	return func(r chart.Renderer, canvas chart.Box, _ chart.Style) {
		left, right := canvas.Left, canvas.Right-6
		gainedRight := right - 170
		top := canvas.Bottom + 50
		rightAligned := func(text string, x, y int) {
			r.Text(text, x-r.MeasureText(text).Width(), y)
		}

		r.SetFont(chartFont)
		r.SetFontSize(10)
		r.SetFontColor(chartText.WithAlpha(150))
		r.Text("曲", left+56, top)
		rightAligned("この期間の伸び", gainedRight, top)
		rightAligned("今の再生数", right, top)

		for i, l := range lines {
			y := top + 8 + TRENDING_ROW_HEIGHT*(i+1)
			// 1 行おきに帯を敷く
			if i%2 == 0 {
				chart.Draw.Box(r, chart.Box{Left: left, Top: y - TRENDING_ROW_HEIGHT + 9, Right: right + 6, Bottom: y + 9},
					chart.Style{FillColor: chartCanvas, StrokeColor: chartCanvas})
			}
			r.SetStrokeColor(l.color)
			r.SetStrokeWidth(4)
			r.MoveTo(left+8, y-5)
			r.LineTo(left+30, y-5)
			r.Stroke()

			// Draw.Box がフォントを書き換えるので、毎回設定し直す
			r.SetFont(chartFont)
			r.SetFontSize(13)
			r.SetFontColor(l.color)
			r.Text(fmt.Sprintf("%d.", i+1), left+36, y)
			r.SetFontColor(chartText)
			r.Text(truncate(l.video.Title, TRENDING_TABLE_TITLE_LENGTH), left+56, y)
			rightAligned("+"+ShortCount(l.gained[len(l.gained)-1]), gainedRight, y)
			rightAligned(ShortCount(l.total)+"回", right, y)
		}
	}
}
