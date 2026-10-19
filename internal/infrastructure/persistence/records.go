package persistence

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"gorm.io/gorm"
)

// このファイルのレコードは、DB のテーブルの行の形をそのまま表す。ドメインのモデルは
// 保存の仕方を知らないので、リポジトリがレコードとの間で変換する。
//
// カラムの名前と型を変えないように、フィールドの名前は変換前のドメインのモデルと同じにし、
// 型は値オブジェクトではなく基本の型にしている。永続化するフィールドを足すときは、
// ドメインのモデル・レコード・変換の関数の 3 か所を直す。

type guildRecord struct {
	gorm.Model

	DiscordID string
}

func (guildRecord) TableName() string {
	return "guilds"
}

func toGuildRecord(g *guild.Guild) *guildRecord {
	return &guildRecord{
		ID:        g.ID,
		DiscordID: string(g.DiscordID),
	}
}

func (r *guildRecord) toDomain() *guild.Guild {
	return &guild.Guild{
		ID:        r.ID,
		DiscordID: guild.DiscordID(r.DiscordID),
	}
}

type subscriptionRecord struct {
	gorm.Model

	YoutubeID     string
	SendChannelID string
	Title         string
	PickInterval  string
	GuildID       uint
	NotifiedUntil time.Time
}

// テーブル名はモデルのリネーム前の名前のままにしている。
func (subscriptionRecord) TableName() string {
	return "playlists"
}

func toSubscriptionRecord(s *subscription.Subscription) *subscriptionRecord {
	return &subscriptionRecord{
		ID:            s.ID,
		YoutubeID:     string(s.YoutubeID),
		SendChannelID: string(s.SendChannelID),
		Title:         s.Title,
		PickInterval:  string(s.PickInterval),
		GuildID:       s.GuildID,
		NotifiedUntil: s.NotifiedUntil,
	}
}

func (r *subscriptionRecord) toDomain() *subscription.Subscription {
	return &subscription.Subscription{
		ID:            r.ID,
		YoutubeID:     library.PlaylistID(r.YoutubeID),
		SendChannelID: subscription.ChannelID(r.SendChannelID),
		Title:         r.Title,
		PickInterval:  subscription.PickInterval(r.PickInterval),
		GuildID:       r.GuildID,
		NotifiedUntil: r.NotifiedUntil,
	}
}

type playlistItemRecord struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PlaylistYoutubeID string `gorm:"index"`
	ItemID            string `gorm:"uniqueIndex"`
	VideoYoutubeID    string `gorm:"index"`
	AddedAt           time.Time
}

func (playlistItemRecord) TableName() string {
	return "playlist_items"
}

func toPlaylistItemRecord(i *library.PlaylistItem) *playlistItemRecord {
	return &playlistItemRecord{
		PlaylistYoutubeID: string(i.PlaylistYoutubeID),
		ItemID:            string(i.ItemID),
		VideoYoutubeID:    string(i.VideoYoutubeID),
		AddedAt:           i.AddedAt,
	}
}

func (r *playlistItemRecord) toDomain() *library.PlaylistItem {
	return &library.PlaylistItem{
		PlaylistYoutubeID: library.PlaylistID(r.PlaylistYoutubeID),
		ItemID:            library.ItemID(r.ItemID),
		VideoYoutubeID:    library.VideoID(r.VideoYoutubeID),
		AddedAt:           r.AddedAt,
	}
}

type videoRecord struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	YoutubeID        string `gorm:"uniqueIndex"`
	Title            string
	Views            uint64
	PrivacyStatus    string
	PublishedAt      time.Time
	ViewMilestone    *uint64
	Likes            uint64
	Comments         uint64
	ChannelYoutubeID string `gorm:"index"`
	Duration         time.Duration
	Tags             []string `gorm:"serializer:json"`
	CategoryID       string
}

func (videoRecord) TableName() string {
	return "videos"
}

