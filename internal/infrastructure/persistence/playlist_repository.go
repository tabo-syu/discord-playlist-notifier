package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
)

type playlistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *playlistRepository {
	return &playlistRepository{db}
}

func (r *playlistRepository) Find(id library.PlaylistID) (*library.Playlist, error) {
	var recs []*playlistItemRecord

	err := r.db.Where("playlist_youtube_id = ?", id).Order("added_at").Find(&recs).Error
	if err != nil {
		return nil, err
	}

	return &library.Playlist{YoutubeID: id, Items: mapSlice(recs, (*playlistItemRecord).toDomain)}, nil
}

// サブクエリで、動画を含むプレイリストの全てのアイテムを 1 回で読む。
func (r *playlistRepository) FindContaining(videoIDs []library.VideoID) ([]*library.Playlist, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}

	containing := r.db.Model(&playlistItemRecord{}).Select("playlist_youtube_id").Where("video_youtube_id IN ?", videoIDs)

	var recs []*playlistItemRecord

	err := r.db.Where("playlist_youtube_id IN (?)", containing).Order("playlist_youtube_id, added_at").Find(&recs).Error
	if err != nil {
		return nil, err
	}

	var playlists []*library.Playlist

	for _, rec := range recs {
		id := library.PlaylistID(rec.PlaylistYoutubeID)
		if len(playlists) == 0 || playlists[len(playlists)-1].YoutubeID != id {
			playlists = append(playlists, &library.Playlist{YoutubeID: id})
		}

		p := playlists[len(playlists)-1]
		p.Items = append(p.Items, rec.toDomain())
	}

	return playlists, nil
}

// Save は保存済みの行と YouTube のアイテム ID（一意）で突き合わせ、外れたアイテムを
// 削除し、増えたアイテムを追加する。残ったアイテムの行は書き換えない。
func (r *playlistRepository) Save(p *library.Playlist) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var storedIDs []string

		err := tx.Model(&playlistItemRecord{}).Where("playlist_youtube_id = ?", p.YoutubeID).Pluck("item_id", &storedIDs).Error
		if err != nil {
			return err
		}

		keep := map[string]bool{}
		for _, item := range p.Items {
			keep[string(item.ItemID)] = true
		}

		stored := map[string]bool{}

		var removed []string

		for _, id := range storedIDs {
			stored[id] = true
			if !keep[id] {
				removed = append(removed, id)
			}
		}

		var added []*playlistItemRecord

		for _, item := range p.Items {
			if !stored[string(item.ItemID)] {
				added = append(added, toPlaylistItemRecord(item))
			}
		}

		if len(removed) > 0 {
			if err = tx.Where("playlist_youtube_id = ? AND item_id IN ?", p.YoutubeID, removed).Delete(&playlistItemRecord{}).Error; err != nil {
				return err
			}
		}

		if len(added) > 0 {
			return tx.CreateInBatches(added, 500).Error
		}

		return nil
	})
}
