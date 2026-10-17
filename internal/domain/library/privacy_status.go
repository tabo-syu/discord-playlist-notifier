package library

// PrivacyStatus はプレイリストのアイテムから見た YouTube の動画の公開状態。
type PrivacyStatus string

const (
	PrivacyPublic   PrivacyStatus = "public"
	PrivacyUnlisted PrivacyStatus = "unlisted"
	PrivacyPrivate  PrivacyStatus = "private"
	PrivacyDeleted  PrivacyStatus = "deleted"
)

// Watchable はその状態の動画が視聴できるかを返す。
func (s PrivacyStatus) Watchable() bool {
	return s == PrivacyPublic || s == PrivacyUnlisted
}

// Hidden はその状態の動画が見られなくなったかを返す。
func (s PrivacyStatus) Hidden() bool {
	return s == PrivacyPrivate || s == PrivacyDeleted
}
