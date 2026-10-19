---
paths:
  - "internal/presentation/discord/**"
---

# Discord のコマンド

- コマンドはグローバルではなくギルドごとに登録します（`registrar.go`）。そのためコマンドは
  ボットの稼働中しか存在せず、`GuildCreate` のたび（再接続のたび）に登録し直されます。
- ハンドラは応答（`command.Response`: 本文と添付ファイル）を組み立てて返すだけにし、
  Discord のセッションには触れません。`InteractionResponse` にくるむのは `server.go` の役目です。
  この形のおかげでハンドラを mock だけでテストできます。文字列だけの応答は
  `command.Text(...)` で包みます。
- ハンドラはアプリケーションサービスだけを呼び、文字列の引数を値オブジェクトに変換して渡します。
- サブコマンドの手本は `playlist_notifier/` です。各サブコマンドは
  `*discordgo.ApplicationCommandOption` の変数と `*PlaylistNotifier` のメソッドを 1 ファイルに
  まとめます（`add.go` など）。

## スラッシュコマンドを追加する

1. `command/` 配下に `command.Command` を実装するパッケージを作る。
2. `cmd/bot/main.go` の `commands` スライスに追加する。router と registrar はこのスライスから
   導出されるので、他に変える箇所はありません。
