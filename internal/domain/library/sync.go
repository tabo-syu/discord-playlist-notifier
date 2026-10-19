package library

import (
	"maps"
	"slices"
	"time"
)

// PlaylistMeta はプレイリストのアイテムを取得し直す必要があるかの判断に使う、
// プレイリスト単位の情報。
type PlaylistMeta struct {
	YoutubeID PlaylistID
	Title     string
	ItemCount int64
}

// プレイリストのアイテム数が変わっていなくても、この間隔が経てばアイテムを
// 取得し直す。動画の除外や公開状態の変化と同時に起きた追加を拾うため。
const FULL_SYNC_INTERVAL = time.Hour

// FullSync は、プレイリストのアイテムを最後に全件取得したときの記録。アイテムの
// 取得はプレイリストの大きさに比例してクォータを使うので、変わっていなさそうな
// ときは省く。ゼロ値は、まだ全件取得していないことを表す。
type FullSync struct {
	itemCount int64
	at        time.Time
}

// NewFullSync は、meta のプレイリストのアイテムを at に全件取得したことを記録する。
func NewFullSync(meta *PlaylistMeta, at time.Time) FullSync {
	return FullSync{itemCount: meta.ItemCount, at: at}
}

// Due は、今回の同期でアイテムを全件取得し直すべきかを返す。まだ全件取得して
// いないか、前回からアイテム数が変わったか、FULL_SYNC_INTERVAL が経ったときに
// 真になる。
func (f FullSync) Due(meta *PlaylistMeta, now time.Time) bool {
	return f.at.IsZero() || f.itemCount != meta.ItemCount || now.Sub(f.at) >= FULL_SYNC_INTERVAL
}

// FetchedItem は YouTube から取得したプレイリストのアイテム 1 件。
type FetchedItem struct {
	ItemID        ItemID
	VideoID       VideoID
	AddedAt       time.Time
	PrivacyStatus PrivacyStatus
}

// VideoSync は、YouTube から取得したアイテムの公開状態を保存済みの動画に反映した結果。
type VideoSync struct {
	// 取得したアイテムの全ての動画。初めて見る動画も含む
	Videos map[VideoID]*Video
	// 非公開・削除になった動画のイベント。動画を保存した後に配る
	Hidden []*VideoHidden
	// 視聴可能で、詳細をまだ一度も取得していない動画
	NeedDetails []VideoID

	changed map[VideoID]*Video
}

// SyncVideos は取得したアイテムの公開状態を、その動画に反映する。複数の動画の
// 集約をまとめて見るので、どの集約にも属さないドメインサービスにしている。
// videos には取得したアイテムの保存済みの動画を渡す。このマップは返す VideoSync が
// 引き継ぎ、初めて見る動画を追加する。
//
// 動画の公開状態はプレイリストのアイテムから取る。videos のエンドポイントとは
// 違い、アイテムは非公開・削除済みの動画も返すため。
func SyncVideos(fetched []*FetchedItem, videos map[VideoID]*Video) *VideoSync {
	s := &VideoSync{Videos: videos, changed: map[VideoID]*Video{}}

	needDetails := map[VideoID]bool{}

	for _, f := range fetched {
		video, ok := videos[f.VideoID]
		if !ok {
			video = &Video{YoutubeID: f.VideoID}
			videos[f.VideoID] = video
		}

		changed, hidden := video.ChangePrivacy(f.PrivacyStatus)
		if hidden != nil {
			s.Hidden = append(s.Hidden, hidden)
		}

		if changed {
			s.changed[f.VideoID] = video
		}

		if !video.HasDetails() && f.PrivacyStatus.Watchable() && !needDetails[f.VideoID] {
			needDetails[f.VideoID] = true
			s.NeedDetails = append(s.NeedDetails, f.VideoID)
		}
	}

	return s
}

// ApplyDetails は取得した詳細を NeedDetails の動画にコピーする。
func (s *VideoSync) ApplyDetails(details []*Video) {
	for _, d := range details {
		video := s.Videos[d.YoutubeID]
		video.ApplyDetails(d)
		video.InitMilestone()
		s.changed[d.YoutubeID] = video
	}
}

// ChangedVideos は保存が必要な動画を返す。
func (s *VideoSync) ChangedVideos() []*Video {
	return slices.Collect(maps.Values(s.changed))
}
