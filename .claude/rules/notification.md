---
paths:
  - "internal/presentation/scheduler/**"
  - "internal/presentation/notifier/**"
  - "internal/application/notification_service*.go"
  - "internal/application/notifier.go"
  - "internal/domain/notification/**"
  - "internal/domain/subscription/**"
---

# 定期実行と通知

## スケジュール（`presentation/scheduler`）

`scheduler.go` はジョブの間隔の設定だけを持ちます。各ジョブは `guard` で包み、`recover` で
panic を捕まえて、1 回の失敗でボット全体が落ちないようにしています。

| 間隔 | 呼ぶもの |
|---|---|
| 5 分ごと | `NotificationService.NotifyUpdates` |
| 6 時間ごと | `NotificationService.RefreshVideos`（再生数などの更新と節目の通知） |
| 毎日 12:00 | `NotificationService.PostPicks` |
| 12/31 21:00 | `NotificationService.PostWrapped` |
| 起動時と 1 日ごと | `LibraryService.RefreshCategories`（直接呼ぶ） |

## `NotifyUpdates` の順序

全ての通知設定を読む → `LibraryService.Sync` でプレイリストの中身を同期する → 非公開・削除に
なった動画を通知する → `notification.FindNewVideos` で差分を取る → `UpdatedAt` を更新する →
投稿する、の順です。

基準時刻は送信の**前**に保存し、保存に成功した通知設定だけに送ります。通知を取りこぼす
可能性と引き換えに、二重通知を防いでいます。順序を入れ替えないでください。

## 「新しい動画」の判定

保存済みのアイテムと「通知済みか」を突き合わせる処理はなく、基準は
`Subscription.UpdatedAt` です。`FindNewVideos` は、同期結果のアイテムのうち
`Subscription.IsNew(item)`（`!item.AddedAt.Before(s.UpdatedAt)`）が真で、動画が
`Available()` なものを、追加時刻の古い順に返します。

- `After` ではなく `!Before` なのは、基準時刻ちょうどに追加された動画を含めるためです。
- `UpdatedAt` を過去に戻せば、その時刻以降の動画を再通知できます。
- プレイリストへの追加時刻は `PlaylistItem.AddedAt`、動画の投稿時刻は `Video.PublishedAt` です。

## Embed（`presentation/notifier`）

- 動画 1 件につき 1 通の `ChannelMessageSendEmbed` を送ります。`notifier.New` は
  `*discordgo.Session` ではなく、それだけを持つ `notifier.Sender` を受け取ります。
- 動画の通知は、通知ごとに違う部分（色・見出し・フィールド・画像の大きさ）だけを
  `videoNotice` に並べ、タイトル・URL・サムネイル・投稿者は `video.go` の `sendVideo` が
  組み立てます。1 ファイルに通知 1 種類です。
- 投稿者（チャンネル名とアイコン）は送る直前に `LibraryService.LiveVideos` で取り、
  取れなかったときは投稿者なしで送ります。
