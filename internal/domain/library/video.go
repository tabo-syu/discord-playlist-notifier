package library

import "time"

// Video holds the latest known state of a video that appears in a watched
// playlist, once per video. Details (title, views, ...) are kept after the
// video becomes private or deleted, so they can still be shown.
type Video struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	YoutubeID     VideoID `gorm:"uniqueIndex"`
	Title         string
	Views         ViewCount
	PrivacyStatus PrivacyStatus
	// When the video itself was published
	PublishedAt time.Time
	// The highest view milestone already reached. Nil until the first time
	// the views are known, so that existing views are not notified.
	ViewMilestone *ViewCount
}

func (Video) TableName() string {
	return "videos"
}

// HasDetails reports whether the video details have been fetched at least once.
func (v *Video) HasDetails() bool {
	return v.Title != ""
}

// Available reports whether the video can be watched and shown to users.
func (v *Video) Available() bool {
	return v.HasDetails() && v.PrivacyStatus.Watchable()
}

// Thumbnail returns the URL of the video thumbnail, which YouTube derives from the video ID.
func (v *Video) Thumbnail() string {
	return "https://i.ytimg.com/vi/" + string(v.YoutubeID) + "/hqdefault.jpg"
}

// ChangePrivacy sets the privacy status, and reports whether it changed and
// whether the video was hidden by the change. Videos seen for the first time
// are not reported as hidden.
func (v *Video) ChangePrivacy(status PrivacyStatus) (changed bool, hidden bool) {
	if v.PrivacyStatus == status {
		return false, false
	}
	hidden = v.Available() && status.Hidden()
	v.PrivacyStatus = status

	return true, hidden
}

// ApplyDetails copies the details fetched from the videos endpoint. The
// privacy status is left untouched because it is owned by the playlist items.
func (v *Video) ApplyDetails(details *Video) {
	v.Title = details.Title
	v.Views = details.Views
	v.PublishedAt = details.PublishedAt
}

// InitMilestone records the milestone of the current views the first time the
// views are known, so that they are not notified.
func (v *Video) InitMilestone() {
	if v.ViewMilestone == nil {
		reached := v.Views.Milestone()
		v.ViewMilestone = &reached
	}
}

// UpdateMilestone records the milestone of the current views, and reports
// whether the video reached a new one.
func (v *Video) UpdateMilestone() bool {
	milestone := v.Views.Milestone()
	if v.ViewMilestone == nil {
		v.ViewMilestone = &milestone
		return false
	}
	if milestone > *v.ViewMilestone {
		v.ViewMilestone = &milestone
		return true
	}

	return false
}

// LiveVideo is information fetched from YouTube right before a video is
// shown. The channel is not stored, and the views are fresher than the stored ones.
type LiveVideo struct {
	ChannelName string
	ChannelIcon string
	Views       ViewCount
}
