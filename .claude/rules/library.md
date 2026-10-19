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
- 同期の突き合わせは 2 つに分かれます。アイテムは `Playlist.Reconcile`（集約のメソッド）、
  動画の公開状態と詳細を取る動画の選択は `library.SyncVideos`（複数の `Video` をまとめて見る
  ドメインサービス）です。
- 同期では `PlaylistRepository.Save` を `VideoRepository.Save` より先に呼びます。動画の保存に
  失敗してもその回はエラーになり、次の全件取得で公開状態の変化と新しい動画をもう一度検知します。
  逆の順にすると、非公開になった通知が失われることがあります。同じ回に同期した別のプレイリストにも
  同じ動画があって、そちらで公開状態を保存できたときは、失敗したプレイリストのチャンネルには
  通知されません（1 つのトランザクションだったときからある制限です）。
- `PlaylistRepository.Save` はプレイリストを丸ごと受け取り、DB の行と `item_id` で比べて、
  外れたアイテムを削除し増えたアイテムを追加します（残ったアイテムの行は書き換えません）。

## YouTube API のレスポンス

- 一部のフィールドが欠けることがあるので、nil を考慮して読みます。
- プレイリストアイテムの `snippet.channelId` は、ドキュメント上は「追加したユーザー」ですが、
  実際には常にプレイリストの所有者を返します。追加者は取得できません。

## スキーマ

- `subscription.Subscription` は以前 `Playlist` だったので、テーブル名は `playlists` のままです
  （`persistence` の `subscriptionRecord.TableName()` で固定）。
- `playlists` の `(guild_id, youtube_id)` には一意の部分インデックス `idx_playlists_guild_youtube`
  （`WHERE deleted_at IS NULL`）があります。解除した通知設定は論理削除なので、部分インデックスに
  しないと登録し直せません。`AutoMigrate` は同じ名前のインデックスがあれば定義を比べないので、
  定義を変えるときは名前も変えます。論理削除していない重複の行があるとインデックスを作れず、
  `cmd/migrate` が失敗します（重複は手で直します）。
- 新着動画の通知の基準時刻は、以前は `playlists.updated_at` でした。今は `notified_until` で、
  `persistence.Migrate` が空の行を毎回 `updated_at` で埋めます（何度実行しても同じ結果になる）。
- 動画の詳細を取得済みかは、以前はタイトルが空でないかで判断していました。今は
  `videos.details_fetched` で、`persistence.Migrate` が空（NULL）の行を毎回タイトルの有無で埋めます。
- `videos` テーブルは以前、通知した動画の履歴でした。古い形（`playlist_id` カラムがある）は、
  `persistence.Migrate`（`cmd/migrate`）が一度だけ削除します。
