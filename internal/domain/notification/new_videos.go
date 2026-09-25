// Package notification decides what is notified to which channel, from the
// subscriptions and the contents of their playlists.
package notification

import (
	"sort"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// NewVideos are the videos added to a subscribed playlist since its last notification.
type NewVideos struct {
	Subscription *subscription.Subscription
	// The current title of the playlist on YouTube
	PlaylistTitle string
	// Ordered by the time they were added
	Videos []*library.PlaylistVideo
}

// FindNewVideos picks, for each subscription, the available videos added to
// the playlist since its last notification. Subscriptions whose playlist
// failed to sync or was deleted are left out.
func FindNewVideos(subscriptions []*subscription.Subscription, snapshots map[library.PlaylistID]*library.Snapshot) []*NewVideos {
	var updates []*NewVideos
	for _, sub := range subscriptions {
		latest, ok := snapshots[sub.YoutubeID]
		if !ok || latest.Deleted {
			continue
		}

		var added []*library.PlaylistVideo
		for _, item := range latest.Items {
			if !sub.IsNew(item) {
				continue
			}
			video, ok := latest.Videos[item.VideoYoutubeID]
			if !ok || !video.Available() {
				continue
			}
			added = append(added, &library.PlaylistVideo{PlaylistID: sub.YoutubeID, Item: item, Video: video})
		}
		if len(added) != 0 {
			// Notify in the order the videos were added, regardless of the playlist order
			sort.SliceStable(added, func(i, j int) bool {
				return added[i].Item.AddedAt.Before(added[j].Item.AddedAt)
			})
			updates = append(updates, &NewVideos{Subscription: sub, PlaylistTitle: latest.Title, Videos: added})
		}
	}

	return updates
}
