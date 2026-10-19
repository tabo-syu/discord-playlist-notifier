---
paths:
  - "**/*_test.go"
  - "**/testdata/**"
  - "internal/mocks/**"
  - ".mockery.yml"
---

# テスト

## 層ごとの注意点

- **アプリケーション** — 保存や送信は「期待した引数で呼ばれたこと」で確かめます
  （`Update(sub(..., now))` など）。ログは `TestMain` で捨てています。
- **コマンドハンドラ** — 本物のアプリケーションサービスを mock で組み立てて `Handle` を呼び、
  返した `*command.Response` を比べます。本文の整形は view のゴールデンテストに任せ、
  グラフの PNG は比べられないので、グラフを添付しない場合だけを確かめます。
- **MCP のツール** — 出力は JSON の文字列で書いて `assert.JSONEq` で比べます。日時を JST で
  返すことを確かめるため、テストデータの時刻は UTC で書きます。
- **リポジトリ** — 1 つの DB を `newDB(t)` が毎回 `TRUNCATE ... RESTART IDENTITY` するので、
  ID は 1 から振られ、並列にはしません。`go test -short` では飛ばします。DB の行の準備（`seed`）と
  確認（`rows`）にはレコード（`export_test.go` がテストにだけ公開する `persistence.VideoRecord` など）を、
  リポジトリが返す値にはドメインのモデルを使います。
- **YouTube クライアント** — `httptest` の偽の API が受けたリクエストを記録するので、返した値に
  加えて `wantRequests` を比べ、50 件ごとの分割・ページング・part を確かめます。
- **view のテキスト** — 文言を変えたら `go test ./internal/presentation/view/ -update` で
  golden を書き直し、差分で見た目を確かめます。

## テストの書き方

読むときに頭を使わずに済むよう、全てのテストを同じ形で書きます。

```go
func TestComputeWrapped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string                   // 何を確かめるか（日本語の文）
		contents []*library.PlaylistVideo // 入力。名前は引数と同じにする
		want     *library.Wrapped         // 期待値。複数なら wantXxx
	}{
		{
			name:     "...",
			contents: ...,
			want: &library.Wrapped{
				Year:  2026,
				Added: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.ComputeWrapped(tt.contents, 2026, jst)

			assert.Equal(t, tt.want, got)
		})
	}
}
```

- 1 つの関数・メソッドにつき 1 つのテーブル（`[]struct` + `t.Run(tt.name, ...)`）。
  ケースが 1 つでもこの形にします。
- テストケースは 1 フィールド 1 行、スライスやマップの要素は 1 要素 1 行で書きます
  （表の行として読めるように）。
- ケースは 1 つのことだけを確かめ、入力は最小限にします。`name` で何を確かめるかを書きます。
- ケース同士で状態を共有せず、mock はケースごとに作ります。
- ループの中身は「呼ぶ → 空行 → `assert`」だけにします。分岐が要るなら、ケースか
  テーブルを分けます。
- 期待値は加工せず、戻り値の型そのものを書きます（ID の一覧などに変換しない）。
  状態を変えるメソッドは、呼んだ後のレシーバーを `want` に書きます。
- 比べ方は `assert.Equal(t, tt.want, got)` が基本です。順序が決まらない結果は
  `assert.ElementsMatch`、ランダムな結果は `assert.Len` / `assert.Subset`。エラーは
  `assert.ErrorIs(t, err, tt.wantErr)`（成功なら `wantErr: nil`）、続けられないエラーは `require`。
- テストデータを作る関数（`video`、`added`、`sub` など）はパッケージごとに
  `helpers_test.go` に置き、何を返すかをコメントに書きます。
- CLAUDE.md の「既知の挙動」は、今の挙動のまま固定します。
- view のゴールデンテストだけは、ケース名から golden のファイル名を決めるため
  `map[string]string`（ケース名 → 整形した結果）のテーブルにしています。

## mock

依存先の mock は mockery で `internal/mocks` に生成します（対象は `.mockery.yml`、生成は
`go tool task mock`）。パッケージごとに、使う mock をまとめた `deps` と `newDeps(t)` を
`helpers_test.go` に置き、テーブルの `setup func(m *deps)` で期待を書きます。

```go
setup: func(m *deps) {
	m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return(subs, nil)
	m.subscriptions.EXPECT().Update(picking(subscription.PickDaily)).Return(nil)
},
```

ループの中は「`m := newDeps(t)` → `tt.setup(m)` → サービスを組み立てる → 空行 → 呼ぶ → 空行 →
`assert`」です。

- `setup` には、そのケースで呼ばれる全てのメソッドを引数と戻り値つきで書きます。
  何も呼ばれないケースは `func(m *deps) {}` です。
- 引数は型を合わせて書きます（`"PL"` ではなく `library.PlaylistID("PL")`、nil のスライスは
  `[]*library.Video(nil)`）。mock は型も含めて比べます。
- 回数は `.Times(n)` / `.Once()`、順序は `.NotBefore(...)`、順序の決まらない引数は
  `sameElements(want)`、呼ばれなくてもよい期待（ケース間で共有する準備）は `.Maybe()` です。
- 何度も同じ組み合わせで期待する呼び出しは `m.expectUnchangedSync(...)` のような `deps` の
  メソッドにまとめ、何を期待するかをコメントに書きます。
- mock が返した値をサービスが書き換えるときは、その値をケースごとに新しく作ります。
- 何も返さないメソッドは `setup` の期待だけで確かめ、ループの中に `assert` を書きません。
