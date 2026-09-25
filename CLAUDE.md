# CLAUDE.md

このリポジトリで作業する AI アシスタント向けのガイドです。

## このプロジェクトについて

YouTube のプレイリストを 5 分ごとにポーリングし、新しい動画が追加されたら
Discord のテキストチャンネルに Embed を投稿する Discord ボット（Go 製）です。
状態は GORM を通して PostgreSQL に永続化します。設定はサーバー（ギルド）ごとに
`/playlist-notifier` スラッシュコマンドで行います。

モジュールパス: `github.com/tabo-syu/discord-playlist-notifier`

## コマンド

```bash
# ビルド / 静的チェック（テストスイートは存在しない。「テスト」の節を参照）
go build ./...
go vet ./...
gofmt -l .          # 何も出力されないこと。「既知の挙動」の節も参照

# スタック全体（ボット + PostgreSQL）を起動する。通常はこちらを使う
cp .env.example .env   # コピー後、実際の値を記入する
docker compose up -d
docker compose logs -f bot

# ボットを直接実行する。到達可能な PostgreSQL と、シェルにエクスポートされた
# 全ての環境変数が必要（アプリは .env を読み込まない。「設定」の節を参照）
go run cmd/server/main.go
```

Makefile、CI ワークフロー、golangci-lint の設定ファイルはいずれもリポジトリに
ありません。devcontainer が VS Code の設定として `go.lintTool: golangci-lint` を
指定しているだけなので、lint はエディタ側の機能であり、どこでも強制されていません。

## アーキテクチャ

一方向の厳密なレイヤ構成です。下位のレイヤが上位のレイヤを import することは
ありません。

```
cmd/server/main.go            コンポジションルート — 全ての組み立てと依存注入
        │
        ├── internal/server/          Discord ゲートウェイ: セッション、イベント、コマンド
        │       └── command/          コマンド定義とハンドラ
        ├── internal/scheduler/       5 分間隔のポーリングと Embed の描画
        │
        ├── internal/service/         ビジネスロジック（リポジトリを組み合わせる）
        ├── internal/repository/      データアクセス: PostgreSQL (GORM) と YouTube Data API v3
        ├── internal/domain/          GORM のモデルとセンチネルエラー
        └── internal/env/             プロセスの環境変数
```

アプリを駆動する経路は独立して 2 つあります。

1. **対話的な経路** — Discord ゲートウェイのイベント → `internal/server` → `service` → `repository`
2. **定期実行の経路** — gocron が 5 分ごとに `scheduler.schedule.Notify` を呼ぶ → `service` → `repository` → `scheduler.renderer` が Discord に投稿

### 依存の組み立て

依存関係を構築している場所は `cmd/server/main.go` だけです。`init()` が 4 つの
パッケージレベルのシングルトン（`sr` gocron、`db` GORM、`dc` discordgo、
`yt` YouTube サービス）を構築し、失敗時は全て `log.Fatalf` で落とします。
その後 `main()` がリポジトリ → サービス → server/scheduler の順に組み立て、
`SIGINT` を待ってブロックします。新しい依存を追加する場合はここに書きます。
DI コンテナや設定フレームワークの類はありません。

### `internal/server`

- `server.go` — discordgo のセッションを保持し、3 つのハンドラを登録します。
  `GuildCreate`（スラッシュコマンドの登録 + ギルドレコードの作成）、
  `GuildDelete`（ギルドレコードの削除）、`InteractionCreate`（コマンドハンドラへのルーティング）。
- `registrar.go` — コマンドを**ギルドごとに**登録し（グローバル登録ではなく
  ギルド ID 付きの `ApplicationCommandCreate`）、`Stop()` で削除できるように
  メモリ上に保持します。結果として、コマンドはボットの稼働中しか存在せず、
  `GuildCreate` のたび（＝再接続のたび）に登録し直されます。
- `router.go` — トップレベルのコマンド名から `command.HandleType` へのマップ。
- `event.go` — ゲートウェイイベントと `GuildService` をつなぐ薄いアダプタ。

### `internal/server/command`

`Command` は全てのコマンドが実装するインターフェースです（`Handle`、
`GetCommand`、`SetCommand`）。ハンドラのシグネチャは次のとおりです。

```go
func(request *discordgo.ApplicationCommandInteractionData, guildId, channelId string) string
```

**ハンドラは文字列を返すだけで、Discord のセッションには一切触れません。**
返された文字列を `InteractionResponse` にくるむのは `server.go` の役目です。
この形を保ってください。ハンドラが容易にテストできるのはこの設計のおかげです。

サブコマンドの構成は `playlist_notifier/` が手本になります。
`playlist_notifier.go` がトップレベルのコマンドを宣言し、`data.Options[0].Name`
でディスパッチします。各サブコマンドは自身のファイルを持ち、
`*discordgo.ApplicationCommandOption` の変数と `*PlaylistNotifier` のメソッドを
同じファイルにまとめます（`add.go`、`list.go`、`delete.go`、`source.go`）。

