# Discord Playlist Notifier

YouTubeのプレイリストを監視し、新しい動画が追加されるとDiscordチャンネルに通知を送信するDiscordボットです。

## 機能

- 5分ごとにYouTubeプレイリストを自動的にチェックして新しい動画を検出
- 新しい動画が検出されると、設定されたDiscordチャンネルに通知を送信
- Discordのスラッシュコマンドによる簡単な設定
- 複数のプレイリストとチャンネルをサポート
- Dockerによる簡単なデプロイ
- PostgreSQLによる永続的なデータストレージ

## 必要条件

- [Docker](https://www.docker.com/)と[Docker Compose](https://docs.docker.com/compose/)
- DiscordボットトークンDeveloper Portal](https://discord.com/developers/applications)から取得)
- YouTube API キー([Google Cloud Console](https://console.cloud.google.com/)から取得)

## インストール

1. リポジトリをクローンします:
   ```bash
   git clone https://github.com/tabo-syu/discord-playlist-notifier.git
   ```

2. プロジェクトディレクトリに移動します:
   ```bash
   cd discord-playlist-notifier
   ```

3. 環境設定ファイルをコピーして設定します:
   ```bash
   cp .env.example .env
   ```

4. `.env`ファイルを編集して設定を行います:
   ```
   # DB設定
   DB_HOST=db
   DB_TIMEZONE=Asia/Tokyo  # またはお好みのタイムゾーン
   DB_PORT=5432
   DB_NAME=your_db_name
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   
   # APIキー
   YOUTUBE_APIKEY=your_youtube_api_key
   DISCORD_ACCESS_TOKEN=your_discord_bot_token
   ```

5. Docker Composeでアプリケーションを起動します:
   ```bash
   docker compose up -d
   ```

6. [Discord Developer Portal](https://discord.com/developers/applications)のOAuth2 URL生成ツールを使用して、ボットをDiscordサーバーに招待します。
   - 必要な権限: `bot`と`applications.commands`
   - 必要なボット権限: `Send Messages`, `Embed Links`, `Use Slash Commands`

## 設定

### 環境変数

| 変数 | 説明 | デフォルト値 |
|----------|-------------|---------|
| DB_HOST | PostgreSQLホスト | db |
| DB_TIMEZONE | データベースタイムゾーン | Asia/Tokyo |
| DB_PORT | PostgreSQLポート | 5432 |
| DB_NAME | PostgreSQLデータベース名 | - |
| DB_USER | PostgreSQLユーザー名 | - |
| DB_PASSWORD | PostgreSQLパスワード | - |
| YOUTUBE_APIKEY | YouTube Data API v3キー | - |
| DISCORD_ACCESS_TOKEN | Discordボットトークン | - |

## 使用方法

ボットが起動してサーバーに招待されたら、以下のスラッシュコマンドを使用してプレイリスト通知を設定できます。

### スラッシュコマンド

#### `/playlist-notifier list`

サーバーで現在監視されているプレイリストの一覧を表示します。

#### `/playlist-notifier add [playlist-id]`

監視するYouTubeプレイリストを追加します。このプレイリストの通知は、このコマンドが実行されたチャンネルに送信されます。

**パラメータ:**
- `playlist-id`: YouTubeプレイリストID。YouTubeプレイリストURLの`list=`の後の部分です。

**例:**
プレイリストURL `https://www.youtube.com/playlist?list=PLexample123` の場合、プレイリストIDは `PLexample123` です。

#### `/playlist-notifier delete [playlist-id]`

監視中のYouTubeプレイリストを削除します。

**パラメータ:**
- `playlist-id`: 監視を停止するYouTubeプレイリストID。

#### `/playlist-notifier source`

ボットのGitHubリポジトリへのリンクを表示します。

## 動作の仕組み

1. ボットは5分ごとに登録されたすべてのYouTubeプレイリストをチェックします。
2. 新しい動画が検出されると、ボットは設定されたDiscordチャンネルに通知メッセージを送信します。
3. ボットは既に通知した動画を追跡して、重複通知を避けます。

## トラブルシューティング

### ボットがコマンドに応答しない

- ボットがDiscordサーバーで必要な権限を持っていることを確認してください
- ボットがオンラインで実行中であることを確認してください
- `.env`ファイルのDiscordトークンが正しいことを確認してください

### 通知が送信されない

- YouTubeプレイリストが公開されていてアクセス可能であることを確認してください
- YouTube APIキーが有効で、YouTube Data API v3が有効になっていることを確認してください
- ボットのログでエラーメッセージを確認してください

### データベース接続の問題

- PostgreSQLが実行中でアクセス可能であることを確認してください
- `.env`ファイルのデータベース認証情報が正しいことを確認してください
- データベースとユーザーがPostgreSQLに存在するか確認してください

## 開発

### 前提条件

- Go 1.27以上
- PostgreSQL

### ローカル開発環境のセットアップ

1. リポジトリをクローンします
2. 環境変数を設定します
3. アプリケーションを実行します:
   ```bash
   go run cmd/server/main.go
   ```

### プロジェクト構造

DDD のレイヤードアーキテクチャで構成しています。依存は外側から内側への一方向です。

- `cmd/server/`: アプリケーションのエントリーポイント（依存の組み立て）
- `internal/domain/`: ドメインモデル（集約ごとのパッケージ）、リポジトリのインターフェース、エラー
  - `guild/`: ボットが参加しているサーバー
  - `subscription/`: サーバーごとのプレイリストの通知設定
  - `library/`: プレイリストの中身（アイテムと動画）
  - `notification/`: どのチャンネルに何を通知するかを決めるドメインサービス
- `internal/application/`: ユースケース
- `internal/infrastructure/`: データベース（GORM）と YouTube Data API の実装
- `internal/presentation/`: Discord のコマンド処理、定期実行、通知の投稿、MCP サーバー

### MCP サーバー

ほかのボット（Claude Code など）から、通知登録されているプレイリストの中身を調べられるように、
読み取り専用の MCP サーバーを内蔵しています。`docker compose` で起動すると、Docker の
`playlist-api` ネットワーク上の `http://playlist-notifier:8080/mcp` で待ち受けます
（ホストにはポートを公開しません）。

| ツール | 内容 |
|---|---|
| `list_playlists` | 通知登録されているプレイリストの一覧と曲数 |
| `find_videos` | 曲名・追加された期間で曲を探す（新しい順・古い順・再生数順） |
| `playlist_stats` | プレイリストの統計（`/playlist-notifier stats` と同じ内容） |
| `playlist_wrapped` | 1 年間のまとめ（`/playlist-notifier wrapped` と同じ内容） |
| `random_videos` | ランダムに曲を選ぶ |

使う側は `playlist-api` ネットワークに参加し、MCP の設定に次のように書きます。

```json
{ "mcpServers": { "playlist": { "type": "http", "url": "http://playlist-notifier:8080/mcp" } } }
```

## ライセンス

このプロジェクトは[LICENSE](LICENSE)ファイルに記載されている条件の下でライセンスされています。
