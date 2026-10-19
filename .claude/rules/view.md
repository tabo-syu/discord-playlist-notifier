---
paths:
  - "internal/presentation/view/**"
---

# 文言とグラフの整形（`internal/presentation/view`）

## テキスト

`summary`・`trending`・`history`・`random`・`wrapped` の本文は `view/templates/*.md.tmpl`
（`text/template`、`go:embed`）が整形します。

- 文言はテンプレートに書き、`text.go` にはテンプレートに渡すデータの用意（棒グラフの長さ
  などの計算）とテンプレート関数（`number`、`date`、`link`、`rank`、`first`、`last`、`change`）
  だけを置きます。
- テンプレートは 1 ファイルに 1 つです。中身がないときなどメッセージの形ごと変わる場合は、
  Go 側で別のテンプレート（`summary_empty` など）を選びます。テンプレートの中の分岐は、
  あるときだけ出す部分（ランキングなど）の `{{with}}` / `{{if}}` に限ります。
- `date` のタイムゾーンと `link` の書き方（`InMessage` / `InEmbed`）は整形のたびに差し替えます。
- 改行は `{{-` / `-}}` で制御しているので、テンプレートを変えたら
  `go test ./internal/presentation/view/ -update` で golden を書き直し、差分で出力を確かめます。

## グラフ

`HistoryChart` と `TrendingChart` は `github.com/wcharczuk/go-chart/v2` で描きます。

- 見た目（夜空の色、丸ゴシックの日本語フォント、光る線、凡例、目盛り）は `chart.go` に
  まとめています。フォントは `fonts/` の M PLUS Rounded 1c（OFL、`OFL.txt`）を埋め込んでいます。
- go-chart の `LegendThin` は枠の色の指定を無視し、値札は点の右にしか置けないので、凡例
  （`history`）と、`trending` の線の端の値札・グラフの下の表は自前で描いています。
