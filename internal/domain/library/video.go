package library

import "time"

// Video は通知対象のプレイリストに含まれる動画の、わかっている最新の状態を
// 動画ごとに 1 つ持つ。詳細（タイトル、再生数など）は動画が非公開・削除に
// なっても表示できるように残す。
type Video struct {
	YoutubeID VideoID
	// 詳細（タイトルから CategoryID まで）を一度でも取得したか。ApplyDetails で
	// 立て、その後は下ろさない。最初から非公開・削除の動画は詳細を取得できないので、
	// 偽のまま残る。
	DetailsFetched bool
	Title          string
	Views          ViewCount
	PrivacyStatus  PrivacyStatus
	// 動画自体が投稿された時刻
	PublishedAt time.Time
	// すでに達した再生数の最も高い節目。既存の再生数を通知しないように、
	// 初めて再生数がわかるまでは nil。
	ViewMilestone *ViewCount
	// 所有者が非表示にしたか無効にしたときはゼロ
	Likes    uint64
	Comments uint64
	// 動画を投稿したチャンネル。詳細を取得するまでは空
	ChannelYoutubeID ChannelID
	// まだ取得していないとき、または動画がライブ配信中・配信予定のときはゼロ
	Duration time.Duration
	// 投稿者が設定したタグ。空のことが多い
	Tags       []string
	CategoryID CategoryID
}

// HasDetails は動画の詳細を一度でも取得したかを返す。
func (v *Video) HasDetails() bool {
	return v.DetailsFetched
}

// Available は動画が視聴でき、ユーザーに見せてよいかを返す。
func (v *Video) Available() bool {
	return v.HasDetails() && v.PrivacyStatus.Watchable()
}

// Thumbnail は動画のサムネイルの URL を返す。YouTube は動画 ID からこれを作る。
func (v *Video) Thumbnail() string {
	return "https://i.ytimg.com/vi/" + string(v.YoutubeID) + "/hqdefault.jpg"
}

// ChangePrivacy は公開状態を設定し、変わったかどうかを返す。その変化で動画が
// 見られなくなったときは VideoHidden も返す（それ以外は nil）。初めて見る動画は
// 見られなくなったとはみなさない。
func (v *Video) ChangePrivacy(status PrivacyStatus) (changed bool, hidden *VideoHidden) {
	if v.PrivacyStatus == status {
		return false, nil
	}

	if v.Available() && status.Hidden() {
		hidden = &VideoHidden{Video: v}
	}

	v.PrivacyStatus = status

	return true, hidden
}

// ApplyDetails は videos のエンドポイントから取得した詳細をコピーし、詳細を
// 取得済みにする。公開状態はプレイリストのアイテムが持つものなので変えない。
func (v *Video) ApplyDetails(details *Video) {
	v.DetailsFetched = true
	v.Title = details.Title
	v.Views = details.Views
	v.Likes = details.Likes
	v.Comments = details.Comments
	v.PublishedAt = details.PublishedAt
	v.ChannelYoutubeID = details.ChannelYoutubeID
	v.Duration = details.Duration
	v.Tags = details.Tags
	v.CategoryID = details.CategoryID
}

// InitMilestone は初めて再生数がわかったときに、その再生数の節目を記録して
// 通知されないようにする。
func (v *Video) InitMilestone() {
	if v.ViewMilestone == nil {
		reached := v.Views.Milestone()
		v.ViewMilestone = &reached
	}
}

// UpdateMilestone は現在の再生数の節目を記録し、新しい節目に達したときは
// MilestoneReached を返す（それ以外は nil）。
func (v *Video) UpdateMilestone() *MilestoneReached {
	milestone := v.Views.Milestone()
	if v.ViewMilestone == nil {
		v.ViewMilestone = &milestone

		return nil
	}

	if milestone > *v.ViewMilestone {
		v.ViewMilestone = &milestone

		return &MilestoneReached{Video: v, Milestone: milestone}
	}

	return nil
}

// LiveVideo は動画を表示する直前に YouTube から取得する情報。チャンネルは
// 保存せず、再生数は保存済みのものより新しい。
type LiveVideo struct {
	ChannelName string
	ChannelIcon string
	Views       ViewCount
}
