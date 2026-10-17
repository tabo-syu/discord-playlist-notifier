package library

// YouTube はプレイリストの中身の取得元。1 日のクォータに上限があるので、
// 保存済みの中身（Repository）で足りるときはそちらを使う。
type YouTube interface {
	// FindPlaylist はプレイリストが存在しないか非公開のときに
	// domain.ErrYouTubePlaylistNotFound を返す。アイテム数は取得しない。
	FindPlaylist(id PlaylistID) (*PlaylistMeta, error)
	// FetchPlaylistMetas はもう存在しないプレイリストを結果から除く。
	FetchPlaylistMetas(ids ...PlaylistID) (map[PlaylistID]*PlaylistMeta, error)
	FetchPlaylistItems(id PlaylistID) ([]*FetchedItem, error)
	// FetchVideos は指定した動画の詳細と、それを投稿したチャンネルを返す。
	// 見られない動画（非公開、削除済み）は結果に含まれない。
	FetchVideos(ids []VideoID) ([]*Video, []*Channel, error)
	// FetchCategories はボットの地域の動画カテゴリを、その言語での名前で
	// 返す。
	FetchCategories() ([]*Category, error)
	// FetchLiveVideos は指定した動画のチャンネルと現在の再生数を、動画 ID を
	// キーにして返す。見られない動画は結果に含まれない。
	FetchLiveVideos(ids []VideoID) (map[VideoID]*LiveVideo, error)
}