### `internal/scheduler`

- `scheduler.go` — 5 分間隔の設定（`Every(5).Minutes()`）。スケジューラの
  `*time.Location` をジョブに渡します。
- `schedule.go` — 各ジョブは `recover` で panic を捕まえ、1 回の失敗でボット
  全体が落ちないようにしています。
  `Notify`: 全プレイリストを読み込む → `LibraryService.Sync` でプレイリストの中身を DB に同期する → 差分を取る →
  `UpdatedAt` を更新する → 描画する、という流れです。順序に意味があります。
  タイムスタンプの永続化はメッセージ送信の**前**に行われ、更新に成功した
  プレイリストだけが通知対象になります。通知を取りこぼす可能性と引き換えに、
  二重通知が起きないようにしています。
- `renderer.go` — `discordgo.MessageEmbed`（ラベルは日本語）を組み立て、
  動画 1 件につき 1 通の `ChannelMessageSendEmbed` を送ります。

### プレイリストの中身の保存（`LibraryService`）

YouTube プレイリストの中身は、サーバーごとの通知設定（`Playlist`）とは別に、
YouTube のプレイリスト ID 単位で 2 つのテーブルに保存します。

- `playlist_items`（`domain.PlaylistItem`）— プレイリストの各アイテム。
  `AddedAt` はプレイリストに追加された時刻です。
- `videos`（`domain.Video`）— 動画ごとの最新の情報（タイトル、再生数、
  公開状態、投稿日時）。非公開や削除になっても、タイトルなどの詳細は消さずに
  残します。`Available()` が真の動画だけをユーザーに見せます。サムネイルは
  動画 ID から作れるので保存しません（`Thumbnail()`）。

投稿者（チャンネル名とアイコン）は保存しません。通知を送る直前に
`LibraryService.LiveVideos` で取得し（最新の再生数も一緒に取れます）、
取得に失敗したときは投稿者なしで送ります。

`LibraryService.Sync` はまず `playlists.list` でアイテム数を確認し、アイテム数が
変わったとき、または前回の全件取得から `FULL_SYNC_INTERVAL`（1 時間）が
経ったときだけアイテムを全件取得します。動画の詳細（`videos.list`）は、まだ
詳細を持っていない動画の分だけ取得します。動画の公開状態はアイテムの
`status.privacyStatus` から取ります（`videos.list` は非公開・削除済みの動画を
返さないため）。全件取得の状態はメモリ上にあるので、再起動直後は全件取得します。

実行が重なっても同じ `videos` を書き換えないよう、`LibraryService` の
mutex で直列化しています。

### 「新しい動画」の判定方法

保存済みのアイテムと、通知済みかどうかを突き合わせる処理はありません。
基準となるのは `Playlist.UpdatedAt` です。`PlaylistService.GetDiffFromLatest` は
同期結果のアイテムのうち `!item.AddedAt.Before(last.UpdatedAt)` が真で、動画が
`Available()` なものを新着とみなし、追加時刻の古い順に並べます。
`After` ではなく `!Before` なのは意図的で、基準時刻とちょうど同時刻に追加された
動画を含めるためです。`UpdatedAt` を過去に戻せば、その時刻以降の動画を再通知できます。
プレイリストに追加された時刻は `PlaylistItem.AddedAt`、動画自体の投稿時刻は
`Video.PublishedAt` です。

`videos` テーブルは以前、通知した動画の履歴でした。古い形のテーブル
（`playlist_id` カラムがある）は、起動時に `main.go` が一度だけ削除します。

## 規約

### Go のスタイル

- **インターフェースを受け取り、構造体を返す**（コミット `0a14cd7`）。
  リポジトリのインターフェースは `internal/repository` で宣言し、コンストラクタは
  非公開の具象型を返します（`*playlistRepository`、`*registrar`、`*router`、
  `*scheduler`、`*schedule`、`*renderer`、`*event`）。これに倣ってください。
  型名を参照したいという理由だけで構造体を公開しないこと。
- コンストラクタは構築済みの依存を受け取る `NewX(...)` の形で、複合リテラルに
  位置指定で代入します（`return &Server{s, rg, e, rt}`）。
- エラーは `internal/domain/errors.go` に置いた**センチネル値**で、発生源
  （Discord / YouTube / DB）ごとにグループ分けされています。リポジトリと
  サービスがこれを返し、コマンド層がユーザー向けの文言に変換します。
- ログは標準ライブラリの `log` を使い、フィールドをスペース区切りで並べます
  （`log.Println("Guild record created:", guildId)`）。構造化ロガーは使いません。

### 言語

- **ユーザーの目に触れる文言は日本語**です。スラッシュコマンドの説明、コマンドの
  応答文字列、Embed のフィールドラベルが該当します。既存のトーンに合わせて
  ください（例: `"エラー！システムに問題があります！"`）。
