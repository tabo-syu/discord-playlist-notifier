package fake

import (
	"errors"
	"slices"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// GuildRepository は guild.Repository の偽物。
type GuildRepository struct {
	Guilds []*guild.Guild
}

func (r *GuildRepository) Exists(id guild.DiscordID) (bool, error) {
	_, err := r.FindByDiscordID(id)
	return err == nil, nil
}

func (r *GuildRepository) Add(g *guild.Guild) error {
	g.ID = uint(len(r.Guilds) + 1)
	r.Guilds = append(r.Guilds, clone(g))
	return nil
}

func (r *GuildRepository) FindByDiscordID(id guild.DiscordID) (*guild.Guild, error) {
	for _, g := range r.Guilds {
		if g.DiscordID == id {
			return clone(g), nil
		}
	}
	return nil, domain.ErrDBRecordNotFound
}

func (r *GuildRepository) Delete(g *guild.Guild) error {
	r.Guilds = slices.DeleteFunc(r.Guilds, func(stored *guild.Guild) bool { return stored.ID == g.ID })
	return nil
}

// SubscriptionRepository は subscription.Repository の偽物。本物と同じく、
// 結果が空のときは domain.ErrDBRecordNotFound を返す。
type SubscriptionRepository struct {
	Guilds        *GuildRepository
	Subscriptions []*subscription.Subscription
	// Update が UpdatedAt に書き込む時刻。本物（GORM の Save）は保存した時刻で上書きする
	SavedAt time.Time
	// Update に失敗させる通知設定のプレイリスト
	UpdateFails []library.PlaylistID
}

func (r *SubscriptionRepository) Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error) {
	subscriptions, err := r.FindByGuild(guildID)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return false, nil
	}
	return slices.ContainsFunc(subscriptions, func(s *subscription.Subscription) bool { return s.YoutubeID == playlistID }), nil
}

func (r *SubscriptionRepository) Add(s *subscription.Subscription) error {
	s.ID = uint(len(r.Subscriptions) + 1)
	r.Subscriptions = append(r.Subscriptions, clone(s))
	return nil
}

func (r *SubscriptionRepository) Update(s *subscription.Subscription) error {
	if slices.Contains(r.UpdateFails, s.YoutubeID) {
		return Err
	}
	s.UpdatedAt = r.SavedAt
	for i, stored := range r.Subscriptions {
		if stored.ID == s.ID {
			r.Subscriptions[i] = clone(s)
			return nil
		}
	}
	return domain.ErrDBRecordNotFound
}

func (r *SubscriptionRepository) UpdatePickInterval(s *subscription.Subscription) error {
	for _, stored := range r.Subscriptions {
		if stored.ID == s.ID {
			stored.PickInterval = s.PickInterval
			return nil
		}
	}
	return domain.ErrDBRecordNotFound
}

func (r *SubscriptionRepository) FindAll() ([]*subscription.Subscription, error) {
	return r.find(func(*subscription.Subscription) bool { return true })
}

func (r *SubscriptionRepository) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	g, err := r.Guilds.FindByDiscordID(guildID)
	if err != nil {
		return nil, domain.ErrDBRecordNotFound
	}
	return r.find(func(s *subscription.Subscription) bool { return s.GuildID == g.ID })
}

func (r *SubscriptionRepository) find(match func(*subscription.Subscription) bool) ([]*subscription.Subscription, error) {
	var found []*subscription.Subscription
	for _, s := range r.Subscriptions {
		if match(s) {
			found = append(found, clone(s))
		}
	}
	if len(found) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}
	return found, nil
}

func (r *SubscriptionRepository) DeleteAll(subscriptions []*subscription.Subscription) error {
	for _, s := range subscriptions {
		r.Subscriptions = slices.DeleteFunc(r.Subscriptions, func(stored *subscription.Subscription) bool { return stored.ID == s.ID })
	}
	return nil
}

// LibraryRepository は library.Repository の偽物。
type LibraryRepository struct {
	Items      []*library.PlaylistItem
	Videos     []*library.Video
	Snapshots  []*library.VideoSnapshot
	Channels   []*library.Channel
	Categories []*library.Category
}

func (r *LibraryRepository) FindItems(playlistIDs ...library.PlaylistID) ([]*library.PlaylistItem, error) {
	var found []*library.PlaylistItem
	for _, item := range r.Items {
		if slices.Contains(playlistIDs, item.PlaylistYoutubeID) {
			found = append(found, clone(item))
		}
	}
	slices.SortStableFunc(found, func(a, b *library.PlaylistItem) int { return a.AddedAt.Compare(b.AddedAt) })
	return found, nil
}

