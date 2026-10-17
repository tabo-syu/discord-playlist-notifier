package library

import "time"

// Repository は通知対象の YouTube プレイリストの中身を保存する。
type Repository interface {
	// FindItems は指定したプレイリストのアイテムを、追加された時刻の順に返す。
	FindItems(playlistIDs ...PlaylistID) ([]*PlaylistItem, error)
	FindVideos(ids []VideoID) (map[VideoID]*Video, error)
	// FindVideosInPlaylists は指定したプレイリストの全アイテムの動画を返す。
	FindVideosInPlaylists(playlistIDs ...PlaylistID) (map[VideoID]*Video, error)
	// FindListedVideos はいずれかのプレイリストに含まれる全ての動画を返す。
	FindListedVideos() (map[VideoID]*Video, error)
	// FindItemsByVideos は指定した動画を含む全プレイリストのアイテムを返す。
	FindItemsByVideos(ids []VideoID) ([]*PlaylistItem, error)
	// ApplyItems はプレイリストの保存済みのアイテムを指定したものに置き換え、
	// 動画を保存する。1 つのトランザクションで行う。
	ApplyItems(playlistID PlaylistID, added []*PlaylistItem, removed []*PlaylistItem, videos []*Video) error
	// SaveVideos は指定した動画の全カラムを上書きする。
	SaveVideos(videos []*Video) error
	// AddSnapshots は指定した動画の状態を履歴に追記する。
	AddSnapshots(snapshots []*VideoSnapshot) error
	// FindSnapshots は since 以降に記録された動画の状態を、
	// 古い順に返す。
	FindSnapshots(id VideoID, since time.Time) ([]*VideoSnapshot, error)
	// FindEarliestSnapshots は指定した動画ごとに、since 以降に記録された最初の
	// 状態を返す。該当する状態がない動画は結果に含まれない。
	FindEarliestSnapshots(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
	// FindLatestSnapshots は指定した動画ごとに、since 以降に記録された最後の
	// 状態を返す。該当する状態がない動画は結果に含まれない。
	FindLatestSnapshots(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
}