- **ログとコードコメントは英語**です（`playlist_repository.go` に古い日本語の
  コメントがいくつか残っています）。
- `README.md` は日本語です。コミットの件名は `[種別] 概要` の形式で、概要は
  日本語と英語が混在しています。履歴に出てくる種別は `feat`、`fix`、`chore`、
  `doc`、`refactor`、`add`、`update` です。

### スラッシュコマンドを追加する

1. `internal/server/command/` 配下に `command.Command` を実装するパッケージを作る。
2. `cmd/server/main.go` の `commands` スライスに追加する。router と registrar は
   どちらもこのスライスから導出されるので、他に変更すべき箇所はありません。

### 永続化するフィールドを追加する

`internal/domain/domain.go` の構造体に追加します。スキーマの変更は `main.go` の
`init()` にある `db.AutoMigrate(...)` が適用します。新しいモデルを追加したら
ここにも追加してください。**マイグレーションファイルはありません。** AutoMigrate はカラムの
追加しか行わないため、リネームや削除はデータベースに対して手動で行う必要があります。

## 設定

`internal/env/env.go` は全ての変数を**パッケージ初期化時**にパッケージレベルの
`var` へ読み込みます。Go 側に dotenv の読み込み処理はありません。`.env` を
読むのは `docker compose`（`env_file:`）であって、バイナリではありません。
Docker の外で実行する場合は、自分で環境変数をエクスポートする必要があります。
変数が未設定の場合はエラーにならず、空文字列になります。

| 変数 | 用途 |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | PostgreSQL の DSN の構成要素（`sslmode=disable`） |
| `DB_TIMEZONE` | IANA のタイムゾーン名。DSN と gocron のロケーションの両方に使われる |
| `DISCORD_ACCESS_TOKEN` | `env.DISCORD_TOKEN` に読み込まれる |
| `YOUTUBE_APIKEY` | `env.YOUTUBE_TOKEN` に読み込まれる |

最後の 2 つは環境変数名と Go の識別子名が一致していない点に注意してください。

`.env`（gitignore 済み）をコミットしないこと。また、実際のトークンをコード、
テスト、コミットメッセージ、PR の説明に貼り付けないこと。

## YouTube API の使い方

`youtube_repository.go` は全ての呼び出しを `MAX_BATCH_SIZE = 50` でバッチ化し
（プレイリスト、動画、チャンネル）、プレイリストアイテムは
`MAX_RESULTS_PER_PAGE = 50` でページングします。API のレスポンスは一部の
フィールドが欠けることがあるので、nil を考慮して読みます。タイムスタンプの
パースには `YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"` を使います。

主な制約はクォータ（標準で 1 日 10,000 ユニット、list 系は 1 回 1 ユニット）です。
YouTube から直接データを取る前に、DB に保存した中身（`LibraryRepository`）で
済ませられないか検討してください。

プレイリストアイテムの `snippet.channelId` は、ドキュメント上は「追加した
ユーザー」ですが、実際には常にプレイリストの所有者を返します。追加者は
取得できません。

## テスト

現時点でリポジトリに **`_test.go` ファイルは 1 つもありません**。テストコードは書かない
方針です。

変更の検証は最低でも `go build ./... && go vet ./...` で行ってください。
Discord や YouTube の挙動に関わる変更は実際の認証情報とテスト用サーバーが
必要です。検証できていない場合は、検証したと述べずにその旨を明示してください。

## 既知の挙動

いずれも既存の挙動です。ついでに「修正」しないでください。ただし把握しておく
必要があります。

- `gofmt -l .` が `internal/scheduler/schedule.go` を報告します（行末の空白）。
  コミットされた時点からこの状態です。
- `PlaylistRepository.FindAll` と `FindByDiscordId` は、結果が空のときに空の
  スライスではなく `domain.ErrDBRecordNotFound` を返します。そのため、どこにも
  プレイリストが登録されていない状態では、スケジューラが 5 分ごとに
  `Could not notify cause: record not found` を出力し続けます。
- 上記の影響で、プレイリストを 1 つも持たないギルドでは
  `GuildService.Unregister` が途中で return するため、`GuildDelete` の際に
  ギルドのレコードが削除されません。
- `add.go` と `delete.go` はエラーを `switch err { case ... }`（完全一致）で
  比較していますが、`list.go` は `errors.Is` を使っています。前者では
  ラップされたエラーが `default` に流れます。
- `PlaylistRepository.DeleteAll` はプレイリストを論理削除（`gorm.Model`）します。
  `playlist_items` と `videos` は他のサーバーと共有しているので削除しません。
- Dockerfile は `go run` を実行します（バイナリのビルドもマルチステージ
  ビルドもしません）。また `docker-compose.yml` はソースをコンテナに
  バインドマウントします。