func (r *LibraryRepository) FindVideos(ids []library.VideoID) (map[library.VideoID]*library.Video, error) {
	found := map[library.VideoID]*library.Video{}
	for _, v := range r.Videos {
		if slices.Contains(ids, v.YoutubeID) {
			found[v.YoutubeID] = clone(v)
		}
	}
	return found, nil
}

func (r *LibraryRepository) FindVideosInPlaylists(playlistIDs ...library.PlaylistID) (map[library.VideoID]*library.Video, error) {
	items, _ := r.FindItems(playlistIDs...)
	var ids []library.VideoID
	for _, item := range items {
		ids = append(ids, item.VideoYoutubeID)
	}
	return r.FindVideos(ids)
}

func (r *LibraryRepository) FindListedVideos() (map[library.VideoID]*library.Video, error) {
	var ids []library.VideoID
	for _, item := range r.Items {
		ids = append(ids, item.VideoYoutubeID)
	}
	return r.FindVideos(ids)
}

func (r *LibraryRepository) FindItemsByVideos(ids []library.VideoID) ([]*library.PlaylistItem, error) {
	var found []*library.PlaylistItem
	for _, item := range r.Items {
		if slices.Contains(ids, item.VideoYoutubeID) {
			found = append(found, clone(item))
		}
	}
	return found, nil
}

func (r *LibraryRepository) ApplyItems(playlistID library.PlaylistID, added []*library.PlaylistItem, removed []*library.PlaylistItem, videos []*library.Video) error {
	for _, item := range removed {
		r.Items = slices.DeleteFunc(r.Items, func(stored *library.PlaylistItem) bool { return stored.ItemID == item.ItemID })
	}
	for _, item := range added {
		r.Items = append(r.Items, clone(item))
	}
	return r.SaveVideos(videos)
}

func (r *LibraryRepository) SaveVideos(videos []*library.Video) error {
	for _, v := range videos {
		i := slices.IndexFunc(r.Videos, func(stored *library.Video) bool { return stored.YoutubeID == v.YoutubeID })
		if i < 0 {
			r.Videos = append(r.Videos, clone(v))
		} else {
			r.Videos[i] = clone(v)
		}
	}
	return nil
}

func (r *LibraryRepository) SaveChannels(channels []*library.Channel) error {
	for _, c := range channels {
		r.Channels = slices.DeleteFunc(r.Channels, func(stored *library.Channel) bool { return stored.YoutubeID == c.YoutubeID })
		r.Channels = append(r.Channels, clone(c))
	}
	return nil
}

func (r *LibraryRepository) FindChannels(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	found := map[library.ChannelID]*library.Channel{}
	for _, c := range r.Channels {
		if slices.Contains(ids, c.YoutubeID) {
			found[c.YoutubeID] = clone(c)
		}
	}
	return found, nil
}

func (r *LibraryRepository) SaveCategories(categories []*library.Category) error {
	for _, c := range categories {
		r.Categories = slices.DeleteFunc(r.Categories, func(stored *library.Category) bool { return stored.YoutubeID == c.YoutubeID })
		r.Categories = append(r.Categories, clone(c))
	}
	return nil
}

func (r *LibraryRepository) FindCategories() (map[library.CategoryID]*library.Category, error) {
	found := map[library.CategoryID]*library.Category{}
	for _, c := range r.Categories {
		found[c.YoutubeID] = clone(c)
	}
	return found, nil
}

func (r *LibraryRepository) AddSnapshots(snapshots []*library.VideoSnapshot) error {
	for _, s := range snapshots {
		r.Snapshots = append(r.Snapshots, clone(s))
	}
	return nil
}

func (r *LibraryRepository) FindSnapshots(id library.VideoID, since time.Time) ([]*library.VideoSnapshot, error) {
	var found []*library.VideoSnapshot
	for _, s := range r.Snapshots {
		if s.VideoYoutubeID == id && !s.RecordedAt.Before(since) {
			found = append(found, clone(s))
		}
	}
	slices.SortStableFunc(found, func(a, b *library.VideoSnapshot) int { return a.RecordedAt.Compare(b.RecordedAt) })
	return found, nil
}

func (r *LibraryRepository) FindEarliestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	found := map[library.VideoID]*library.VideoSnapshot{}
	for _, id := range ids {
		if snapshots, _ := r.FindSnapshots(id, since); len(snapshots) > 0 {
			found[id] = snapshots[0]
		}
	}
	return found, nil
}

func (r *LibraryRepository) FindLatestSnapshots(ids []library.VideoID, since time.Time) (map[library.VideoID]*library.VideoSnapshot, error) {
	found := map[library.VideoID]*library.VideoSnapshot{}
	for _, id := range ids {
		if snapshots, _ := r.FindSnapshots(id, since); len(snapshots) > 0 {
			found[id] = snapshots[len(snapshots)-1]
		}
	}
	return found, nil
}
