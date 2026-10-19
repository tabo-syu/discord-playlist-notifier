package library

// library で起きたことを表すドメインイベント。名前は起きたことを過去形で表す。
// 動画のイベントは集約のメソッドが戻り値で返し、アプリケーションサービスが動画を
// 保存した後に配る。
//
// 通知にはタイトルなど動画の詳細がほぼ全て要るので、動画はコピーせずに持たせる。
// 配った後に動画を書き換えない。

// VideoHidden は、視聴できた動画が非公開・削除になったこと。
type VideoHidden struct {
	Video *Video
}

// MilestoneReached は、動画の再生数が新しい節目に達したこと。
type MilestoneReached struct {
	Video     *Video
	Milestone ViewCount
}

// PlaylistSynced は、通知対象のプレイリストの中身を YouTube と同期し終えたこと。
// YouTube から消えていたプレイリストも、Snapshot.Deleted を立てて出す。
//
// 同期の結果は複数の集約（Playlist と Video）から組み立てるので、集約のメソッドでは
// なく、それを組み立てた LibraryService がプレイリストと動画を保存した後に出す。
// 購読する側（新着を選んで通知設定を進める）が、同期した側を知らずに済むようにする。
type PlaylistSynced struct {
	Snapshot *Snapshot
}
