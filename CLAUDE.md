# CLAUDE.md

このリポジトリで作業する AI アシスタント向けのガイドです。人向けの説明（機能、データの
ライフサイクル、トラブルシューティング）は `README.md` にあります。

<!--
メンテナー向け: このファイルは毎回のセッションで全文が読み込まれるので、200 行未満に保つ。
一部のディレクトリにだけ関係する内容は .claude/rules/ に paths 付きで書く。
「この行を消したら Claude が間違えるか？」で残すかを決める。
-->

## このプロジェクトについて

YouTube のプレイリストを 5 分ごとにポーリングし、新しい動画が追加されたら
Discord のテキストチャンネルに Embed を投稿する Discord ボット（Go 製）です。
状態は GORM を通して PostgreSQL に永続化します。設定はサーバー（ギルド）ごとに
`/playlist-notifier` スラッシュコマンドで行います。

モジュールパス: `github.com/tabo-syu/discord-playlist-notifier`

## コマンド

変更を終えたら `go tool task check`（build・go mod tidy のずれ・lint・test）を通してから
完了と報告します。Docker が使えないときは
`go build ./... && go tool task lint && go tool task test:short` を実行し、DB のテストを
飛ばしたことを報告に書きます。

```bash
go tool task fmt          # gofumpt + goimports で整形する
go tool task lint         # golangci-lint（.golangci.yml）
go tool task lint:fix     # 空行など、自動で直せる指摘を直す（fmt では直らない）
go tool task test         # Testcontainers で PostgreSQL を立てるので Docker が必要
go tool task test:short   # Docker なしで、DB のテストだけ飛ばす
go tool task check        # コミット前はこれ
go tool task mock         # インターフェースを変えたら mock を生成し直す
go tool task hooks        # lefthook の Git フックを入れる（clone したら 1 回だけ）

# スタック全体（マイグレーション + ボット + MCP サーバー + PostgreSQL）を起動する。通常はこちら
cp .env.example .env && docker compose up -d && docker compose logs -f bot mcp
```

- Task・mockery・lefthook は `go.mod` の `tool`、golangci-lint は `Taskfile.yml` の
  `go run ...@v2.15.0` で版を固定しているので、インストールは要りません。CI はありません。
- lefthook はコミット前に整形・tidy・lint、プッシュ前に `go tool task test` を実行します。
  Claude Code の hook（`.claude/settings.json`）が、セッションの開始時に lefthook を入れ、
  フックを飛ばすコマンド（`--no-verify`、`LEFTHOOK=0` など）を止めます。

## アーキテクチャ

DDD のレイヤードアーキテクチャです。依存は外側から内側への一方向で、depguard が
強制します（`domain` は他のレイヤと GORM を、`application` は `infrastructure` と `presentation` を
import しない。`internal/mocks` はテストからだけ import する）。

```
cmd/bot/main.go              ボットのコンポジションルート — 全ての組み立てと依存注入
cmd/mcp/main.go              MCP サーバーのコンポジションルート（読み取りに要るものだけ）
cmd/migrate/main.go          スキーマを更新して終了する（persistence.Migrate）
internal/presentation/       外界との入出力（discord・scheduler・notifier・mcpserver・view）。ユーザーに見える日本語はここだけ
internal/infrastructure/     ドメインが宣言したインターフェースの実装（persistence・youtube）
internal/application/        ユースケース（リポジトリから集約を読み、ドメインに判断させ、保存する）
internal/domain/             集約ごとのパッケージ（guild・subscription・library・notification）
internal/env/                プロセスの環境変数
internal/mocks/              mockery が生成した mock（手で編集しない）
```

アプリを駆動する経路は 3 つです。

1. **対話的な経路** — Discord のイベント → `presentation/discord` → `GuildService`・
   `SubscriptionService`・`LibraryService` → リポジトリ
2. **定期実行の経路** — `presentation/scheduler` → `NotificationService` → リポジトリと
   `LibraryService` → `application.Notifier`（実装は `presentation/notifier`）が Discord に投稿
3. **MCP の経路** — 外部のクライアント → `presentation/mcpserver` → `QueryService`（読み取りのみ）

### ドメインモデル

- 集約は互いに ID でのみ参照します（`Subscription.GuildID`、`Subscription.YoutubeID`）。
  GORM の関連付けは使いません。
