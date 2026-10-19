---
paths:
  - "**/*_test.go"
  - "**/testdata/**"
  - "internal/mocks/**"
  - ".mockery.yml"
---

# テスト

## 層ごとのテスト

- **ドメイン**（`internal/domain/library`、`notification`、`subscription`）— 純粋な
  関数とメソッドのテスト。境界の値（基準時刻ちょうど、タイムゾーンをまたぐ
  月・年の区切り、節目ちょうどの再生数）とチャンネルごとの重複の排除を固定しています。
- **アプリケーション**（`internal/application`）— リポジトリ・YouTube・`Notifier` を
  mock に差し替えて、ユースケースの流れを確かめます。保存や送信は「期待した引数で
  呼ばれたこと」で確かめます（`Update(sub(..., now))` など）。ログは各パッケージの
  `TestMain` で捨てています。
- **コマンドハンドラ**（`command/playlist_notifier`）— 本物のアプリケーションサービスを
  mock で組み立て、`Handle` に Discord の入力を渡して、返した `*command.Response` を
  比べます。テストのファイルはサブコマンドのファイルと 1 対 1（`add.go` と `add_test.go`）。
  本文の整形は view のゴールデンテストに任せ、中身のないプレイリストで確かめます。
  グラフの PNG は比べられないので、グラフを添付しない場合だけを確かめます。
- **MCP のツール**（`presentation/mcpserver`）— mock で組み立てたサーバーの `Handler()` を
  `httptest` で立て、SDK のクライアントで実際にツールを呼びます。出力は JSON の文字列で
  書いて `assert.JSONEq` で比べ、エラーのときは出力が `null` で文言を `wantError` と比べます。
  日時を JST で返すことを確かめるため、テストデータの時刻は UTC で書きます。
- **通知の Embed**（`presentation/notifier`）— `mocks.MockSender` に、送るはずの宛先と
  `*discordgo.MessageEmbed` を丸ごと期待します。
- **リポジトリ**（`infrastructure/persistence`）— Testcontainers で `postgres:14.4-alpine` を
  `TestMain` が 1 つ立て、各テストは `newDB(t)` で全テーブルを
  `TRUNCATE ... RESTART IDENTITY` してから使います（ID は 1 から振られる）。GORM の
  `NowFunc` を `now` に固定し、読み出した時刻は UTC にそろえています。前提の行は
  `seed(t, db, ...)`、書き込んだ結果は `rows[T](t, db)`（論理削除も含めて ID の順）で読みます。
  `go test -short` では飛ばします。
- **YouTube クライアント**（`infrastructure/youtube`）— `httptest` で偽の YouTube Data API を
  立て、`option.WithEndpoint` でクライアントを向けます。偽の API はエンドポイント
  （`/playlists` など）ごとに、呼ばれた順に JSON を返し（`ok(...)` / `failure()`）、受けた
  リクエストを「/エンドポイント キー=値 ...」の文字列で記録します。返した値に加え、
  記録したリクエストを `wantRequests` と比べて、50 件ごとの分割・ページング・part を
  確かめます。エラーは文言（`message(err)`）を `wantErr` と比べます。
- **ルーター**（`presentation/discord`）— コマンド名からハンドラを引くことを、
  `mocks.MockCommand` で確かめます。Discord のセッションに触れるその他の部分と、
  スケジューラ（gocron への登録だけ）、グラフの PNG はテストしていません。
- **view のテキスト** — `text_test.go` が整形した結果を `view/testdata/*.golden` と比べます。
  文言を変えたら `go test ./internal/presentation/view/ -update` で書き直し、golden の差分を
  確認します。

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
- テスト関数と `t.Run` の先頭で `t.Parallel()` を呼び、mock はケースごとに作ります。
  リポジトリのテストだけは 1 つの DB を共有するので並列にしません。
- ループの中身は「呼ぶ → 空行 → `assert`」だけにします。分岐が要るなら、ケースか
  テーブルを分けます。
- 期待値は加工せず、戻り値の型そのものを書きます（ID の一覧などに変換しない）。
  状態を変えるメソッドは、呼んだ後のレシーバーを `want` に書きます。
- 比べ方は `assert.Equal(t, tt.want, got)` が基本です。順序が決まらない結果は
  `assert.ElementsMatch`、ランダムな結果は `assert.Len` / `assert.Subset`。エラーは
  `assert.ErrorIs(t, err, tt.wantErr)`（成功なら `wantErr: nil`）、続けられないエラーは `require`。
- テストデータを作る関数（`video`、`added`、`sub` など）はパッケージごとに
  `helpers_test.go` に置き、何を返すかをコメントに書きます。
- 外部のテストパッケージ（`package library_test`）から、公開された API だけを使います。
- CLAUDE.md の「既知の挙動」は、今の挙動のまま固定します。
- view のゴールデンテストだけは、ケース名から golden のファイル名を決めるため
  `map[string]string`（ケース名 → 整形した結果）のテーブルにしています。

## mock

依存先のインターフェースの mock は mockery（testify の `mock.Mock`）で `internal/mocks` に
生成します。対象は `.mockery.yml` に並べてあり、インターフェースを変えたら
`go tool task mock` で生成し直します。

```go
tests := []struct {
	name  string
	setup func(m *deps) // 呼ばれるはずのメソッドと戻り値
	want  ...
}{
	{
		name: "...",
		setup: func(m *deps) {
			m.subscriptions.EXPECT().FindByGuild(guild.DiscordID("guild")).Return(subs, nil)
			m.subscriptions.EXPECT().UpdatePickInterval(picking(subscription.PickDaily)).Return(nil)
		},
		want: ...,
	},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()

		m := newDeps(t)
		tt.setup(m)
		s := application.NewSubscriptionService(m.youtube, m.subscriptions, m.guilds)

		err := s.SetPickInterval(...)

		assert.ErrorIs(t, err, tt.wantErr)
	})
}
```

- パッケージごとに、使う mock をまとめた `deps` と `newDeps(t)` を `helpers_test.go` に置きます。
- `setup` には、そのケースで呼ばれる全てのメソッドを引数と戻り値つきで書きます。
  何も呼ばれないケースは `func(m *deps) {}` です。
- 引数は型を合わせて書きます（`"PL"` ではなく `library.PlaylistID("PL")`、nil のスライスは
  `[]*library.Video(nil)`）。mock は型も含めて比べます。
- 回数を確かめたいときは `.Times(n)` / `.Once()`、順序は `.NotBefore(...)`、順序の決まらない
  引数は `sameElements(want)` を使います。呼ばれても呼ばれなくてもよい期待（複数のケースで
  共有する mock の準備など）には `.Maybe()` を付けます。
- 何度も同じ組み合わせで期待する呼び出しは `m.expectUnchangedSync(...)`・`m.guildHas(...)`・
  `m.playlistHas(...)` のような `deps` のメソッドにまとめ、何を期待するかをコメントに書きます。
- mock が返した値をサービスが書き換えるときは、その値をケースごとに新しく作ります。
  読むだけのユースケース（`QueryService`、MCP のツール）では使い回して構いません。
- 何も返さないメソッドは `setup` の期待だけで確かめ、ループの中に `assert` を書きません。
