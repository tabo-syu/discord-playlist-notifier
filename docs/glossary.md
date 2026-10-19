# 用語集（ユビキタス言語）

このボットのコード・コメント・issue・会話で使う言葉と、その意味をまとめます。ボット全体が
1 つの境界づけられたコンテキストなので（[コンテキストマップ](context-map.md)）、ここに載せた言葉は
どこで使っても同じ意味です。決めた経緯は
[issue #76 の設計コメント](https://github.com/tabo-syu/discord-playlist-notifier/issues/76#issuecomment-6276539290) にあります。

- 名前を付けるとき・変えるときは、ここにある言葉を使います。ない言葉を新しく使うときは、
  同じ PR でここに足します。
- 1 つの言葉は 1 つの意味で使います。似た言葉が別の意味を持つものは
  「[使い分けに注意する言葉](#使い分けに注意する言葉)」にまとめています。
- 「コード」の列は今のコードの名前です。名前を変えると決めたもの（※）は
  「[名前の揺れの直し方](#名前の揺れの直し方)」に変えた後の名前があります。

## 名前の付け方の決まり

1. **外（YouTube・Discord）が発行した、自分自身の識別子は、発行元の名前にする**（`Video.YoutubeID`、
   `Guild.DiscordID`）。ID は YouTube や Discord が発行したものをそのまま使う（順応者）ことを表します。
   保存したときにリポジトリが振る識別子は `ID` です（`Guild.ID`、`Subscription.ID`）。
2. **他の集約を指すフィールドは「参照先の名前 + ID」にする**（`Subscription.GuildID`、
   `FetchedItem.VideoID`）。エンティティでない値（`FetchedItem`・`VideoDetails` など）が集約を指すときも
   同じです。参照先の識別子が値オブジェクトなら、その型の名前（`library.PlaylistID` など）とそろいます。
3. **1 つの言葉は 1 つの意味にする**。2 つの意味を持ってしまった言葉は、片方に別の名前を付けます。
4. **DB のテーブルとカラムの名前は変えない**。ドメインの名前を変えるときは、
   `infrastructure/persistence/records.go` の変換の関数で対応させます（[DB の名前との対応](#db-の名前との対応)）。

## Discord とサーバー

| 用語 | コード | 意味 |
|---|---|---|
| サーバー（ギルド） | `guild.Guild`・`guild.DiscordID` | ボットが参加している Discord のサーバー。画面に出す文言では「サーバー」、コードでは Discord API の言葉の guild と呼ぶ（同じもの）。`Guild.ID` は保存したときに振られる識別子で、通知設定が参照する |
| テキストチャンネル | `subscription.ChannelID`※・`Subscription.SendChannelID`※ | 通知を投稿する Discord のチャンネル。通知設定ごとに 1 つで、`add` を実行したチャンネルになる。「チャンネル」とだけ言うと YouTube のチャンネルと区別できないので、必ず「テキストチャンネル」と呼ぶ |
| コマンド | `/playlist-notifier` のサブコマンド | ギルドで使えるスラッシュコマンド。名前と、この用語集の言葉との対応は [外に公開している名前](#外に公開している名前) |

## 通知（コアのサブドメイン）

| 用語 | コード | 意味 |
|---|---|---|
| 通知設定 | `subscription.Subscription` | ギルドがプレイリストの通知を受けるための設定。プレイリスト・テキストチャンネル・ピックの間隔・新着の基準時刻を持つ。同じギルドに同じプレイリストの通知設定は 1 つだけ |
| 登録する・解除する | `SubscriptionService.Subscribe`・`Unsubscribe`、コマンドの `add`・`delete` | 通知設定を作る・論理削除する。「通知登録されている」は「登録されている」と同じ意味で、名詞としては「通知設定」を使う |
| 通知対象のプレイリスト | `application.WatchedPlaylist`※ | どこかのギルドが登録しているプレイリスト。ライブラリが写すのも、MCP が見せるのも、これだけ |
| プレイリストのタイトル（通知設定の写し） | `Subscription.Title`※ | 表示に使う、プレイリストのタイトルの写し（キャッシュ）。登録したときと、新着を通知したときにだけ YouTube 上のタイトルに合わせる。そのため古いことがあり、同じプレイリストでもギルドによって違うことがある |
| 新着 | `Subscription.TakeNewVideos`・`subscription.VideosAdded` | 新着の基準時刻より後にプレイリストへ追加された、観られる動画。追加された時刻の順に 1 本ずつ通知する |
| 新着の基準時刻 | `Subscription.NotifiedUntil` | これより前に追加された動画は、通知済みか登録より前からあったので新着にしない（ちょうど同じ時刻は新着に含める）。登録した時刻から始まり、新着を選ぶたびに進む |
| 通知する | `application.Notifier` の `NotifyXxx` | ドメインで起きたこと（新着・非公開・節目）を、テキストチャンネルに知らせること |
| 投稿する（定期の投稿） | `Notifier.PostWrapped`・`PostPick`、`NotificationService.PostXxx` | 起きたことではなく、時刻が来たら行う投稿（年間のまとめ・ピック） |
| 非公開・削除の通知 | `notification.HiddenVideo`・`FindHiddenVideos` | 観られた動画が見られなくなったことの通知 |
| 節目 | `library.ViewMilestones`・`ViewCount.Milestone` | 通知する価値のある再生数（10 万・100 万・1000 万・1 億回） |
| 節目の通知 | `notification.Milestone`・`FindMilestones` | 動画の再生数が新しい節目に達したことの通知 |
| 振り分け | `notification.FindHiddenVideos`・`FindMilestones` | 動画のイベントを、その動画を含むプレイリストの通知設定のテキストチャンネルへ割り当てること。同じテキストチャンネルには 1 回だけ送る |
| ピック | `subscription.PickInterval`・`NotificationService.PostPicks` | 決まった間隔（毎日・毎週月曜・毎月 1 日の 12:00）に、プレイリストからランダムに選んだ 1 本を投稿すること。画面では「今日の一本」「今週の一本」「今月の一本」 |
| 年間のまとめ | `library.Wrapped` | その年にプレイリストへ追加された動画のまとめ。12/31 21:00 に投稿し、`wrapped` コマンドでも見られる。追加がなければ（`Wrapped.Empty`）投稿しない |

## ライブラリ（プレイリストの写し）

| 用語 | コード | 意味 |
|---|---|---|
| ライブラリ | `domain/library` | 通知対象のプレイリストの中身の写し。プレイリストごとに 1 つだけ保存し、全てのギルドで共有する |
| 取得元 | `library.Source` | ライブラリの中身を取ってくる先。実装は YouTube Data API（`infrastructure/youtube`） |
| プレイリスト | `library.Playlist`・`library.PlaylistID` | YouTube のプレイリスト。ライブラリでは、アイテムの並び（追加された時刻の順）を持つ集約。タイトルは保存しない |
| アイテム | `library.PlaylistItem`・`library.ItemID` | プレイリストの 1 項目。同じ動画を 2 回追加すれば 2 つのアイテムになる。アイテムの ID は動画の ID とは別 |
| 動画 | `library.Video`・`library.VideoID` | YouTube の動画の、わかっている最新の状態。動画ごとに 1 つで、複数のプレイリストで共有する |
| 詳細 | `library.VideoDetails`・`Video.DetailsFetched`・`HasDetails` | 動画のタイトル・再生数・投稿された時刻・チャンネルなど。一度でも取得したかを `DetailsFetched` で覚え、見られなくなっても消さない |
| 公開状態 | `library.PrivacyStatus` | 公開・限定公開・非公開・削除。動画の詳細からではなく、アイテムから取る |
| 視聴できる | `PrivacyStatus.Watchable` | 公開状態が公開か限定公開であること |
| 観られる（見せてよい） | `Video.Available` | 詳細を取得済みで、視聴できること。ユーザーに見せる動画はこれだけ |
| 見られなくなる | `PrivacyStatus.Hidden`・`library.VideoHidden` | 観られた動画が非公開か削除になること |
| チャンネル（YouTube のチャンネル） | `library.Channel`・`library.ChannelID` | 動画を投稿した YouTube のチャンネル。名前は、そのチャンネルの動画を最後に更新したときのもの。アイコンは保存しない |
| カテゴリ | `library.Category`・`library.CategoryID` | YouTube の動画カテゴリ（音楽、ゲームなど）。名前はボットの地域と言語で YouTube が表示するもの |
| 再生数 | `library.ViewCount`・`Video.Views` | 動画の再生回数 |
| 追加された時刻 | `PlaylistItem.AddedAt` | 動画がプレイリストに追加された時刻。新着・統計・まとめはこちらを使う |
| 投稿された時刻 | `Video.PublishedAt` | 動画そのものが YouTube に投稿された時刻 |
| 中身 | `library.PlaylistVideo`・`Catalog.Contents` | アイテムと、その動画の組の並び |
| カタログ | `library.Catalog` | 保存済みの中身を、プレイリストと動画の 2 つの集約をまたいで読むドメインサービス。見せてよい範囲（指定したプレイリストにある、詳細を取得済みの動画）を決める |
| 動画の記録（スナップショット） | `library.VideoSnapshot` | ある時点の動画の状態（タイトル・公開状態・再生数・高評価数・コメント数）。更新のたびに、観られる動画のうち YouTube が詳細を返したものについて追記し、書き換えない。見られなくなった動画は記録しない |

## 同期と更新

| 用語 | コード | 意味 |
|---|---|---|
| 同期 | `LibraryService.Sync` | 5 分ごとに、プレイリストのアイテムと動画の公開状態を YouTube に合わせること。詳細をまだ持っていない視聴できる動画は、詳細も取る |
| 全件取得 | `library.FullSync`・`FULL_SYNC_INTERVAL` | 同期でアイテムを全て取り直すこと。クォータを節約するため、アイテム数が変わったときか 1 時間経ったときだけ行う |
| メタ情報 | `library.PlaylistMeta` | プレイリストのタイトルとアイテム数。全件取得するかの判断に使う |
| 取得したアイテム | `library.FetchedItem` | YouTube から取得したアイテム 1 件（公開状態を含む） |
| 突き合わせ | `Playlist.Reconcile`・`library.SyncVideos` | 保存済みのアイテムと動画を、取得したものに合わせること。アイテムはアイテムの ID で比べる |
| 同期したプレイリスト | `library.Snapshot`※ | 同期した直後のプレイリストの、アイテム・YouTube 上のタイトル・YouTube から消えたか・動画 |
| 同期の結果 | `library.PlaylistSynced` | プレイリストを同期し終えたことを表すドメインイベント。同期したプレイリストを運ぶ。YouTube から消えていたプレイリストでも出す |
| 更新 | `LibraryService.RefreshVideos` | 6 時間ごとに、観られる動画の詳細と再生数を取り直し、動画の記録を追記し、節目を調べること。詳細をまだ持っていない動画と見られなくなった動画は対象外で、それらの詳細は同期が取る |
| カテゴリの更新 | `LibraryService.RefreshCategories` | 起動時と 1 日ごとに、カテゴリの名前を取り直すこと |
| ドメインイベント | `library.VideoHidden`・`MilestoneReached`・`PlaylistSynced`、`subscription.VideosAdded` | ドメインで起きたこと。名前は過去形で、それを起こした集約を保存できた後に配る |

## 分析

#93 で `domain/analytics` に分けます。

| 用語 | コード | 意味 |
|---|---|---|
| 統計 | `library.Stats`・`ComputeStats` | プレイリストの動画数（観られる・非公開・削除の内訳）、追加のペース、再生数のランキング、月ごとの追加数 |
| 年間のまとめ | `library.Wrapped`・`ComputeWrapped` | 上の「通知」を参照 |
| 伸びている動画 | `library.Trending`・`ComputeTrending` | 期間中に再生数が最も伸びた動画。伸びは、期間中の最初の記録から最後の記録までの再生数の差 |
| 推移 | `library.VideoHistory` | 1 本の動画の、期間中の記録の並び |

## 外に公開している名前

使う人やほかのボットの設定が覚えている名前なので、用語集の言葉とずれていても変えません。

| 公開している名前 | 用語集の言葉 |
|---|---|
| コマンド `add`・`delete` | 登録する・解除する |
| コマンド `list` | このギルドの通知設定の一覧 |
| コマンド `pick` | ピックの間隔を決める |
| コマンド `summary`、MCP `playlist_stats` | 統計 |
| コマンド `trending` | 伸びている動画 |
| コマンド `history`、MCP `video_history` | 推移 |
| コマンド `random`、MCP `random_videos` | 中身からランダムに選ぶ（ピックとは別。投稿はしない） |
| コマンド `wrapped`、MCP `playlist_wrapped` | 年間のまとめ |
| コマンド `source` | （ドメインの言葉ではない）このボットのソースコードへのリンク |
| MCP `list_playlists` | 通知対象のプレイリスト（と統計）の一覧 |
| MCP `find_videos` | 通知対象のプレイリストの中身を絞り込む |

## DB の名前との対応

テーブルとカラムの名前は変えないので、ドメインの名前とずれているものがあります。

| テーブル・カラム | 用語集の言葉 |
|---|---|
| `guilds` | サーバー（ギルド） |
| `playlists` | **通知設定**（ライブラリのプレイリストではない。モデルが `Playlist` という名前だった頃のまま） |
| `playlists.youtube_id` | プレイリストの ID |
| `playlists.send_channel_id` | テキストチャンネルの ID |
| `playlists.title` | プレイリストのタイトル（通知設定の写し） |
| `playlists.notified_until` | 新着の基準時刻 |
| `playlist_items` | アイテム。ライブラリのプレイリスト（集約）は、このテーブルの行の集まりで、プレイリストそのもののテーブルはない |
| `playlist_items.playlist_youtube_id`・`video_youtube_id` | プレイリストの ID・動画の ID |
| `videos`・`videos.channel_youtube_id` | 動画・チャンネルの ID |
| `channels`・`categories` | チャンネル（YouTube）・カテゴリ |
| `video_snapshots`・`video_snapshots.video_youtube_id` | 動画の記録・動画の ID |

## 使い分けに注意する言葉

- **チャンネル**：YouTube のチャンネル（`library.Channel`）だけを指します。Discord のほうは
  「テキストチャンネル」と呼びます。
- **スナップショット**：動画の記録（`VideoSnapshot`）だけを指します。同期した直後のプレイリストの中身は
  「同期したプレイリスト」、それを運ぶイベントは「同期の結果」（`PlaylistSynced`）と呼びます。
- **プレイリスト**：YouTube のプレイリストのことで、コードでは `library.Playlist` です。
  ギルドごとの設定は「通知設定」で、テーブル名の `playlists` はこちらなので注意します。
- **視聴できる・観られる・見られなくなる**：視聴できる（`Watchable`）は公開状態だけの話、観られる
  （`Available`）は詳細も取得済みでユーザーに見せてよいこと、見られなくなる（`Hidden`）は非公開か
  削除になったことです。
- **通知する・投稿する**：通知は起きたことを知らせるもの（新着・非公開・節目）、投稿は時刻が来たら
  行うもの（ピック・年間のまとめ）です。
- **ピック・ランダム**：ピックは決まった間隔で投稿するもの、ランダム（`random` コマンド、
  `PickRandom`）はその場で選んで見せるものです。選び方（`PickRandom`）は同じです。
- **同期・更新**：同期（5 分）はアイテムと公開状態、更新（6 時間）は詳細と再生数です。
- **追加された時刻・投稿された時刻**：前者はプレイリストへの追加（`AddedAt`）、後者は動画そのもの
  （`PublishedAt`）です。

## 名前の揺れの直し方

上の決まりに合わせるため、次の名前を変えます。作業は #92 で行い、終わったら上の表の
「コード」の列とこの節を直します。DB のテーブルとカラムの名前は変えません。

| 今 | 直した後 | 当てはめた決まり |
|---|---|---|
| `Subscription.YoutubeID` | `Subscription.PlaylistID` | 2 |
| `PlaylistItem.PlaylistYoutubeID` | `PlaylistItem.PlaylistID` | 2 |
| `PlaylistItem.VideoYoutubeID` | `PlaylistItem.VideoID` | 2 |
| `Video.ChannelYoutubeID` | `Video.ChannelID` | 2 |
| `VideoSnapshot.VideoYoutubeID` | `VideoSnapshot.VideoID` | 2 |
| `PlaylistMeta.YoutubeID` | `PlaylistMeta.PlaylistID` | 2（エンティティでない値が指す ID。`VideoDetails.VideoID` とそろう） |
| `subscription.ChannelID`（型） | `subscription.TextChannelID` | 3（チャンネル） |
| `Subscription.SendChannelID`、`VideosAdded`・`notification.HiddenVideo`・`notification.Milestone` の `ChannelID` | `TextChannelID`（引数名・変数名と、Discord のチャンネルを指すコメントの「チャンネル」も「テキストチャンネル」に） | 3（チャンネル） |
| `library.Snapshot` | `library.SyncedPlaylist`（`LibraryService.snapshot`、変数名、コメントの「スナップショット」も） | 3（スナップショット） |
| `Subscription.Title` | `Subscription.PlaylistTitle` | 何のタイトルかを名前で表す |
| `application.WatchedPlaylist` | `application.SubscribedPlaylist` | 3（`Watchable` と紛らわしい） |

分析のパッケージを分ける作業（#93）と、読み取りのユースケースを `QueryService` にまとめる作業（#94）は
[コンテキストマップ](context-map.md) にあります。
