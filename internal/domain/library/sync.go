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

// FetchedItem は YouTube から取得したプレイリストのアイテム 1 件。
type FetchedItem struct {
	ItemID        ItemID
	VideoID       VideoID
	AddedAt       time.Time
	PrivacyStatus PrivacyStatus
	// YouTube が返したアイテムのタイトルと status.privacyStatus。PrivacyStatus を
	// どう判定したかをログで確かめるためだけに持つ
	Title            string
	RawPrivacyStatus string
}

// Sync はプレイリストの保存済みのアイテムと、YouTube から取得したアイテムを
// 突き合わせた結果。
type Sync struct {
	Added   []*PlaylistItem
	Removed []*PlaylistItem
	// 取得したアイテムの全ての動画。初めて見る動画も含む
	Videos map[VideoID]*Video
	// 非公開・削除になった動画
	Hidden []*Video
	// 視聴可能で、詳細をまだ一度も取得していない動画
	NeedDetails []VideoID

	changed map[VideoID]*Video
}

// Reconcile はプレイリストの保存済みのアイテムと取得したアイテムを突き合わせる。
// videos には取得したアイテムの保存済みの動画を渡す。このマップは返す Sync が
// 引き継ぎ、初めて見る動画を追加する。
//
// 動画の公開状態はプレイリストのアイテムから取る。videos のエンドポイントとは
// 違い、アイテムは非公開・削除済みの動画も返すため。
func Reconcile(playlistID PlaylistID, stored []*PlaylistItem, fetched []*FetchedItem, videos map[VideoID]*Video) *Sync {
	s := &Sync{Videos: videos, changed: map[VideoID]*Video{}}

	storedByID := map[ItemID]*PlaylistItem{}
	for _, item := range stored {
		storedByID[item.ItemID] = item
	}

	fetchedByID := map[ItemID]bool{}
	for _, f := range fetched {
		fetchedByID[f.ItemID] = true
		if _, ok := storedByID[f.ItemID]; !ok {
			s.Added = append(s.Added, &PlaylistItem{
				PlaylistYoutubeID: playlistID,
				ItemID:            f.ItemID,
				VideoYoutubeID:    f.VideoID,
				AddedAt:           f.AddedAt,
			})
		}
	}

	for _, item := range stored {
		if !fetchedByID[item.ItemID] {
			s.Removed = append(s.Removed, item)
		}
	}

	needDetails := map[VideoID]bool{}

	for _, f := range fetched {
		video, ok := videos[f.VideoID]
		if !ok {
			video = &Video{YoutubeID: f.VideoID}
			videos[f.VideoID] = video
		}

		changed, hidden := video.ChangePrivacy(f.PrivacyStatus)
		if hidden {
			s.Hidden = append(s.Hidden, video)
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
func (s *Sync) ApplyDetails(details []*Video) {
	for _, d := range details {
		video := s.Videos[d.YoutubeID]
		video.ApplyDetails(d)
		video.InitMilestone()
		s.changed[d.YoutubeID] = video
	}
}

// RestoreWatchable は、Hidden のうち videos のエンドポイントが見られる状態で
// 返した動画を Hidden から除き、公開状態をその値に戻す。playlistItems の
// エンドポイントは、公開中の動画をページ単位でまとめて削除済みとして返すことが
// あるため。live には videos のエンドポイントから取得した動画を渡す。戻した
// 動画を返す。
func (s *Sync) RestoreWatchable(live []*Video) []*Video {
	watchable := map[VideoID]PrivacyStatus{}

	for _, v := range live {
		if v.PrivacyStatus.Watchable() {
			watchable[v.YoutubeID] = v.PrivacyStatus
		}
	}

	var hidden, restored []*Video

	for _, v := range s.Hidden {
		status, ok := watchable[v.YoutubeID]
		if !ok {
			hidden = append(hidden, v)

			continue
		}

		v.ChangePrivacy(status)
		restored = append(restored, v)
	}

	s.Hidden = hidden

	return restored
}

// ChangedVideos は保存が必要な動画を返す。
func (s *Sync) ChangedVideos() []*Video {
	return slices.Collect(maps.Values(s.changed))
}
