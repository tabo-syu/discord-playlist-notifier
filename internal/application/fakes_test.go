package application_test

import (
	"errors"
	"slices"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// このファイルはリポジトリと外部サービスを、メモリ上の偽物に差し替える。
// 本物の DB と同じく、読み書きのたびに値をコピーし、呼び出し側が持っている
// ポインタを書き換えても保存した値は変わらないようにする。

var errFake = errors.New("fake error")

func clone[T any](v *T) *T {
	c := *v
	return &c
}

// guildRepository は guild.Repository の偽物。
type guildRepository struct {
	guilds []*guild.Guild
}

func (r *guildRepository) Exists(id guild.DiscordID) (bool, error) {
	_, err := r.FindByDiscordID(id)
	return err == nil, nil
}

func (r *guildRepository) Add(g *guild.Guild) error {
	g.ID = uint(len(r.guilds) + 1)
	r.guilds = append(r.guilds, clone(g))
	return nil
}

func (r *guildRepository) FindByDiscordID(id guild.DiscordID) (*guild.Guild, error) {
	for _, g := range r.guilds {
		if g.DiscordID == id {
			return clone(g), nil
		}
	}
	return nil, domain.ErrDBRecordNotFound
}

func (r *guildRepository) Delete(g *guild.Guild) error {
	r.guilds = slices.DeleteFunc(r.guilds, func(stored *guild.Guild) bool { return stored.ID == g.ID })
	return nil
}

// subscriptionRepository は subscription.Repository の偽物。本物と同じく、
// 結果が空のときは domain.ErrDBRecordNotFound を返す。
type subscriptionRepository struct {
	guilds        *guildRepository
	subscriptions []*subscription.Subscription
	// Update に失敗させる通知設定のプレイリスト
	updateFails []library.PlaylistID
}

func (r *subscriptionRepository) Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error) {
	subscriptions, err := r.FindByGuild(guildID)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return false, nil
	}
	return slices.ContainsFunc(subscriptions, func(s *subscription.Subscription) bool { return s.YoutubeID == playlistID }), nil
}

func (r *subscriptionRepository) Add(s *subscription.Subscription) error {
	s.ID = uint(len(r.subscriptions) + 1)
	r.subscriptions = append(r.subscriptions, clone(s))
	return nil
}

// Update は本物（GORM の Save）と同じく、UpdatedAt を保存した時刻（now）で上書きする。
func (r *subscriptionRepository) Update(s *subscription.Subscription) error {
	if slices.Contains(r.updateFails, s.YoutubeID) {
		return errFake
	}
	s.UpdatedAt = now
	return r.replace(s)
}

func (r *subscriptionRepository) UpdatePickInterval(s *subscription.Subscription) error {
	for _, stored := range r.subscriptions {
		if stored.ID == s.ID {
			stored.PickInterval = s.PickInterval
			return nil
		}
	}
	return domain.ErrDBRecordNotFound
}

func (r *subscriptionRepository) replace(s *subscription.Subscription) error {
	for i, stored := range r.subscriptions {
		if stored.ID == s.ID {
			r.subscriptions[i] = clone(s)
			return nil
		}
	}
	return domain.ErrDBRecordNotFound
}

func (r *subscriptionRepository) FindAll() ([]*subscription.Subscription, error) {
	return r.find(func(*subscription.Subscription) bool { return true })
}

func (r *subscriptionRepository) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	g, err := r.guilds.FindByDiscordID(guildID)
	if err != nil {
		return nil, domain.ErrDBRecordNotFound
	}
	return r.find(func(s *subscription.Subscription) bool { return s.GuildID == g.ID })
}

func (r *subscriptionRepository) find(match func(*subscription.Subscription) bool) ([]*subscription.Subscription, error) {
	var found []*subscription.Subscription
	for _, s := range r.subscriptions {
		if match(s) {
			found = append(found, clone(s))
		}
	}
	if len(found) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}
	return found, nil
}

func (r *subscriptionRepository) DeleteAll(subscriptions []*subscription.Subscription) error {
	for _, s := range subscriptions {
		r.subscriptions = slices.DeleteFunc(r.subscriptions, func(stored *subscription.Subscription) bool { return stored.ID == s.ID })
	}
	return nil
}

// libraryRepository は library.Repository の偽物。使わないメソッドは空の結果を返す。
type libraryRepository struct {
	library.Repository

	items     []*library.PlaylistItem
	videos    []*library.Video
	snapshots []*library.VideoSnapshot
}

func (r *libraryRepository) FindItems(playlistIDs ...library.PlaylistID) ([]*library.PlaylistItem, error) {
	var found []*library.PlaylistItem
	for _, item := range r.items {
		if slices.Contains(playlistIDs, item.PlaylistYoutubeID) {
			found = append(found, clone(item))
		}
	}
	slices.SortStableFunc(found, func(a, b *library.PlaylistItem) int { return a.AddedAt.Compare(b.AddedAt) })
	return found, nil
}

func (r *libraryRepository) FindVideos(ids []library.VideoID) (map[library.VideoID]*library.Video, error) {
	found := map[library.VideoID]*library.Video{}
	for _, v := range r.videos {
		if slices.Contains(ids, v.YoutubeID) {
			found[v.YoutubeID] = clone(v)
		}
	}
	return found, nil
}

func (r *libraryRepository) FindVideosInPlaylists(playlistIDs ...library.PlaylistID) (map[library.VideoID]*library.Video, error) {
	items, _ := r.FindItems(playlistIDs...)
	var ids []library.VideoID
	for _, item := range items {
		ids = append(ids, item.VideoYoutubeID)
	}
	return r.FindVideos(ids)
}

