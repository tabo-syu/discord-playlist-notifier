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

## `NotifyUpdates` の順序

全ての通知設定を読む → `LibraryService.Sync` でプレイリストの中身を同期する → 非公開・削除に
なった動画を通知する → `notification.FindNewVideos` で差分を取る → `NotifiedUntil` を進めて保存する →
投稿する、の順です。

基準時刻は送信の**前**に保存し、保存に成功した通知設定だけに送ります。通知を取りこぼす
可能性と引き換えに、二重通知を防いでいます。順序を入れ替えないでください。

## 「新しい動画」の判定

保存済みのアイテムと「通知済みか」を突き合わせる処理はなく、基準は
`Subscription.NotifiedUntil` です。`Subscription.IsNew(item)` は
`!item.AddedAt.Before(s.NotifiedUntil)` で判定します。

- `After` ではなく `!Before` なのは、基準時刻ちょうどに追加された動画を含めるためです。
- 登録したときは `subscription.New` が登録した時刻を入れ、通知したときは `MarkNotified` が進めます。
  `UpdatedAt`（GORM が保存のたびに書き換える）は判定に使いません。
- `NotifiedUntil` を過去に戻せば、その時刻以降の動画を再通知できます。
- プレイリストへの追加時刻は `PlaylistItem.AddedAt`、動画の投稿時刻は `Video.PublishedAt` です。

## その他

- 各ジョブは `guard` で包み、`recover` で panic を捕まえて、1 回の失敗でボット全体が落ちない
  ようにしています。新しいジョブも `guard` で包みます。
- 通知の Embed に出す投稿者は送る直前に `LibraryService.LiveVideos` で取り、取れなかったときは
  投稿者なしで送ります。
