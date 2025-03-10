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

- Go 1.23以上
- PostgreSQL

### ローカル開発環境のセットアップ

1. リポジトリをクローンします
2. 環境変数を設定します
3. アプリケーションを実行します:
   ```bash
   go run cmd/server/main.go
   ```

### プロジェクト構造

- `cmd/server/`: アプリケーションのエントリーポイント
- `internal/domain/`: ドメインモデルとエラー
- `internal/repository/`: データアクセス層
- `internal/scheduler/`: プレイリストチェックと通知ロジック
- `internal/server/`: Discordボットサーバーとコマンド処理
- `internal/service/`: ビジネスロジック

## ライセンス

このプロジェクトは[LICENSE](LICENSE)ファイルに記載されている条件の下でライセンスされています。
