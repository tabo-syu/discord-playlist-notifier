---
paths:
  - "internal/domain/library/**"
  - "internal/application/library_service*.go"
  - "internal/infrastructure/youtube/**"
  - "internal/infrastructure/persistence/**"
---

# プレイリストの中身と YouTube API

プレイリストの中身はプレイリスト ID 単位で保存し、全サーバーで共有します。各テーブルの役割と
書き込むタイミングは `README.md` の「データのライフサイクル」にあります。

- 動画が非公開・削除になっても、タイトルなどの詳細は消さずに残し、ユーザーには
  `Available()` が真の動画だけを見せます。
- `videos.list` の part を増やしてもクォータは変わらないので、必要な情報は同じ呼び出しで取ります。
- 動画の公開状態はアイテムの `status.privacyStatus` から取ります。`videos.list` は非公開・
  削除済みの動画を返さないためです。
- `Sync` は `FULL_SYNC_INTERVAL` 以内でアイテム数が変わらなければ全件取得を省き、`videos.list` は
  詳細を持っていない動画の分だけ呼びます。全件取得の状態はメモリにあるので、再起動直後は全件
  取得します。
- `Sync` と `RefreshVideos` は同じ `videos` を書き換えるので、`LibraryService` の mutex で
  直列化しています。

## YouTube API のレスポンス

- 一部のフィールドが欠けることがあるので、nil を考慮して読みます。
- プレイリストアイテムの `snippet.channelId` は、ドキュメント上は「追加したユーザー」ですが、
  実際には常にプレイリストの所有者を返します。追加者は取得できません。

## スキーマ

- `subscription.Subscription` は以前 `Playlist` だったので、テーブル名は `playlists` のままです
  （`TableName()` で固定）。
- `videos` テーブルは以前、通知した動画の履歴でした。古い形（`playlist_id` カラムがある）は、
  `persistence.Migrate`（`cmd/migrate`）が一度だけ削除します。
