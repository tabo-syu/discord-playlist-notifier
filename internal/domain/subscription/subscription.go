// Package subscription is the aggregate of the per-guild settings to be
// notified of a YouTube playlist.
package subscription

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
)

// ChannelID is the ID of the Discord text channel notifications are sent to.
type ChannelID string

// Subscription is the setting of a guild to be notified of a YouTube
// playlist. The contents of the playlist itself belong to the library.
//
// UpdatedAt is the base time of new video notifications: videos added to the
// playlist at or after it are new (see IsNew).
type Subscription struct {
	gorm.Model
	YoutubeID     library.PlaylistID
	SendChannelID ChannelID
	Title         string
	// How often to post a random video from the playlist
	PickInterval PickInterval
	// The guild.Guild the subscription belongs to. Other aggregates are
	// referenced by ID only.
	GuildID uint
}

// The table keeps the name the model had before it was renamed.
func (Subscription) TableName() string {
	return "playlists"
}

func New(g *guild.Guild, channelID ChannelID, youtubeID library.PlaylistID, title string) *Subscription {
	return &Subscription{
		YoutubeID:     youtubeID,
		SendChannelID: channelID,
		Title:         title,
		GuildID:       g.ID,
	}
}

// IsNew reports whether the item was added after the last notification.
// Uses !Before instead of After to include items added at exactly the same time.
func (s *Subscription) IsNew(item *library.PlaylistItem) bool {
	return !item.AddedAt.Before(s.UpdatedAt)
}

// MarkNotified moves the base time of new video notifications. Moving it back
// makes the videos added since then notified again.
func (s *Subscription) MarkNotified(at time.Time) {
	s.UpdatedAt = at
}

// Rename follows the title of the playlist on YouTube.
func (s *Subscription) Rename(title string) {
	s.Title = title
}

func (s *Subscription) ChangePickInterval(interval PickInterval) {
	s.PickInterval = interval
}
