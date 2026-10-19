package library

// 動画の集約が起こすドメインイベント。名前は起きたことを過去形で表す。集約の
// メソッドが戻り値で返し、アプリケーションサービスが動画を保存した後に配る。
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
