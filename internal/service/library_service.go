package service

import (
	"log"
	"sync"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

// PlaylistSnapshot is the current content of a YouTube playlist.
type PlaylistSnapshot struct {
	YoutubeID string
	Title     string
	// Deleted is true when the playlist no longer exists on YouTube.
	Deleted bool
	Items   []*domain.PlaylistItem
	Videos  map[string]*domain.Video
}

// PlaylistVideo is a video together with the playlist item it was found in.
type PlaylistVideo struct {
	PlaylistID string
	Item       *domain.PlaylistItem
	Video      *domain.Video
}

// LibraryService keeps the contents of the watched playlists in the database.
type LibraryService struct {
	youtube repository.YouTubeRepository
	library repository.LibraryRepository

	// Keeps overlapping runs from writing the same videos
	mu sync.Mutex
}

func NewLibraryService(y repository.YouTubeRepository, l repository.LibraryRepository) *LibraryService {
	return &LibraryService{youtube: y, library: l}
}

// Sync brings the stored contents of the given playlists up to date and
// returns them. Playlists that failed to sync are logged and left out.
func (s *LibraryService) Sync(playlistIds []string) (map[string]*PlaylistSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := unique(playlistIds)
	metas, err := s.youtube.FetchPlaylistMetas(ids...)
	if err != nil {
		return nil, err
	}

	snapshots := map[string]*PlaylistSnapshot{}
	for _, id := range ids {
		meta, ok := metas[id]
		if !ok {
			snapshots[id] = &PlaylistSnapshot{YoutubeID: id, Deleted: true}
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

func (s *LibraryService) syncPlaylist(meta *repository.PlaylistMeta) (*PlaylistSnapshot, error) {
	stored, err := s.library.FindItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	fetched, err := s.youtube.FetchPlaylistItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	storedById := map[string]*domain.PlaylistItem{}
	for _, item := range stored {
		storedById[item.ItemID] = item
	}
	fetchedById := map[string]bool{}
	var added []*domain.PlaylistItem
	var videoIds []string
	for _, f := range fetched {
		fetchedById[f.ItemID] = true
		videoIds = append(videoIds, f.VideoID)
		if _, ok := storedById[f.ItemID]; !ok {
			added = append(added, &domain.PlaylistItem{
				PlaylistYoutubeID: meta.YoutubeID,
				ItemID:            f.ItemID,
				VideoYoutubeID:    f.VideoID,
				AddedAt:           f.AddedAt,
			})
		}
	}
	var removed []*domain.PlaylistItem
	for _, item := range stored {
		if !fetchedById[item.ItemID] {
			removed = append(removed, item)
		}
	}

	videos, err := s.library.FindVideosInPlaylists(meta.YoutubeID)
	if err != nil {
		return nil, err
	}
	// Videos that are new to this playlist may already be stored for another one
	var missing []string
	for _, id := range unique(videoIds) {
		if _, ok := videos[id]; !ok {
			missing = append(missing, id)
		}
	}
	others, err := s.library.FindVideos(missing)
	if err != nil {
		return nil, err
	}
	for id, v := range others {
		videos[id] = v
	}

	// The privacy of a video is taken from its playlist items, which also
	// report private and deleted videos, unlike the videos endpoint.
	changed := map[string]*domain.Video{}
	var needDetails []string
	for _, f := range fetched {
		video, ok := videos[f.VideoID]
		if !ok {
			video = &domain.Video{YoutubeID: f.VideoID}
			videos[f.VideoID] = video
		}
		if video.PrivacyStatus != f.PrivacyStatus {
			video.PrivacyStatus = f.PrivacyStatus
			changed[f.VideoID] = video
		}
		if !video.HasDetails() && (f.PrivacyStatus == domain.PrivacyPublic || f.PrivacyStatus == domain.PrivacyUnlisted) {
			needDetails = append(needDetails, f.VideoID)
		}
	}

	if len(needDetails) > 0 {
		details, err := s.youtube.FetchVideos(unique(needDetails))
		if err != nil {
			return nil, err
		}
		for _, d := range details {
			video := videos[d.YoutubeID]
			mergeDetails(video, d)
			if video.ViewMilestone == nil {
				reached := MilestoneFor(video.Views)
				video.ViewMilestone = &reached
			}
			changed[d.YoutubeID] = video
		}
	}

	var toSave []*domain.Video
	for _, v := range changed {
		toSave = append(toSave, v)
	}
	if err := s.library.ApplyItems(meta.YoutubeID, added, removed, toSave); err != nil {
		return nil, err
	}

	if len(added) > 0 || len(removed) > 0 {
		log.Println("Synced playlist:", meta.YoutubeID, "added:", len(added), "removed:", len(removed))
	}

	items, err := s.library.FindItems(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	return &PlaylistSnapshot{YoutubeID: meta.YoutubeID, Title: meta.Title, Items: items, Videos: videos}, nil
}

// LiveVideos fetches the channel and current views of the given videos right
// before they are shown. Nothing is stored. Costs about 2 units per 50 videos.
func (s *LibraryService) LiveVideos(videoIds []string) (map[string]*domain.LiveVideo, error) {
	return s.youtube.FetchLiveVideos(unique(videoIds))
}

// mergeDetails copies the details fetched from the videos endpoint. The
// privacy status is left untouched because it is owned by Sync.
func mergeDetails(dst *domain.Video, src *domain.Video) {
	dst.Title = src.Title
	dst.Views = src.Views
	dst.PublishedAt = src.PublishedAt
}

func unique(ids []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}

// RefreshVideos updates the details (views, title, ...) of every video in the
// watched playlists, and returns the videos that reached a new view
// milestone. Costs 1 unit per 50 videos.
func (s *LibraryService) RefreshVideos() ([]*domain.Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.library.FindListedVideos()
	if err != nil {
		return nil, err
	}

	var targets []string
	for id, v := range stored {
		if v.Available() {
			targets = append(targets, id)
		}
	}

	fetched, err := s.youtube.FetchVideos(targets)
	if err != nil {
		return nil, err
	}

	var toSave []*domain.Video
	var reached []*domain.Video
	for _, f := range fetched {
		video := stored[f.YoutubeID]
		mergeDetails(video, f)

		milestone := MilestoneFor(video.Views)
		if video.ViewMilestone != nil && milestone > *video.ViewMilestone {
			reached = append(reached, video)
		}
		if video.ViewMilestone == nil || milestone > *video.ViewMilestone {
			video.ViewMilestone = &milestone
		}
		toSave = append(toSave, video)
	}
	if err := s.library.SaveVideos(toSave); err != nil {
		return nil, err
	}

	log.Println("Refreshed videos:", len(toSave), "reached milestones:", len(reached))

	return reached, nil
}

// PlaylistsContaining returns, for each given video, the YouTube IDs of the
// playlists that contain it.
func (s *LibraryService) PlaylistsContaining(videoIds []string) (map[string][]string, error) {
	items, err := s.library.FindItemsByVideos(videoIds)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}
	for _, item := range items {
		result[item.VideoYoutubeID] = append(result[item.VideoYoutubeID], item.PlaylistYoutubeID)
	}

	return result, nil
}
