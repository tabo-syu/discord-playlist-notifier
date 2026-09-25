package library

// PrivacyStatus is the privacy status of a YouTube video as seen through a playlist item.
type PrivacyStatus string

const (
	PrivacyPublic   PrivacyStatus = "public"
	PrivacyUnlisted PrivacyStatus = "unlisted"
	PrivacyPrivate  PrivacyStatus = "private"
	PrivacyDeleted  PrivacyStatus = "deleted"
)

// Watchable reports whether videos with the status can be watched.
func (s PrivacyStatus) Watchable() bool {
	return s == PrivacyPublic || s == PrivacyUnlisted
}

// Hidden reports whether videos with the status became unavailable.
func (s PrivacyStatus) Hidden() bool {
	return s == PrivacyPrivate || s == PrivacyDeleted
}
