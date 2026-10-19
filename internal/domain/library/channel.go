package library

// ChannelID は YouTube がチャンネルに割り当てる ID。
type ChannelID string

// Channel は通知対象のプレイリストにある動画を投稿した YouTube のチャンネル。
// 動画は ID で参照する。アイコンは保存せず、通知を送る直前に取得する。
type Channel struct {
	YoutubeID ChannelID
	// そのチャンネルの動画のいずれかを最後に更新したときの名前
	Title string
}