func toVideoRecord(v *library.Video) *videoRecord {
	var milestone *uint64

	if v.ViewMilestone != nil {
		m := uint64(*v.ViewMilestone)
		milestone = &m
	}

	return &videoRecord{
		YoutubeID:        string(v.YoutubeID),
		Title:            v.Title,
		Views:            uint64(v.Views),
		PrivacyStatus:    string(v.PrivacyStatus),
		PublishedAt:      v.PublishedAt,
		ViewMilestone:    milestone,
		Likes:            v.Likes,
		Comments:         v.Comments,
		ChannelYoutubeID: string(v.ChannelYoutubeID),
		Duration:         v.Duration,
		Tags:             v.Tags,
		CategoryID:       string(v.CategoryID),
	}
}

func (r *videoRecord) toDomain() *library.Video {
	var milestone *library.ViewCount

	if r.ViewMilestone != nil {
		m := library.ViewCount(*r.ViewMilestone)
		milestone = &m
	}

	return &library.Video{
		YoutubeID:        library.VideoID(r.YoutubeID),
		Title:            r.Title,
		Views:            library.ViewCount(r.Views),
		PrivacyStatus:    library.PrivacyStatus(r.PrivacyStatus),
		PublishedAt:      r.PublishedAt,
		ViewMilestone:    milestone,
		Likes:            r.Likes,
		Comments:         r.Comments,
		ChannelYoutubeID: library.ChannelID(r.ChannelYoutubeID),
		Duration:         r.Duration,
		Tags:             r.Tags,
		CategoryID:       library.CategoryID(r.CategoryID),
	}
}

type channelRecord struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	YoutubeID string `gorm:"uniqueIndex"`
	Title     string
}

func (channelRecord) TableName() string {
	return "channels"
}

func toChannelRecord(c *library.Channel) *channelRecord {
	return &channelRecord{
		YoutubeID: string(c.YoutubeID),
		Title:     c.Title,
	}
}

func (r *channelRecord) toDomain() *library.Channel {
	return &library.Channel{
		YoutubeID: library.ChannelID(r.YoutubeID),
		Title:     r.Title,
	}
}

type categoryRecord struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	YoutubeID  string `gorm:"uniqueIndex"`
	Title      string
	Assignable bool
}

func (categoryRecord) TableName() string {
	return "categories"
}

func toCategoryRecord(c *library.Category) *categoryRecord {
	return &categoryRecord{
		YoutubeID:  string(c.YoutubeID),
		Title:      c.Title,
		Assignable: c.Assignable,
	}
}

func (r *categoryRecord) toDomain() *library.Category {
	return &library.Category{
		YoutubeID:  library.CategoryID(r.YoutubeID),
		Title:      r.Title,
		Assignable: r.Assignable,
	}
}

type videoSnapshotRecord struct {
	ID             uint      `gorm:"primarykey"`
	VideoYoutubeID string    `gorm:"index:idx_video_snapshots_video_recorded,priority:1"`
	RecordedAt     time.Time `gorm:"index:idx_video_snapshots_video_recorded,priority:2"`
	Title          string
	PrivacyStatus  string
	Views          uint64
	Likes          uint64
	Comments       uint64
}

func (videoSnapshotRecord) TableName() string {
	return "video_snapshots"
}

func toVideoSnapshotRecord(s *library.VideoSnapshot) *videoSnapshotRecord {
	return &videoSnapshotRecord{
		VideoYoutubeID: string(s.VideoYoutubeID),
		RecordedAt:     s.RecordedAt,
		Title:          s.Title,
		PrivacyStatus:  string(s.PrivacyStatus),
		Views:          uint64(s.Views),
		Likes:          s.Likes,
		Comments:       s.Comments,
	}
}

func (r *videoSnapshotRecord) toDomain() *library.VideoSnapshot {
	return &library.VideoSnapshot{
		VideoYoutubeID: library.VideoID(r.VideoYoutubeID),
		RecordedAt:     r.RecordedAt,
		Title:          r.Title,
		PrivacyStatus:  library.PrivacyStatus(r.PrivacyStatus),
		Views:          library.ViewCount(r.Views),
		Likes:          r.Likes,
		Comments:       r.Comments,
	}
}

// mapSlice はレコードとドメインのモデルのスライスを相互に変換する。
func mapSlice[T, U any](from []*T, f func(*T) *U) []*U {
	to := make([]*U, 0, len(from))
	for _, x := range from {
		to = append(to, f(x))
	}

	return to
}
