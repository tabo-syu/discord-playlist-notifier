package application

import (
	"errors"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// QueryService answers read-only questions about the subscribed playlists,
// for clients outside Discord. Only the stored contents are read, so no
// YouTube quota is used. Playlists no guild subscribes to are not shown.
type QueryService struct {
	subscriptions subscription.Repository
	library       *LibraryService
}

func NewQueryService(s subscription.Repository, l *LibraryService) *QueryService {
	return &QueryService{s, l}
}

// WatchedPlaylist is a playlist subscribed by at least one guild.
type WatchedPlaylist struct {
	ID    library.PlaylistID
	Title string
}

// PlaylistOverview is a watched playlist together with its statistics.
type PlaylistOverview struct {
	WatchedPlaylist
	Stats *library.Stats
}

// VideoSort is the order of FindVideos.
type VideoSort string

const (
	SortAddedDesc VideoSort = "added_desc"
	SortAddedAsc  VideoSort = "added_asc"
	SortViewsDesc VideoSort = "views_desc"
)

// VideoQuery narrows down the videos of the watched playlists. Zero values
// mean no condition.
type VideoQuery struct {
	PlaylistID library.PlaylistID
	// Case-insensitive part of the title
	Title string
	// Case-insensitive part of the channel name
	Channel   string
	ChannelID library.ChannelID
	// Case-insensitive part of one of the tags
	Tag        string
	CategoryID library.CategoryID
	// Length at least / at most. Videos whose length is not known yet are
	// left out when either is set.
	MinDuration time.Duration
	MaxDuration time.Duration
	// Added at or after
	AddedFrom time.Time
	// Added before
	AddedUntil time.Time
	// Include private and deleted videos, whose title may be unknown
	IncludeUnavailable bool
	Sort               VideoSort
	Offset             int
	Limit              int
}

// VideoPage is a page of FindVideos, with the number of all the matches.
type VideoPage struct {
	Total  int
	Videos []*library.PlaylistVideo
}

// Playlists returns the watched playlists, once per YouTube playlist, ordered
// by title. Returns an empty slice when there is none.
func (s *QueryService) Playlists() ([]*WatchedPlaylist, error) {
	subscriptions, err := s.subscriptions.FindAll()
	if errors.Is(err, domain.ErrDBRecordNotFound) {
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

// Overviews returns the watched playlists with their statistics.
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

// Playlist returns the watched playlist, or domain.ErrDBRecordNotFound.
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

	return nil, domain.ErrDBRecordNotFound
}

// scope returns the given playlist, or every watched playlist when id is empty.
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

// FindVideos returns the items of the watched playlists that match the query.
// A video in several playlists, or added twice, appears once per item.
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

	// Channels whose name matches, when the name is a condition
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

	// Contents are ordered by the time they were added
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

// Stats aggregates the stored contents of a watched playlist.
func (s *QueryService) Stats(id library.PlaylistID) (*library.Stats, error) {
	if _, err := s.Playlist(id); err != nil {
		return nil, err
	}

	return s.library.Stats(id)
}

// Wrapped aggregates the items added to a watched playlist in the year.
func (s *QueryService) Wrapped(id library.PlaylistID, year int) (*library.Wrapped, error) {
	if _, err := s.Playlist(id); err != nil {
		return nil, err
	}

	return s.library.Wrapped(id, year)
}

// History returns the states of a video in the watched playlists recorded at
// or after since, or domain.ErrDBRecordNotFound.
func (s *QueryService) History(id library.VideoID, since time.Time) (*library.VideoHistory, error) {
	ids, err := s.scope("")
	if err != nil {
		return nil, err
	}

	return s.library.History(ids, id, since)
}

// Channels returns the stored channels among the given IDs.
func (s *QueryService) Channels(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	return s.library.Channels(ids)
}

// channelsNamed returns the channels of the contents whose name contains
// name, ignoring case.
func (s *QueryService) channelsNamed(contents []*library.PlaylistVideo, name string) (map[library.ChannelID]bool, error) {
	var ids []library.ChannelID
	for _, c := range contents {
		if c.Video.ChannelYoutubeID != "" {
			ids = append(ids, c.Video.ChannelYoutubeID)
		}
	}
	stored, err := s.library.Channels(ids)
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

// RandomVideos picks up to n distinct available videos from a watched
// playlist, or from all of them when id is empty.
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
