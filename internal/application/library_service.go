package application

import (
	"log"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// プレイリストのアイテム数が変わっていなくても、この間隔が経てばアイテムを
// 取得し直す。動画の除外や公開状態の変化と同時に起きた追加を拾うため。
const FULL_SYNC_INTERVAL = time.Hour

type syncState struct {
	itemCount  int64
	fullSyncAt time.Time
}

// LibraryService は通知対象のプレイリストの中身を DB に保持する。
type LibraryService struct {
	youtube library.YouTube
	library library.Repository
	// 日付を月や年で区切るのに使うタイムゾーン
	location *time.Location

	// 重なって実行されたときに同じ動画を書き込まないようにする
	mu    sync.Mutex
	state map[library.PlaylistID]syncState
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewLibraryService(y library.YouTube, l library.Repository, loc *time.Location, now func() time.Time) *LibraryService {
	return &LibraryService{youtube: y, library: l, location: loc, state: map[library.PlaylistID]syncState{}, now: now}
}

// Sync は指定したプレイリストの保存済みの中身を最新にして返す。
// 同期に失敗したプレイリストはログに出して結果から除く。
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

// RefreshVideos は通知対象のプレイリストにある全動画の詳細（再生数、タイトルなど）を
// 更新し、その状態を履歴に記録して、再生数の新しい節目に達した動画を返す。
// 50 動画ごとに 1 ユニットを消費する。
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
	// チャンネルは動画の検索にしか使わない
	if err := s.library.SaveChannels(channels); err != nil {
		log.Println("Could not save channels cause:", err)
	}

	now := s.now()

	var (
		toSave    []*library.Video
		snapshots []*library.VideoSnapshot
		reached   []*library.Video
	)

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
	// 履歴は節目の判定に使わないので、記録に失敗しても節目は通知する
	if err := s.library.AddSnapshots(snapshots); err != nil {
		log.Println("Could not record video snapshots cause:", err)
	}

	log.Println("Refreshed videos:", len(toSave), "reached milestones:", len(reached))

	return reached, nil
}

// RefreshCategories は動画のカテゴリを、YouTube での現在の名前で保存する。
// 1 ユニットを消費する。
func (s *LibraryService) RefreshCategories() {
	categories, err := s.youtube.FetchCategories()
	if err != nil {
		log.Println("Could not fetch video categories cause:", err)

		return
	}

	if err := s.library.SaveCategories(categories); err != nil {
		log.Println("Could not save video categories cause:", err)

		return
	}

	log.Println("Refreshed video categories:", len(categories))
}

// LiveVideos は指定した動画のチャンネルと現在の再生数を、表示する直前に取得する。
// 何も保存しない。50 動画ごとに約 2 ユニットを消費する。
func (s *LibraryService) LiveVideos(videoIDs []library.VideoID) (map[library.VideoID]*library.LiveVideo, error) {
	return s.youtube.FetchLiveVideos(unique(videoIDs))
}

// History は since 以降に記録された動画の状態を返す。動画が指定した
// プレイリストのどれかに含まれていなければ library.ErrVideoNotFound を
// 返す。
func (s *LibraryService) History(playlistIDs []library.PlaylistID, id library.VideoID, since time.Time) (*library.VideoHistory, error) {
	containing, err := s.PlaylistsContaining([]library.VideoID{id})
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(containing[id], func(p library.PlaylistID) bool { return slices.Contains(playlistIDs, p) }) {
		return nil, library.ErrVideoNotFound
	}

	videos, err := s.library.FindVideos([]library.VideoID{id})
	if err != nil {
		return nil, err
	}

	video, ok := videos[id]
	if !ok || !video.HasDetails() {
		return nil, library.ErrVideoNotFound
	}

	snapshots, err := s.library.FindSnapshots(id, since)
	if err != nil {
		return nil, err
	}

	return &library.VideoHistory{Video: video, Snapshots: snapshots}, nil
}

// PlaylistsContaining は指定した動画ごとに、それを含むプレイリストを返す。
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

// Contents は指定したプレイリストの保存済みのアイテムを、追加された時刻の順に
// 動画と一緒に返す。
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

// RandomVideos は指定したプレイリストから、重複のない視聴可能な動画を最大 n 件選ぶ。
func (s *LibraryService) RandomVideos(playlistIDs []library.PlaylistID, n int) ([]*library.PlaylistVideo, error) {
	contents, err := s.Contents(playlistIDs)
	if err != nil {
		return nil, err
	}

	return library.PickRandom(contents, n), nil
}

// Trending は since 以降に記録された状態のうちで再生数が最も伸びたプレイリストの動画を、
// その状態と一緒に返す。
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

// Stats はプレイリストの保存済みの中身を現時点で集計する。
func (s *LibraryService) Stats(playlistID library.PlaylistID) (*library.Stats, error) {
	contents, err := s.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeStats(contents, s.now(), s.location), nil
}

// Wrapped はその年にプレイリストへ追加されたアイテムを集計する。
func (s *LibraryService) Wrapped(playlistID library.PlaylistID, year int) (*library.Wrapped, error) {
	contents, err := s.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeWrapped(contents, year, s.location), nil
}

// syncPlaylist は、アイテム数が変わったか前回の全件取得から
// FULL_SYNC_INTERVAL より経っているときだけアイテムを取得し、
// 動画の詳細はまだ持っていない動画の分だけ取得する。
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
	// このプレイリストで初めて見る動画でも、別のプレイリストの分として保存済みのことがある
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

	maps.Copy(videos, others)

	changes := library.Reconcile(meta.YoutubeID, stored, fetched, videos)
	if len(changes.NeedDetails) > 0 {
		var (
			details  []*library.Video
			channels []*library.Channel
		)

		details, channels, err = s.youtube.FetchVideos(changes.NeedDetails)
		if err != nil {
			return nil, err
		}

		changes.ApplyDetails(details)
		// チャンネルは動画の検索にしか使わない
		if err = s.library.SaveChannels(channels); err != nil {
			log.Println("Could not save channels cause:", err)
		}
	}

	if err = s.library.ApplyItems(meta.YoutubeID, changes.Added, changes.Removed, changes.ChangedVideos()); err != nil {
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
