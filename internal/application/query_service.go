package application

import (
	"errors"
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

	title := strings.ToLower(q.Title)
	matched := []*library.PlaylistVideo{}
	for _, c := range contents {
		switch {
		case !q.IncludeUnavailable && !c.Video.Available():
		case title != "" && !strings.Contains(strings.ToLower(c.Video.Title), title):
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