- ID や状態は値オブジェクト（`guild.DiscordID`、`library.VideoID` など）です。文字列との
  変換は境界（コマンドハンドラ、`infrastructure`）で行います。
- ドメインのモデルは保存の仕方を知りません。テーブルの行の形（`gorm.Model`、タグ、
  `TableName()`）は `infrastructure/persistence/records.go` のレコードが持ち、リポジトリが
  ドメインのモデルとの間で変換します。代理キーの `ID` をドメインに残しているのは、集約の
  識別子として使う `Guild` と `Subscription` だけで、`library` のモデルは YouTube の ID で識別します。
  監査用の時刻（`created_at`・`updated_at`）と論理削除（`deleted_at`）はレコードにだけあります。
- フィールドは公開していますが、状態は直接代入せず、メソッド
  （`Subscription.MarkNotified`、`Video.ChangePrivacy` など）を通して変えます。
- リポジトリと外部サービスのインターフェース（`library.YouTube`）は各集約の
  パッケージで宣言します。

### 依存の組み立て

依存を組み立てるのは `cmd/bot/main.go` と `cmd/mcp/main.go` だけです。DI コンテナはありません。
新しい依存もここに追加します。

- スキーマを更新するのは `cmd/migrate` だけで、ボットと MCP サーバーは更新しません。
  compose では一度だけ動く `migrate` が `db` の healthcheck を待って実行し、`bot` と `mcp` は
  `service_completed_successfully` でその完了を待ちます。`bot` だけが再起動しても
  マイグレーションは走りません。

- 現在時刻は `time.Now` をコンストラクタで渡します。テストで時刻を固定するためで、
  サービスの中では `time.Now()` を直接呼びません。
- 日付の表示や月・年の区切りには、スケジューラの `*time.Location`（`DB_TIMEZONE`）を
  コンストラクタで渡して使います。

## 規約

### Go のスタイル

- インターフェースを受け取り、構造体を返します。インターフェースは使う側で宣言し、
  コンストラクタは非公開の具象型（`*subscriptionRepository` など）を返します。
  外から型を参照したいときも、構造体は非公開のままインターフェースで受けます。
- コンストラクタは構築済みの依存を受け取る `NewX(...)` で、複合リテラルに位置指定で
  代入します（`return &Server{s, rg, e, rt}`）。
- エラーは、何が起きたかをユビキタス言語で表すセンチネル値を返し、コマンド層がユーザー向けの
  文言に変換します。業務の失敗は関係する集約の `errors.go`（`subscription.ErrNotSubscribed`
  など）、外部サービスに繋がらないなどユースケースの失敗は `application`
  （`application.ErrYouTubeUnavailable`）に置きます。リポジトリは「ない」「もうある」を
  集約のエラーに訳します。それ以外の技術的な失敗（接続できないなど）は、そのまま
  か `fmt.Errorf("...: %w", err)` で包んで上へ返します。
- 内側のスコープで `err` を宣言し直さず、外側に `err` があれば `if err = f(); ...` と
  代入します（govet の shadow）。
- 空行は nlreturn と wsl_v5 が決めます。指摘は `go tool task lint:fix` で直します。
- ログは標準ライブラリの `log` で、フィールドをスペース区切りで並べます
  （`log.Println("Guild record created:", guildId)`）。

### 言語

- ユーザーの目に触れる文言（コマンドの説明・応答、Embed のラベル）は日本語で、既存の
  トーン（`"エラー！システムに問題があります！"`）に合わせます。
- コードコメントは日本語（識別子や固有名詞はそのまま）、ログは英語です。
- コミットの件名は `[種別] 概要` です。種別は `feat`、`fix`、`chore`、`doc`、`refactor`、
  `add`、`update` から選びます。

### 永続化するフィールドを追加する

