package library

import "time"

// リポジトリは集約ごとに 1 つ置く。集約同士は ID で参照するので、別の集約の ID を
// 検索の条件として受け取るメソッドはあるが、別の集約を保存するメソッドはない。

// PlaylistRepository は Playlist 集約を保存する。
type PlaylistRepository interface {
	// Find はプレイリストを、保存済みのアイテムと一緒に返す。アイテムをまだ保存して
	// いなければ、アイテムのないプレイリストを返す。
	Find(id PlaylistID) (*Playlist, error)
	// FindContaining は指定した動画のどれかを含むプレイリストを、全てのアイテムと
	// 一緒に返す。
	FindContaining(videoIDs []VideoID) ([]*Playlist, error)
	// Save は保存済みのアイテムを、プレイリストのアイテムに置き換える。
	Save(p *Playlist) error
}

// VideoRepository は Video 集約を保存する。
type VideoRepository interface {
	Find(ids []VideoID) (map[VideoID]*Video, error)
	// FindInPlaylists は指定したプレイリストの全アイテムの動画を返す。
	FindInPlaylists(playlistIDs ...PlaylistID) (map[VideoID]*Video, error)
	// FindListed はいずれかのプレイリストに含まれる全ての動画を返す。
	FindListed() (map[VideoID]*Video, error)
	// Save は指定した動画の全カラムを上書きする。動画同士に不変条件はないので、
	// DB への往復を減らすためにまとめて保存する。
	Save(videos []*Video) error
}

// ChannelRepository は Channel 集約を保存する。
type ChannelRepository interface {
	// Find は指定した ID のうち保存済みのチャンネルを返す。
	Find(ids []ChannelID) (map[ChannelID]*Channel, error)
	// Save は指定したチャンネルを追加するか、名前を更新する。
	Save(channels []*Channel) error
}

// CategoryRepository は Category 集約を保存する。
type CategoryRepository interface {
	// FindAll は保存済みの全てのカテゴリを返す。
	FindAll() (map[CategoryID]*Category, error)
	// Save は指定したカテゴリを追加するか、更新する。
	Save(categories []*Category) error
}

// VideoSnapshotRepository は動画の状態の履歴を保存する。履歴は追記するだけで
// 書き換えない。
type VideoSnapshotRepository interface {
	// Add は指定した動画の状態を履歴に追記する。
	Add(snapshots []*VideoSnapshot) error
	// Find は since 以降に記録された動画の状態を、古い順に返す。
	Find(id VideoID, since time.Time) ([]*VideoSnapshot, error)
	// FindEarliest は指定した動画ごとに、since 以降に記録された最初の
	// 状態を返す。該当する状態がない動画は結果に含まれない。
	FindEarliest(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
	// FindLatest は指定した動画ごとに、since 以降に記録された最後の
	// 状態を返す。該当する状態がない動画は結果に含まれない。
	FindLatest(ids []VideoID, since time.Time) (map[VideoID]*VideoSnapshot, error)
}
