---
paths:
  - "internal/presentation/mcpserver/**"
  - "internal/application/query_service*.go"
  - "docker-compose.yml"
---

# MCP サーバー（`presentation/mcpserver` と `QueryService`）

`MCP_ADDR` が設定されているときだけ、Streamable HTTP の MCP サーバーを `/mcp` で起動します
（`github.com/modelcontextprotocol/go-sdk`、ステートレス）。起動に失敗してもログを出すだけで、
ボット本体は動き続けます。

- ツールは全て読み取り専用です（`list_playlists`、`find_videos`、`playlist_stats`、
  `playlist_wrapped`、`random_videos`、`video_history`）。ツールや引数の説明は日本語で書きます。
- `QueryService` は DB に保存した中身だけを読み、YouTube API は呼びません（クォータを使わない）。
- 見せるのは、どこかのギルドが通知登録しているプレイリストだけです。Discord のギルド ID や
  チャンネル ID は返しません。
- 出力は SDK が出力スキーマで検証します。nil のスライスは `null` になってスキーマ違反に
  なるので、空のときも空のスライスを返します。json タグは snake_case です（tagliatelle）。

## 公開範囲

認証がないので、Docker ネットワークで公開範囲を絞っています。`docker-compose.yml` でボットを
`playlist-api` ネットワークにも接続し、`playlist-notifier` という別名で公開します
（`http://playlist-notifier:8080/mcp`）。DB は既定のネットワークにしかいないので、
`playlist-api` に参加した他のプロジェクトからは届きません。ホストにはポートを公開していません。
公開範囲を広げる変更をするときは、認証がないことをユーザーに確認します。
