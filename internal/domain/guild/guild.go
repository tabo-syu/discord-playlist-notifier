// Package guild はボットが参加している Discord サーバーの集約。
package guild

// DiscordID は Discord がギルド（サーバー）に割り当てる ID。
type DiscordID string

// Guild はボットが参加している Discord サーバー。
type Guild struct {
	// 通知設定（subscription.Subscription.GuildID）が参照する識別子。保存したときに
	// リポジトリが振る。
	ID        uint
	DiscordID DiscordID
}

func New(id DiscordID) *Guild {
	return &Guild{DiscordID: id}
}
