---
paths:
  - "internal/presentation/discord/**"
---

# Discord のゲートウェイとコマンド

## ゲートウェイ（`internal/presentation/discord`）

- `server.go` — discordgo のセッションを持ち、`GuildCreate`（コマンドの登録 + ギルドの
  レコード作成）、`GuildDelete`（レコード削除）、`InteractionCreate`（ルーティング）を登録します。
- `registrar.go` — コマンドはグローバルではなく**ギルドごと**に登録し、`Stop()` で削除できる
  ようにメモリに持ちます。コマンドはボットの稼働中しか存在せず、`GuildCreate` のたび
  （再接続のたび）に登録し直されます。
- `router.go` はトップレベルのコマンド名 → `command.HandleType`、`event.go` はイベントと
  `GuildService` をつなぐ薄いアダプタです。

## コマンドハンドラ（`command/`）

```go
func(request *discordgo.ApplicationCommandInteractionData, guildId, channelId string) *command.Response
```

ハンドラは応答（`command.Response`: 本文と添付ファイル）を組み立てて返すだけにし、
Discord のセッションには触れません。`InteractionResponse` にくるむのは `server.go` の役目です。
この形のおかげでハンドラを mock だけでテストできるので、崩さないでください。
文字列だけの応答は `command.Text(...)` で包みます。

- ハンドラはアプリケーションサービスだけを呼び、文字列の引数を値オブジェクトに変換して渡します。
- サブコマンドの手本は `playlist_notifier/` です。`playlist_notifier.go` がトップレベルを宣言して
  `data.Options[0].Name` でディスパッチし、各サブコマンドは
  `*discordgo.ApplicationCommandOption` の変数と `*PlaylistNotifier` のメソッドを 1 ファイルに
  まとめます（`add.go`、`list.go`、`delete.go`、`source.go`）。
- グラフを PNG で添付する例は `history.go` と `trending.go` です（描画は `view`）。

## スラッシュコマンドを追加する

1. `command/` 配下に `command.Command`（`Handle`、`GetCommand`、`SetCommand`）を実装する
   パッケージを作る。
2. `cmd/server/main.go` の `commands` スライスに追加する。router と registrar はこのスライスから
   導出されるので、他に変える箇所はありません。
