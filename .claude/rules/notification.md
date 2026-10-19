---
paths:
  - "internal/presentation/scheduler/**"
  - "internal/presentation/notifier/**"
  - "internal/application/notification_service*.go"
  - "internal/application/library_service*.go"
  - "internal/application/notifier.go"
  - "internal/application/notification_handlers*.go"
  - "internal/application/event_publisher*.go"
  - "internal/domain/notification/**"
  - "internal/domain/subscription/**"
---

# 定期実行と通知

## 5 分ごとの同期と新着の通知の順序

スケジューラが `LibraryService.Sync` を呼ぶ → 全ての通知設定を読み、プレイリストの中身を同期して
保存する → ロックを外してから、動画を保存できたプレイリストの分の `VideoHidden` を発行する（ハンドラが
非公開・削除を通知する）→ 同期の結果 `PlaylistSynced` を発行する → `NotificationService.TakeNewVideos`
（購読したハンドラ）が通知設定ごとに `Subscription.TakeNewVideos` で新着を選んで `NotifiedUntil` を
進め、保存する → 保存できた通知設定の `VideosAdded` だけをまとめて発行する（ハンドラが投稿する）、の順です。
`NotificationService` は `LibraryService` を呼ばず、同期の結果はイベントで受け取ります（issue #74）。

基準時刻は送信の**前**に保存し、保存に成功した通知設定だけに送ります。通知を取りこぼす
可能性と引き換えに、二重通知を防いでいます。これは「イベントは、それを起こした集約を保存できた
後に配る」の当てはめです。順序を入れ替えないでください。

## ドメインイベントとハンドラ

- イベントは集約のメソッドが戻り値で返します（`Video.ChangePrivacy` → `VideoHidden`、
  `Video.UpdateMilestone` → `MilestoneReached`、`Subscription.TakeNewVideos` → `VideosAdded`）。
  何も起きなければ nil です。同期の結果 `PlaylistSynced` だけは、複数の集約から組み立てるので
  `LibraryService` が作ります。
- 配るのは、その集約を保存したアプリケーションサービスです（`application.Publish`）。
  `VideoHidden`・`PlaylistSynced`・`MilestoneReached` は `LibraryService`（`Sync`・`RefreshVideos`）、
  `VideosAdded` は `NotificationService.TakeNewVideos` が発行します。`EventPublisher` は同期的に、
  登録した順にハンドラを呼びます。
- ハンドラ（`NotificationHandlers`）は 1 回の発行のイベントをまとめて受け取ります。チャンネルごとの
  重複の排除と、節目の `FetchLiveVideos` の呼び出しを 1 回にまとめるためです（新着は今までどおり
  通知設定ごとに 1 回呼びます）。
- 非公開と節目のハンドラは、全ての通知設定と「動画を含むプレイリスト」（`library.Catalog.PlaylistsContaining`、
  DB に保存したアイテム）を自分で読み、`notification.FindHiddenVideos` / `FindMilestones` に
  チャンネルごとに 1 回にまとめさせます。そのため、非公開の通知に出すプレイリスト名は
  `Subscription.Title`（保存したもの）で、同期に失敗したプレイリストのチャンネルにも届きます。

## 「新しい動画」の判定

保存済みのアイテムと「通知済みか」を突き合わせる処理はなく、基準は
`Subscription.NotifiedUntil` です。`Subscription.IsNew(item)` は
`!item.AddedAt.Before(s.NotifiedUntil)` で判定します。

- `After` ではなく `!Before` なのは、基準時刻ちょうどに追加された動画を含めるためです。
- 登録したときは `subscription.New` が登録した時刻を入れ、新着を選んだときは `TakeNewVideos` が
  進めます。基準時刻とタイトルを変えるメソッドは `TakeNewVideos` だけです（新着を選ばずに基準時刻
  だけを進める呼び方をできないようにするため）。
  `playlists.updated_at`（GORM が保存のたびに書き換える）はレコードにだけあり、判定に使いません。
- `NotifiedUntil` を過去に戻せば、その時刻以降の動画を再通知できます。
- プレイリストへの追加時刻は `PlaylistItem.AddedAt`、動画の投稿時刻は `Video.PublishedAt` です。

## その他

- 各ジョブは `guard` で包み、`recover` で panic を捕まえて、1 回の失敗でボット全体が落ちない
  ようにしています。新しいジョブも `guard` で包みます。
- 通知の Embed に出す投稿者は送る直前に `library.YouTube.FetchLiveVideos` で取り、取れなかったときは
  投稿者なしで送ります。同じ動画の重複は YouTube のクライアントが除きます。
- 定期のまとめは、その年に追加がなければ（`Wrapped.Empty`）投稿しません。
