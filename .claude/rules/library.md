---
paths:
  - "internal/domain/library/**"
  - "internal/application/library_service*.go"
  - "internal/infrastructure/youtube/**"
  - "internal/infrastructure/persistence/**"
---

# プレイリストの中身と YouTube API

YouTube プレイリストの中身は、サーバーごとの通知設定（`subscription.Subscription`）とは
別に、プレイリスト ID 単位で保存して全サーバーで共有します。各テーブルの役割と、誰がいつ
書き込むかは `README.md` の「データのライフサイクル」にあります。

## 保存のしかた

- 動画が非公開・削除になっても、タイトルなどの詳細は消さずに残します。ユーザーに見せるのは
  `Available()` が真の動画だけです。
- サムネイルは動画 ID から作れるので保存しません（`Thumbnail()`）。チャンネルのアイコンも
  保存しません。
- `videos.list` の part を増やしてもクォータは変わらないので、長さ・タグ・カテゴリなども
  同じ呼び出しで取ります。
- `channels` は動画の詳細を取るたびに作るか名前を上書きし、`categories` は
  `videoCategories.list`（日本・日本語）で上書きします。地域ごとに一覧が違うので、固定の
  一覧は持ちません。どちらも失敗してもログを出すだけです。
- `video_snapshots` は `RefreshVideos` が 6 時間ごとに、更新した全動画の分を同じ時刻で追記
  します。値は `videos.list` の結果から取るのでクォータは増えません。古い行を消す処理は
  ありません。

## `LibraryService.Sync`

1. `playlists.list` でアイテム数を確かめ、数が変わったとき、または前回の全件取得から
   `FULL_SYNC_INTERVAL`（1 時間）が経ったときだけアイテムを全件取得します。全件取得の状態は
   メモリにあるので、再起動直後は全件取得します。
2. 保存済みと取得したアイテムの突き合わせ（追加・除外・公開状態の変化・詳細が要る動画）は
   `library.Reconcile` が行います。
3. `videos.list` は、まだ詳細を持っていない動画の分だけ呼びます。
4. 公開状態はアイテムの `status.privacyStatus` から取ります（`videos.list` は非公開・削除済みの
   動画を返さないため）。

`Sync` と `RefreshVideos` は同じ `videos` を書き換えるので、`LibraryService` の mutex で
直列化しています。

## YouTube API（`infrastructure/youtube/client.go`）

- 呼び出しは `MAX_BATCH_SIZE = 50` でバッチ化し、プレイリストアイテムは
  `MAX_RESULTS_PER_PAGE = 50` でページングします。
- レスポンスは一部のフィールドが欠けることがあるので、nil を考慮して読みます。
- タイムスタンプのパースには `YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"` を使います。
- プレイリストアイテムの `snippet.channelId` は、ドキュメント上は「追加したユーザー」ですが、
  実際には常にプレイリストの所有者を返します。追加者は取得できません。

## スキーマ（`persistence`）

- 各モデルは `TableName()` でテーブル名を固定しています。`subscription.Subscription` は以前
  `Playlist` だったので、テーブル名は `playlists` のままです。
- `videos` テーブルは以前、通知した動画の履歴でした。古い形（`playlist_id` カラムがある）は、
  起動時に `persistence.Migrate` が一度だけ削除します。
