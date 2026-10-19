package library

import "sort"

// Playlist は通知対象の YouTube プレイリストと、その保存済みのアイテムを持つ集約の
// ルート。アイテムにはこのルートを通してだけ触れる。動画は別の集約で、アイテムが
// ID で参照する。タイトルなど YouTube 上の情報は保存しないので持たない（PlaylistMeta）。
type Playlist struct {
	YoutubeID PlaylistID
	// 追加された時刻の順
	Items []*PlaylistItem
}

// Reconcile は保存済みのアイテムを YouTube から取得したアイテムに置き換え、
// 追加したアイテムと除いたアイテムを返す。アイテムは YouTube のアイテム ID で
// 突き合わせる。
func (p *Playlist) Reconcile(fetched []*FetchedItem) (added, removed []*PlaylistItem) {
	fetchedIDs := map[ItemID]bool{}
	for _, f := range fetched {
		fetchedIDs[f.ItemID] = true
	}

	storedIDs := map[ItemID]bool{}

	var items []*PlaylistItem

	for _, item := range p.Items {
		storedIDs[item.ItemID] = true
		if fetchedIDs[item.ItemID] {
			items = append(items, item)
		} else {
			removed = append(removed, item)
		}
	}

	for _, f := range fetched {
		if storedIDs[f.ItemID] {
			continue
		}

		storedIDs[f.ItemID] = true
		item := &PlaylistItem{
			PlaylistYoutubeID: p.YoutubeID,
			ItemID:            f.ItemID,
			VideoYoutubeID:    f.VideoID,
			AddedAt:           f.AddedAt,
		}
		added = append(added, item)
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].AddedAt.Before(items[j].AddedAt)
	})
	p.Items = items

	return added, removed
}

// Contains はプレイリストにその動画のアイテムがあるかを返す。
func (p *Playlist) Contains(videoID VideoID) bool {
	for _, item := range p.Items {
		if item.VideoYoutubeID == videoID {
			return true
		}
	}

	return false
}
