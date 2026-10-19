package library

// Source はライブラリ（プレイリストの中身）の取得元で、実装は YouTube Data API。
// 取得した結果はドメインの値（PlaylistMeta・FetchedItem・VideoDetails など）で返し、
// エンティティを作ったり変えたりするのはドメインに任せる。1 日のクォータに上限が
// あるので、保存済みの中身（PlaylistRepository など）で足りるときはそちらを使う。
type Source interface {
	// FindPlaylist はプレイリストが存在しないか非公開のときに
	// ErrPlaylistNotFound を返す。アイテム数は取得しない。
	FindPlaylist(id PlaylistID) (*PlaylistMeta, error)
	// FetchPlaylistMetas はもう存在しないプレイリストを結果から除く。
	FetchPlaylistMetas(ids ...PlaylistID) (map[PlaylistID]*PlaylistMeta, error)
	FetchPlaylistItems(id PlaylistID) ([]*FetchedItem, error)
	// FetchVideoDetails は指定した動画の詳細を返す。見られない動画（非公開、
	// 削除済み）は結果に含まれない。
	FetchVideoDetails(ids []VideoID) ([]*VideoDetails, error)
	// FetchCategories はボットの地域の動画カテゴリを、その言語での名前で
	// 返す。
	FetchCategories() ([]*Category, error)
}
