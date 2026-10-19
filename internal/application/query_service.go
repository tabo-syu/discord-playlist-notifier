package application

import (
	"errors"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// QueryService は Discord の外のクライアント向けに、通知登録されたプレイリストに
// 関する読み取り専用の問い合わせに答える。保存済みの中身だけを読むので
// YouTube のクォータは使わない。どのギルドも登録していないプレイリストは見せない。
type QueryService struct {
	subscriptions subscription.Repository
	library       *LibraryService
	categories    library.CategoryRepository
	channels      library.ChannelRepository
}

func NewQueryService(s subscription.Repository, l *LibraryService, cat library.CategoryRepository, c library.ChannelRepository) *QueryService {
	return &QueryService{s, l, cat, c}
}

// WatchedPlaylist は 1 つ以上のギルドが通知登録しているプレイリスト。
type WatchedPlaylist struct {
	ID    library.PlaylistID
	Title string
}

// PlaylistOverview は通知対象のプレイリストとその統計。
type PlaylistOverview struct {
	WatchedPlaylist

	Stats *library.Stats
}

// VideoSort は FindVideos の並び順。
type VideoSort string

const (
	SortAddedDesc VideoSort = "added_desc"
	SortAddedAsc  VideoSort = "added_asc"
	SortViewsDesc VideoSort = "views_desc"
)

// VideoQuery は通知対象のプレイリストの動画を絞り込む。ゼロ値は
// 条件なしを意味する。
type VideoQuery struct {
	PlaylistID library.PlaylistID
	// タイトルの一部（大文字・小文字を区別しない）
	Title string
	// チャンネル名の一部（大文字・小文字を区別しない）
	Channel   string
	ChannelID library.ChannelID
	// いずれかのタグの一部（大文字・小文字を区別しない）
	Tag        string
	CategoryID library.CategoryID
	// 長さの下限・上限。どちらかを指定すると、長さがまだわからない動画は
	// 除く。
	MinDuration time.Duration
	MaxDuration time.Duration
	// この時刻以降に追加された
	AddedFrom time.Time
	// この時刻より前に追加された
	AddedUntil time.Time
	// タイトルがわからないこともある非公開・削除済みの動画も含める
	IncludeUnavailable bool
	Sort               VideoSort
	Offset             int
	Limit              int
}

// VideoPage は FindVideos の 1 ページ分で、一致した全件の数を持つ。
type VideoPage struct {
	Total  int
	Videos []*library.PlaylistVideo
}

// Playlists は通知対象のプレイリストを YouTube のプレイリストごとに 1 つずつ、
// タイトル順に返す。1 つもなければ空のスライスを返す。
func (s *QueryService) Playlists() ([]*WatchedPlaylist, error) {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, subscription.ErrNotSubscribed) {
		return []*WatchedPlaylist{}, nil
	}

	if err != nil {
		return nil, err
	}

	seen := map[library.PlaylistID]bool{}
	playlists := []*WatchedPlaylist{}

	for _, sub := range subscriptions {
		if seen[sub.YoutubeID] {
			continue
		}

		seen[sub.YoutubeID] = true
		playlists = append(playlists, &WatchedPlaylist{ID: sub.YoutubeID, Title: sub.Title})
	}

	sort.SliceStable(playlists, func(i, j int) bool {
		return playlists[i].Title < playlists[j].Title
	})

	return playlists, nil
}

// Overviews は通知対象のプレイリストを統計と一緒に返す。
func (s *QueryService) Overviews() ([]*PlaylistOverview, error) {
	playlists, err := s.Playlists()
	if err != nil {
		return nil, err
	}

	overviews := []*PlaylistOverview{}

	for _, p := range playlists {
		stats, err := s.library.Stats(p.ID)
		if err != nil {
			return nil, err
		}

		overviews = append(overviews, &PlaylistOverview{WatchedPlaylist: *p, Stats: stats})
	}

	return overviews, nil
}

