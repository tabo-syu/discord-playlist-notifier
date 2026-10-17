package application

import (
	"log"
	"slices"
	"sync"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Even when the item count of a playlist has not changed, its items are
// fetched again after this interval, to catch additions that happened
// together with removals and privacy changes of videos.
const FULL_SYNC_INTERVAL = time.Hour

type syncState struct {
	itemCount  int64
	fullSyncAt time.Time
}

// LibraryService keeps the contents of the watched playlists in the database.
type LibraryService struct {
	youtube library.YouTube
	library library.Repository
	// Time zone used to group dates by month or year
	location *time.Location

	// Keeps overlapping runs from writing the same videos
	mu    sync.Mutex
	state map[library.PlaylistID]syncState
	now   func() time.Time
}

func NewLibraryService(y library.YouTube, l library.Repository, loc *time.Location) *LibraryService {
	return &LibraryService{youtube: y, library: l, location: loc, state: map[library.PlaylistID]syncState{}, now: time.Now}
}

// Sync brings the stored contents of the given playlists up to date and
// returns them. Playlists that failed to sync are logged and left out.
func (s *LibraryService) Sync(playlistIDs []library.PlaylistID) (map[library.PlaylistID]*library.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := unique(playlistIDs)
	metas, err := s.youtube.FetchPlaylistMetas(ids...)
	if err != nil {
		return nil, err
	}

	snapshots := map[library.PlaylistID]*library.Snapshot{}
	for _, id := range ids {
		meta, ok := metas[id]
		if !ok {
			snapshots[id] = &library.Snapshot{YoutubeID: id, Deleted: true}
			continue
		}

		snapshot, err := s.syncPlaylist(meta)
		if err != nil {
			log.Println("Could not sync playlist:", id, "cause:", err)
			continue
		}
		snapshots[id] = snapshot
	}

	return snapshots, nil
}

// syncPlaylist fetches the items only when the item count changed or the last
// full sync is older than FULL_SYNC_INTERVAL, and the video details only for
// the videos that do not have them yet.
func (s *LibraryService) syncPlaylist(meta *library.PlaylistMeta) (*library.Snapshot, error) {
	stored, err := s.library.FindItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	state, synced := s.state[meta.YoutubeID]
	now := s.now()
	if synced && state.itemCount == meta.ItemCount && now.Sub(state.fullSyncAt) < FULL_SYNC_INTERVAL {
		return s.snapshot(meta, stored)
	}

	fetched, err := s.youtube.FetchPlaylistItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	videos, err := s.library.FindVideosInPlaylists(meta.YoutubeID)
	if err != nil {
		return nil, err
	}
	// Videos that are new to this playlist may already be stored for another one
	var missing []library.VideoID
	for _, f := range fetched {
		if _, ok := videos[f.VideoID]; !ok {
			missing = append(missing, f.VideoID)
		}
	}
	others, err := s.library.FindVideos(unique(missing))
	if err != nil {
		return nil, err
	}
	for id, v := range others {
		videos[id] = v
	}

	changes := library.Reconcile(meta.YoutubeID, stored, fetched, videos)
	if len(changes.NeedDetails) > 0 {
		details, channels, err := s.youtube.FetchVideos(changes.NeedDetails)
		if err != nil {
			return nil, err
		}
		changes.ApplyDetails(details)
		// The channels are only used to search the videos
		if err := s.library.SaveChannels(channels); err != nil {
			log.Println("Could not save channels cause:", err)
		}
	}

	if err := s.library.ApplyItems(meta.YoutubeID, changes.Added, changes.Removed, changes.ChangedVideos()); err != nil {
		return nil, err
	}

	s.state[meta.YoutubeID] = syncState{itemCount: meta.ItemCount, fullSyncAt: now}
	if len(changes.Added) > 0 || len(changes.Removed) > 0 {
		log.Println("Synced playlist:", meta.YoutubeID, "added:", len(changes.Added), "removed:", len(changes.Removed))
	}

	items, err := s.library.FindItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	return &library.Snapshot{YoutubeID: meta.YoutubeID, Title: meta.Title, Items: items, Videos: changes.Videos, Hidden: changes.Hidden}, nil
}

func (s *LibraryService) snapshot(meta *library.PlaylistMeta, items []*library.PlaylistItem) (*library.Snapshot, error) {
	videos, err := s.library.FindVideosInPlaylists(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	return &library.Snapshot{YoutubeID: meta.YoutubeID, Title: meta.Title, Items: items, Videos: videos}, nil
}

// RefreshVideos updates the details (views, title, ...) of every video in the
// watched playlists, records their state in the history, and returns the
// videos that reached a new view milestone. Costs 1 unit per 50 videos.
func (s *LibraryService) RefreshVideos() ([]*library.Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.library.FindListedVideos()
	if err != nil {
		return nil, err
	}

	var targets []library.VideoID
	for id, v := range stored {
		if v.Available() {
			targets = append(targets, id)
		}
	}

	fetched, channels, err := s.youtube.FetchVideos(targets)
	if err != nil {
		return nil, err
	}
	// The channels are only used to search the videos
	if err := s.library.SaveChannels(channels); err != nil {
		log.Println("Could not save channels cause:", err)
	}

	now := s.now()
	var toSave []*library.Video
	var snapshots []*library.VideoSnapshot
	var reached []*library.Video
	for _, f := range fetched {
		video := stored[f.YoutubeID]
		video.ApplyDetails(f)
		if video.UpdateMilestone() {
			reached = append(reached, video)
		}
		toSave = append(toSave, video)
		snapshots = append(snapshots, library.NewVideoSnapshot(video, now))
	}
	if err := s.library.SaveVideos(toSave); err != nil {
		return nil, err
	}
	// The history is not needed for the milestones, so they are still notified
	if err := s.library.AddSnapshots(snapshots); err != nil {
		log.Println("Could not record video snapshots cause:", err)
	}

	log.Println("Refreshed videos:", len(toSave), "reached milestones:", len(reached))

	return reached, nil
}

// RefreshCategories stores the video categories as YouTube currently names
// them. Costs 1 unit.
func (s *LibraryService) RefreshCategories() error {
	categories, err := s.youtube.FetchCategories()
	if err != nil {
		return err
	}
	if err := s.library.SaveCategories(categories); err != nil {
		return err
	}
	log.Println("Refreshed video categories:", len(categories))

	return nil
}

// LiveVideos fetches the channel and current views of the given videos right
// before they are shown. Nothing is stored. Costs about 2 units per 50 videos.
func (s *LibraryService) LiveVideos(videoIDs []library.VideoID) (map[library.VideoID]*library.LiveVideo, error) {
	return s.youtube.FetchLiveVideos(unique(videoIDs))
}

// History returns the states of the video recorded at or after since. The
// video must be in one of the given playlists, or domain.ErrDBRecordNotFound
// is returned.
func (s *LibraryService) History(playlistIDs []library.PlaylistID, id library.VideoID, since time.Time) (*library.VideoHistory, error) {
	containing, err := s.PlaylistsContaining([]library.VideoID{id})
	if err != nil {
		return nil, err
	}
	if !slices.ContainsFunc(containing[id], func(p library.PlaylistID) bool { return slices.Contains(playlistIDs, p) }) {
		return nil, domain.ErrDBRecordNotFound
	}

	videos, err := s.library.FindVideos([]library.VideoID{id})
	if err != nil {
		return nil, err
	}
	video, ok := videos[id]
	if !ok || !video.HasDetails() {
		return nil, domain.ErrDBRecordNotFound
	}

	snapshots, err := s.library.FindSnapshots(id, since)
	if err != nil {
		return nil, err
	}

	return &library.VideoHistory{Video: video, Snapshots: snapshots}, nil
}

// PlaylistsContaining returns, for each given video, the playlists that contain it.
func (s *LibraryService) PlaylistsContaining(videoIDs []library.VideoID) (map[library.VideoID][]library.PlaylistID, error) {
	items, err := s.library.FindItemsByVideos(videoIDs)
	if err != nil {
		return nil, err
	}

	result := map[library.VideoID][]library.PlaylistID{}
	for _, item := range items {
		result[item.VideoYoutubeID] = append(result[item.VideoYoutubeID], item.PlaylistYoutubeID)
	}

	return result, nil
}

// Contents returns the stored items of the given playlists, ordered by the
// time they were added, together with their videos.
func (s *LibraryService) Contents(playlistIDs []library.PlaylistID) ([]*library.PlaylistVideo, error) {
	items, err := s.library.FindItems(unique(playlistIDs)...)
	if err != nil {
		return nil, err
	}

	videos, err := s.library.FindVideosInPlaylists(unique(playlistIDs)...)
	if err != nil {
		return nil, err
	}

	var contents []*library.PlaylistVideo
	for _, item := range items {
		video, ok := videos[item.VideoYoutubeID]
		if !ok {
			continue
		}
		contents = append(contents, &library.PlaylistVideo{PlaylistID: item.PlaylistYoutubeID, Item: item, Video: video})
	}

	return contents, nil
}

// RandomVideos picks up to n distinct available videos from the given playlists.
func (s *LibraryService) RandomVideos(playlistIDs []library.PlaylistID, n int) ([]*library.PlaylistVideo, error) {
	contents, err := s.Contents(playlistIDs)
	if err != nil {
		return nil, err
	}

	return library.PickRandom(contents, n), nil
}

// Trending returns the videos of the playlist that gained the most views in
// the states recorded at or after since, together with those states.
func (s *LibraryService) Trending(playlistID library.PlaylistID, since time.Time) (*library.Trending, error) {
	videos, err := s.library.FindVideosInPlaylists(playlistID)
	if err != nil {
		return nil, err
	}

	ids := make([]library.VideoID, 0, len(videos))
	for id := range videos {
		ids = append(ids, id)
	}
	earliest, err := s.library.FindEarliestSnapshots(ids, since)
	if err != nil {
		return nil, err
	}
	latest, err := s.library.FindLatestSnapshots(ids, since)
	if err != nil {
		return nil, err
	}

	trending := library.ComputeTrending(videos, earliest, latest, library.TRENDING_TOP)
	for _, v := range trending.Videos {
		if v.Snapshots, err = s.library.FindSnapshots(v.Video.YoutubeID, since); err != nil {
			return nil, err
		}
	}

	return trending, nil
}

// Stats aggregates the stored contents of the playlist as of now.
func (s *LibraryService) Stats(playlistID library.PlaylistID) (*library.Stats, error) {
	contents, err := s.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeStats(contents, s.now(), s.location), nil
}

// Wrapped aggregates the items added to the playlist in the year.
func (s *LibraryService) Wrapped(playlistID library.PlaylistID, year int) (*library.Wrapped, error) {
	contents, err := s.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeWrapped(contents, year, s.location), nil
}

func unique[T comparable](ids []T) []T {
	seen := map[T]bool{}
	var result []T
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}
