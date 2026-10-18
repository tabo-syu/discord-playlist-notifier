package view

import (
	"embed"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// メッセージの文言は templates/ のテンプレートに書き、ここでは
// テンプレートに渡すデータを用意するだけにする。中身がないときなど、
// メッセージの形ごと変わる場合分けはここで行い、別のテンプレートを使う。
//
//go:embed templates/*.md.tmpl
var templateFS embed.FS

var templates = template.Must(template.New("").Funcs(funcs(time.UTC, InMessage)).ParseFS(templateFS, "templates/*.md.tmpl"))

// funcs はテンプレートから呼べる関数。日付のタイムゾーンとリンクの書き方は
// 整形するたびに変わるので、render で差し替える。
func funcs(loc *time.Location, target Target) template.FuncMap {
	return template.FuncMap{
		"number": func(n any) (string, error) {
			switch n := n.(type) {
			case int:
				return Number(n), nil
			case uint64:
				return Number(n), nil
			case library.ViewCount:
				return Number(n), nil
			}
			return "", fmt.Errorf("number: unsupported type %T", n)
		},
		// rank は 0 始まりの添字を順位にする
		"rank":  func(i int) int { return i + 1 },
		"date":  func(layout string, t time.Time) string { return t.In(loc).Format(layout) },
		"link":  func(v *library.Video) string { return VideoLink(v, target) },
		"first": func(s []*library.VideoSnapshot) *library.VideoSnapshot { return s[0] },
		"last":  func(s []*library.VideoSnapshot) *library.VideoSnapshot { return s[len(s)-1] },
		"change": func(before, after any) (string, error) {
			b, err := count(before)
			if err != nil {
				return "", err
			}
			a, err := count(after)
			if err != nil {
				return "", err
			}
			return change(b, a), nil
		},
	}
}

func render(name string, data any, loc *time.Location, target Target) string {
	t := template.Must(templates.Clone()).Funcs(funcs(loc, target))

	var sb strings.Builder
	if err := t.ExecuteTemplate(&sb, name, data); err != nil {
		log.Println("Could not render text:", name, "cause:", err)
		return "エラー！システムに問題があります！"
	}

	return sb.String()
}

// Summary はプレイリストの曲数、追加ペース、再生数ランキングを整形する。
func Summary(title string, s *library.Stats, loc *time.Location) string {
	if s.Total == 0 {
		return render("summary_empty", title, loc, InMessage)
	}

	type bar struct {
		Month time.Time
		Bar   string
		Count int
	}

	max := 0
	for _, m := range s.Monthly {
		if m.Count > max {
			max = m.Count
		}
	}
	bars := make([]bar, 0, len(s.Monthly))
	for _, m := range s.Monthly {
		width := 0
		if max > 0 {
			width = (m.Count*20 + max - 1) / max
		}
		bars = append(bars, bar{m.Month, strings.Repeat("█", width), m.Count})
	}

	return render("summary", struct {
		Title string
		Stats *library.Stats
		Bars  []bar
	}{title, s, bars}, loc, InMessage)
}

// Trending は再生数が伸びている曲の一覧を整形する（見出しを除く）。
func Trending(t *library.Trending, loc *time.Location) string {
	switch {
	case !t.To.After(t.From):
		return render("trending_not_enough", t, loc, InMessage)
	case len(t.Videos) == 0:
		return render("trending_no_videos", t, loc, InMessage)
	}

	return render("trending", t, loc, InMessage)
}

// History は動画の再生数、高評価数、コメント数の推移を整形する。
func History(h *library.VideoHistory, days int) string {
	data := struct {
		*library.VideoHistory
		Days int
	}{h, days}
	if len(h.Snapshots) < 2 {
		return render("history_not_enough", data, time.UTC, InMessage)
	}

	return render("history", data, time.UTC, InMessage)
}

// Random はランダムに選んだ曲の一覧を整形する。
func Random(picked []*library.PlaylistVideo) string {
	if len(picked) == 0 {
		return render("random_empty", picked, time.UTC, InMessage)
	}

	return render("random", picked, time.UTC, InMessage)
}

// WrappedTitle は年間のまとめの見出し。
func WrappedTitle(playlistTitle string, year int) string {
	return fmt.Sprintf("🎁 %s の %d 年まとめ", playlistTitle, year)
}

// Wrapped は年間のまとめの本文（見出しを除く）を整形する。
func Wrapped(w *library.Wrapped, loc *time.Location, target Target) string {
	if w.Added == 0 {
		return render("wrapped_empty", w, loc, target)
	}

	return render("wrapped", w, loc, target)
}

// change は before から after への値の変化を整形する。例: 1,000 → 1,200（+200）
func change(before, after uint64) string {
	sign := "+"
	diff := after - before
	if after < before {
		sign = "-"
		diff = before - after
	}

	return fmt.Sprintf("%s → %s（%s%s）", Number(before), Number(after), sign, Number(diff))
}

func count(n any) (uint64, error) {
	switch n := n.(type) {
	case uint64:
		return n, nil
	case library.ViewCount:
		return uint64(n), nil
	}
	return 0, fmt.Errorf("change: unsupported type %T", n)
}
