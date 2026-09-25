package library

import "time"

// PlaylistMeta is the playlist-level information used to decide whether the
// items of a playlist need to be fetched again.
type PlaylistMeta struct {
	YoutubeID PlaylistID
	Title     string
	ItemCount int64
}

// FetchedItem is one playlist item as fetched from YouTube.
type FetchedItem struct {
	ItemID        ItemID
	VideoID       VideoID
	AddedAt       time.Time
	PrivacyStatus PrivacyStatus
}

// Sync is the result of comparing the stored items of a playlist with the
// ones fetched from YouTube.
type Sync struct {
	Added   []*PlaylistItem
	Removed []*PlaylistItem
	// Every video of the fetched items, including the ones seen for the first time
	Videos map[VideoID]*Video
	// Videos that became private or deleted
	Hidden []*Video
	// Watchable videos whose details have never been fetched
	NeedDetails []VideoID

	changed map[VideoID]*Video
}

// Reconcile compares the stored items of a playlist with the fetched ones.
// videos must hold the stored videos of the fetched items; the map is taken
// over by the returned Sync, which adds the videos seen for the first time.
//
// The privacy of a video is taken from its playlist items, which also
// report private and deleted videos, unlike the videos endpoint.
func Reconcile(playlistID PlaylistID, stored []*PlaylistItem, fetched []*FetchedItem, videos map[VideoID]*Video) *Sync {
	s := &Sync{Videos: videos, changed: map[VideoID]*Video{}}

	storedByID := map[ItemID]*PlaylistItem{}
	for _, item := range stored {
		storedByID[item.ItemID] = item
	}
	fetchedByID := map[ItemID]bool{}
	for _, f := range fetched {
		fetchedByID[f.ItemID] = true
		if _, ok := storedByID[f.ItemID]; !ok {
			s.Added = append(s.Added, &PlaylistItem{
				PlaylistYoutubeID: playlistID,
				ItemID:            f.ItemID,
				VideoYoutubeID:    f.VideoID,
				AddedAt:           f.AddedAt,
			})
		}
	}
	for _, item := range stored {
		if !fetchedByID[item.ItemID] {
			s.Removed = append(s.Removed, item)
		}
	}

	needDetails := map[VideoID]bool{}
	for _, f := range fetched {
		video, ok := videos[f.VideoID]
		if !ok {
			video = &Video{YoutubeID: f.VideoID}
			videos[f.VideoID] = video
		}
		changed, hidden := video.ChangePrivacy(f.PrivacyStatus)
		if hidden {
			s.Hidden = append(s.Hidden, video)
		}
		if changed {
			s.changed[f.VideoID] = video
		}
		if !video.HasDetails() && f.PrivacyStatus.Watchable() && !needDetails[f.VideoID] {
			needDetails[f.VideoID] = true
			s.NeedDetails = append(s.NeedDetails, f.VideoID)
		}
	}

	return s
}

// ApplyDetails copies the fetched details of NeedDetails to the videos.
func (s *Sync) ApplyDetails(details []*Video) {
	for _, d := range details {
		video := s.Videos[d.YoutubeID]
		video.ApplyDetails(d)
		video.InitMilestone()
		s.changed[d.YoutubeID] = video
	}
}

// ChangedVideos returns the videos that need to be saved.
func (s *Sync) ChangedVideos() []*Video {
	var videos []*Video
	for _, v := range s.changed {
		videos = append(videos, v)
	}

	return videos
}