func (r *libraryRepository) FindListedVideos() (map[library.VideoID]*library.Video, error) {
	var ids []library.VideoID
	for _, item := range r.items {
		ids = append(ids, item.VideoYoutubeID)
	}
	return r.FindVideos(ids)
}

func (r *libraryRepository) FindItemsByVideos(ids []library.VideoID) ([]*library.PlaylistItem, error) {
	var found []*library.PlaylistItem
	for _, item := range r.items {
		if slices.Contains(ids, item.VideoYoutubeID) {
			found = append(found, clone(item))
		}
	}
	return found, nil
}

func (r *libraryRepository) ApplyItems(playlistID library.PlaylistID, added []*library.PlaylistItem, removed []*library.PlaylistItem, videos []*library.Video) error {
	for _, item := range removed {
		r.items = slices.DeleteFunc(r.items, func(stored *library.PlaylistItem) bool { return stored.ItemID == item.ItemID })
	}
	for _, item := range added {
		r.items = append(r.items, clone(item))
	}
	return r.SaveVideos(videos)
}

func (r *libraryRepository) SaveVideos(videos []*library.Video) error {
	for _, v := range videos {
		i := slices.IndexFunc(r.videos, func(stored *library.Video) bool { return stored.YoutubeID == v.YoutubeID })
		if i < 0 {
			r.videos = append(r.videos, clone(v))
		} else {
			r.videos[i] = clone(v)
		}
	}
	return nil
}

func (r *libraryRepository) SaveChannels([]*library.Channel) error {
	return nil
}

func (r *libraryRepository) AddSnapshots(snapshots []*library.VideoSnapshot) error {
	for _, s := range snapshots {
		r.snapshots = append(r.snapshots, clone(s))
	}
	return nil
}

// youtube は library.YouTube の偽物。呼ばれた回数と、詳細を取得した動画を記録する。
type youtube struct {
	playlists []*youtubePlaylist
	// FetchPlaylistItems に失敗させるプレイリスト
	itemFails []library.PlaylistID
	// FetchVideos が返す動画の再生数。ないものは 0
	views map[library.VideoID]library.ViewCount

	itemFetches   int
	videoRequests [][]library.VideoID
}

// youtubePlaylist は YouTube 上のプレイリスト。
type youtubePlaylist struct {
	id    library.PlaylistID
	title string
	items []*library.FetchedItem
}

func (y *youtube) find(id library.PlaylistID) *youtubePlaylist {
	for _, p := range y.playlists {
		if p.id == id {
			return p
		}
	}
	return nil
}

func (y *youtube) FindPlaylist(id library.PlaylistID) (*library.PlaylistMeta, error) {
	p := y.find(id)
	if p == nil {
		return nil, domain.ErrYouTubePlaylistNotFound
	}
	return &library.PlaylistMeta{YoutubeID: p.id, Title: p.title}, nil
}

func (y *youtube) FetchPlaylistMetas(ids ...library.PlaylistID) (map[library.PlaylistID]*library.PlaylistMeta, error) {
	metas := map[library.PlaylistID]*library.PlaylistMeta{}
	for _, id := range ids {
		if p := y.find(id); p != nil {
			metas[id] = &library.PlaylistMeta{YoutubeID: p.id, Title: p.title, ItemCount: int64(len(p.items))}
		}
	}
	return metas, nil
}

func (y *youtube) FetchPlaylistItems(id library.PlaylistID) ([]*library.FetchedItem, error) {
	y.itemFetches++
	if slices.Contains(y.itemFails, id) {
		return nil, errFake
	}
	var items []*library.FetchedItem
	for _, item := range y.find(id).items {
		items = append(items, clone(item))
	}
	return items, nil
}

// FetchVideos は動画 ID をタイトルにした詳細を返す。
func (y *youtube) FetchVideos(ids []library.VideoID) ([]*library.Video, []*library.Channel, error) {
	y.videoRequests = append(y.videoRequests, ids)
	var videos []*library.Video
	for _, id := range ids {
		videos = append(videos, &library.Video{YoutubeID: id, Title: string(id), Views: y.views[id]})
	}
	return videos, nil, nil
}

func (y *youtube) FetchCategories() ([]*library.Category, error) {
	return nil, nil
}

func (y *youtube) FetchLiveVideos(ids []library.VideoID) (map[library.VideoID]*library.LiveVideo, error) {
	return map[library.VideoID]*library.LiveVideo{}, nil
}

// notifier は application.Notifier の偽物。送った通知を記録する。
type notifier struct {
	// true なら送信に失敗する
	fails bool

	newVideos  []*notification.NewVideos
	hidden     []*notification.HiddenVideo
	milestones []*notification.Milestone
}

func (n *notifier) NotifyNewVideos(news *notification.NewVideos, _ map[library.VideoID]*library.LiveVideo) error {
	if n.fails {
		return errFake
	}
	n.newVideos = append(n.newVideos, news)
	return nil
}

func (n *notifier) NotifyHiddenVideo(notice *notification.HiddenVideo) error {
	if n.fails {
		return errFake
	}
	n.hidden = append(n.hidden, notice)
	return nil
}

func (n *notifier) NotifyMilestone(notice *notification.Milestone, _ map[library.VideoID]*library.LiveVideo) error {
	if n.fails {
		return errFake
	}
	n.milestones = append(n.milestones, notice)
	return nil
}

func (n *notifier) PostWrapped(*subscription.Subscription, *library.Wrapped) error {
	return nil
}

func (n *notifier) PostPick(*subscription.Subscription, *library.PlaylistVideo, map[library.VideoID]*library.LiveVideo) error {
	return nil
}