`internal/domain/` の構造体と、`infrastructure/persistence/records.go` のレコード、その間の
変換の関数（`toXRecord`・`toDomain`）の 3 か所に追加すれば、`cmd/migrate` が
`persistence.Migrate` の `AutoMigrate` でカラムを足します。レコードのフィールドは値オブジェクト
ではなく基本の型（`string`・`uint64` など）にし、インデックスなどのタグもレコードに付けます。
写し忘れは、全てのフィールドに値を入れて保存・読み戻すリポジトリのテスト（`fullVideo` など）で
見つけます。マイグレーションファイルはありません。新しいテーブルはレコードを `Migrate` に
追加し、レコードに `TableName()` を定義します。既存の行に値を入れる必要があるときは、AutoMigrate の後に
何度実行しても同じ結果になる UPDATE を `Migrate` に書きます（`notified_until` の例）。リネームと削除は AutoMigrate がしないので、DB に
手で行う必要があることをユーザーに伝えます。

### テスト

テストを書く・直すときは、先に同じパッケージの `helpers_test.go` を開きます
（`.claude/rules/testing.md` の書き方が読み込まれます）。カバレッジ 100% は目指さず、
壊れると通知の二重送信・取りこぼしや集計の誤りにつながるロジックを優先します。

## 設定

`internal/env/env.go` は全ての変数をパッケージ初期化時に読み込みます。アプリは `.env` を
読まない（読むのは `docker compose` の `env_file:`）ので、Docker の外で `go run` するときは
環境変数を自分でエクスポートします。未設定の変数は空文字列になります。

| 変数 | 用途 |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | PostgreSQL の DSN（`sslmode=disable`） |
| `DB_TIMEZONE` | IANA のタイムゾーン名。DSN と gocron のロケーションの両方に使う |
| `DISCORD_ACCESS_TOKEN` | `env.DISCORD_TOKEN` に読み込まれる（名前が違う） |
| `YOUTUBE_APIKEY` | `env.YOUTUBE_TOKEN` に読み込まれる（名前が違う） |
| `MCP_ADDR` | MCP サーバー（`cmd/mcp`）の待ち受けアドレス（例 `:8080`）。`docker-compose.yml` の `mcp` で設定 |

実際のトークンはコード・テスト・コミットメッセージ・PR の説明に書かず、`.env`
（gitignore 済み）だけに置きます。

## YouTube API

クォータは 1 日 10,000 ユニット（list 系は 1 回 1 ユニット）です。YouTube から取る前に、
DB に保存した中身（`library.Repository`）で済ませられないかを先に検討します。

## 運用

友達のサーバー向けに少人数で動かしているボットです。数日止まっても困る人はいないので、
デプロイやマイグレーション、コンテナの再起動を慎重にやりすぎる必要はありません。
ゼロダウンタイム、ロールバック手順、段階的なリリースなどは考えず、壊れたら直して
もう一度デプロイすれば十分です。ただし、DB のデータ（通知設定や動画の履歴）は
作り直せないので、消す操作だけは事前に確認します。

## 検証できないこと

Discord や YouTube の実際の挙動は、本物の認証情報とテスト用サーバーがないと確かめられ
ません。確かめていない場合は、報告にその旨を書きます。

## 既知の挙動

いずれも既存の挙動で、テストもこの挙動のまま固定しています。頼まれない限り直さず、
関係するときは把握したうえで作業します。

- `subscription.Repository` の `FindAll` と `FindByGuild` は、結果が空のときに空の
  スライスではなく `subscription.ErrNotSubscribed` を返します。そのため、どこにも
  プレイリストが登録されていない状態では、スケジューラが 5 分ごとに
  `Could not notify cause: playlist is not subscribed` を出力し続けます。
- 上記の影響で、プレイリストを 1 つも持たないギルドでは
  `GuildService.Unregister` が途中で return し、`GuildDelete` の際にギルドの
  レコードが削除されません。
- `add.go` と `delete.go` はエラーを `switch err`（完全一致、`//nolint:errorlint`）で
  比較し、`list.go` は `errors.Is` を使っています。前者ではラップされたエラーが
  `default` に流れます。
- `subscription.Repository.DeleteAll` は通知設定を論理削除します。`playlist_items` と
  `videos` は他のサーバーと共有しているので削除しません。
- `playlists.guild_id` の外部キー `fk_guilds_playlists` は、集約を ID で参照するように
  したため `persistence.Migrate` が明示的に作成します。
- Dockerfile は `go run` を実行し（バイナリのビルドはしない。`migrate` と `mcp` は `command:` で
  `cmd/migrate`・`cmd/mcp` に差し替える）、`docker-compose.yml` は
  ソースをコンテナにバインドマウントします。
