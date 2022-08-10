# Discord Playlist Notifier

毎日 12:00, 18:00 に YouTube のプレイリストに追加された動画を検知して Discord のチャンネルにメッセージを送信します。  
通知するプレイリストを登録するには、Bot が参加しているサーバーでスラッシュコマンドを利用します。

## スラッシュコマンド一覧

### `/playlist-notifier list`

サーバーで通知するプレイリストを一覧表示します。

### `/playlist-notifier add `

サーバーで通知するプレイリストを追加します。  
このコマンドを実行したチャンネルで通知されます。

### `/playlist-notifier delete`

サーバーで通知するプレイリストを削除します。

### `/playlist-notifier source`

このリポジトリへのリンクが表示されます。

## デプロイ手順

1. `git clone https://github.com/tabo-syu/discord-playlist-notifier.git`
1. `cd discord-playlist-notifier && cp .env.example .env`
1. `.env` の環境変数を修正
1. `docker compose up -d`
1. サーバーへ Bot を招待
