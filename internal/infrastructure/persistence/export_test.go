package persistence

// テスト（persistence_test）からだけ、レコード（DB の行の形）の型を使えるようにする。
// テストは DB の行をこれで準備し、論理削除した行なども含めて確かめる。
type (
	GuildRecord         = guildRecord
	SubscriptionRecord  = subscriptionRecord
	PlaylistItemRecord  = playlistItemRecord
	VideoRecord         = videoRecord
	ChannelRecord       = channelRecord
	CategoryRecord      = categoryRecord
	VideoSnapshotRecord = videoSnapshotRecord
)