// Playlist は通知対象のプレイリストを返す。なければ subscription.ErrNotSubscribed を返す。
func (s *QueryService) Playlist(id library.PlaylistID) (*WatchedPlaylist, error) {
	playlists, err := s.Playlists()
	if err != nil {
		return nil, err
	}

	for _, p := range playlists {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, subscription.ErrNotSubscribed
}

// FindVideos は通知対象のプレイリストのアイテムのうち、条件に一致するものを返す。
// 複数のプレイリストにある動画や 2 回追加された動画は、アイテムごとに現れる。
func (s *QueryService) FindVideos(q VideoQuery) (*VideoPage, error) {
	ids, err := s.scope(q.PlaylistID)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return &VideoPage{Videos: []*library.PlaylistVideo{}}, nil
	}

	contents, err := s.library.Contents(ids)
	if err != nil {
		return nil, err
	}

	// チャンネル名が条件のときの、名前が一致するチャンネル
	var channels map[library.ChannelID]bool
	if q.Channel != "" {
		if channels, err = s.channelsNamed(contents, q.Channel); err != nil {
			return nil, err
		}
	}

	title, tag := strings.ToLower(q.Title), strings.ToLower(q.Tag)
	byDuration := q.MinDuration > 0 || q.MaxDuration > 0
	matched := []*library.PlaylistVideo{}

	for _, c := range contents {
		v := c.Video
		switch {
		case !q.IncludeUnavailable && !v.Available():
		case title != "" && !strings.Contains(strings.ToLower(v.Title), title):
		case channels != nil && !channels[v.ChannelYoutubeID]:
		case q.ChannelID != "" && v.ChannelYoutubeID != q.ChannelID:
		case tag != "" && !slices.ContainsFunc(v.Tags, func(t string) bool { return strings.Contains(strings.ToLower(t), tag) }):
		case q.CategoryID != "" && v.CategoryID != q.CategoryID:
		case byDuration && v.Duration == 0:
		case q.MinDuration > 0 && v.Duration < q.MinDuration:
		case q.MaxDuration > 0 && v.Duration > q.MaxDuration:
		case !q.AddedFrom.IsZero() && c.Item.AddedAt.Before(q.AddedFrom):
		case !q.AddedUntil.IsZero() && !c.Item.AddedAt.Before(q.AddedUntil):
		default:
			matched = append(matched, c)
		}
	}

	// 中身は追加された時刻の順に並んでいる
	switch q.Sort {
	case SortAddedAsc:
	case SortViewsDesc:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].Video.Views > matched[j].Video.Views
		})
	default:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].Item.AddedAt.After(matched[j].Item.AddedAt)
		})
	}

	page := &VideoPage{Total: len(matched), Videos: []*library.PlaylistVideo{}}
	if q.Offset < len(matched) {
		end := len(matched)
		if q.Limit > 0 {
			end = min(q.Offset+q.Limit, len(matched))
		}

		page.Videos = matched[q.Offset:end]
	}

	return page, nil
}

// Stats は通知対象のプレイリストの保存済みの中身を集計する。
func (s *QueryService) Stats(id library.PlaylistID) (*library.Stats, error) {
	if _, err := s.Playlist(id); err != nil {
		return nil, err
	}

	return s.library.Stats(id)
}

// Wrapped はその年に通知対象のプレイリストへ追加されたアイテムを集計する。
func (s *QueryService) Wrapped(id library.PlaylistID, year int) (*library.Wrapped, error) {
	if _, err := s.Playlist(id); err != nil {
		return nil, err
	}

	return s.library.Wrapped(id, year)
}

// History は通知対象のプレイリストにある動画の、since 以降に記録された状態を
// 返す。なければ library.ErrVideoNotFound を返す。
func (s *QueryService) History(id library.VideoID, since time.Time) (*library.VideoHistory, error) {
	ids, err := s.scope("")
	if err != nil {
		return nil, err
	}

	return s.library.History(ids, id, since)
}

// Categories は保存済みの全ての動画カテゴリを返す。
func (s *QueryService) Categories() (map[library.CategoryID]*library.Category, error) {
	return s.categories.FindAll()
}

// Channels は指定した ID のうち保存済みのチャンネルを返す。
func (s *QueryService) Channels(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	return s.channels.Find(ids)
}

// RandomVideos は通知対象のプレイリスト（id が空なら全て）から、
// 重複のない視聴可能な動画を最大 n 件選ぶ。
func (s *QueryService) RandomVideos(id library.PlaylistID, n int) ([]*library.PlaylistVideo, error) {
	ids, err := s.scope(id)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*library.PlaylistVideo{}, nil
	}

	return s.library.RandomVideos(ids, n)
}

// scope は指定したプレイリストを返す。id が空なら通知対象の全プレイリストを返す。
func (s *QueryService) scope(id library.PlaylistID) ([]library.PlaylistID, error) {
	if id != "" {
		if _, err := s.Playlist(id); err != nil {
			return nil, err
		}

		return []library.PlaylistID{id}, nil
	}

	playlists, err := s.Playlists()
	if err != nil {
		return nil, err
	}

	var ids []library.PlaylistID
	for _, p := range playlists {
		ids = append(ids, p.ID)
	}

	return ids, nil
}

// channelsNamed は中身の動画のチャンネルのうち、名前に name を含むものを
// 返す。大文字・小文字は区別しない。
func (s *QueryService) channelsNamed(contents []*library.PlaylistVideo, name string) (map[library.ChannelID]bool, error) {
	var ids []library.ChannelID

	for _, c := range contents {
		if c.Video.ChannelYoutubeID != "" {
			ids = append(ids, c.Video.ChannelYoutubeID)
		}
	}

	stored, err := s.channels.Find(ids)
	if err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	matched := map[library.ChannelID]bool{}

	for id, ch := range stored {
		if strings.Contains(strings.ToLower(ch.Title), name) {
			matched[id] = true
		}
	}

	return matched, nil
}
