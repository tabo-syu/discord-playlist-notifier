package application

import (
	"log"
	"maps"
	"sync"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

// LibraryService は通知対象のプレイリストの中身を DB に保持する。同期で起きたこと
// （非公開・節目・同期の結果）は、保存した後にドメインイベントとして events に発行し、
// 誰が受け取るかは知らない。
type LibraryService struct {
	source        library.Source
	subscriptions subscription.Repository
	playlists     library.PlaylistRepository
	videos        library.VideoRepository
	channels      library.ChannelRepository
	categories    library.CategoryRepository
	snapshots     library.VideoSnapshotRepository
	catalog       *library.Catalog
	events        *EventPublisher
	// 日付を月や年で区切るのに使うタイムゾーン
	location *time.Location

	// 重なって実行されたときに同じ動画を書き込まないようにする
	mu sync.Mutex
	// プレイリストごとの、前回の全件取得の記録。全件取得するかはドメイン
	// （library.FullSync）が決め、ここは記録を置いておくだけ。再起動すると消える
	fullSyncs map[library.PlaylistID]library.FullSync
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func NewLibraryService(
	src library.Source,
	s subscription.Repository,
	p library.PlaylistRepository,
	v library.VideoRepository,
	c library.ChannelRepository,
	cat library.CategoryRepository,
	vs library.VideoSnapshotRepository,
	ct *library.Catalog,
	e *EventPublisher,
	loc *time.Location,
	now func() time.Time,
) *LibraryService {
	return &LibraryService{
		source:        src,
		subscriptions: s,
		playlists:     p,
		videos:        v,
		channels:      c,
		categories:    cat,
		snapshots:     vs,
		catalog:       ct,
		events:        e,
		location:      loc,
		fullSyncs:     map[library.PlaylistID]library.FullSync{},
		now:           now,
	}
}

// Sync は通知登録された全てのプレイリストの中身を最新にし、保存した後に、非公開・
// 削除になった動画（VideoHidden）と、同期の結果（PlaylistSynced）を発行する。
// 同期に失敗したプレイリストはログに出して、どちらのイベントも出さない。
//
// 発行は全てのプレイリストを同期し終えてから、VideoHidden、PlaylistSynced の順に
// まとめて行う。非公開の通知を新着の通知より先に送るため。
func (s *LibraryService) Sync() {
	subscriptions, err := s.subscriptions.FindAll()
	if err != nil {
		// 文言は CLAUDE.md の「既知の挙動」にあるもののまま
		log.Println("Could not notify cause:", err)

		return
	}

	var ids []library.PlaylistID
	for _, sub := range subscriptions {
		ids = append(ids, sub.YoutubeID)
	}

	synced, hidden, err := s.sync(ids)
	if err != nil {
		log.Println("Could not notify cause:", err)

		return
	}
	// プレイリストと動画は sync の中で保存済み。ロックを外した後に配る
	Publish(s.events, hidden)
	Publish(s.events, synced)
}

// RefreshVideos は通知対象のプレイリストにある全動画の詳細（再生数、タイトルなど）を
// 更新し、その状態を履歴に記録して、再生数の新しい節目に達した動画のイベントを
// 発行する。イベントは動画を保存できたときだけ発行する。
// 50 動画ごとに 1 ユニットを消費する。
func (s *LibraryService) RefreshVideos() {
	reached, err := s.refreshVideos()
	if err != nil {
		log.Println("Could not refresh videos cause:", err)

		return
	}
	// 動画は refreshVideos の中で保存済み。ロックを外した後に配る
	Publish(s.events, reached)
}

// RefreshCategories は動画のカテゴリを、YouTube での現在の名前で保存する。
// 1 ユニットを消費する。
func (s *LibraryService) RefreshCategories() {
	categories, err := s.source.FetchCategories()
	if err != nil {
		log.Println("Could not fetch video categories cause:", err)

		return
	}

	if err := s.categories.Save(categories); err != nil {
		log.Println("Could not save video categories cause:", err)

		return
	}

	log.Println("Refreshed video categories:", len(categories))
}

// History は scope のプレイリストにある動画の、since 以降に記録された状態を返す。
// 見せられない動画（scope の外にしかない、詳細を取得していない）には
// library.ErrVideoNotFound を返す。
func (s *LibraryService) History(scope []library.PlaylistID, id library.VideoID, since time.Time) (*library.VideoHistory, error) {
	video, err := s.catalog.FindVideo(scope, id)
	if err != nil {
		return nil, err
	}

	snapshots, err := s.snapshots.Find(id, since)
	if err != nil {
		return nil, err
	}

	return &library.VideoHistory{Video: video, Snapshots: snapshots}, nil
}

// RandomVideos は指定したプレイリストから、重複のない視聴可能な動画を最大 n 件選ぶ。
func (s *LibraryService) RandomVideos(playlistIDs []library.PlaylistID, n int) ([]*library.PlaylistVideo, error) {
	contents, err := s.catalog.Contents(playlistIDs)
	if err != nil {
		return nil, err
	}

	return library.PickRandom(contents, n), nil
}

// Trending は since 以降に記録された状態のうちで再生数が最も伸びたプレイリストの動画を、
// その状態と一緒に返す。
func (s *LibraryService) Trending(playlistID library.PlaylistID, since time.Time) (*library.Trending, error) {
	videos, err := s.videos.FindInPlaylists(playlistID)
	if err != nil {
		return nil, err
	}

	ids := make([]library.VideoID, 0, len(videos))
	for id := range videos {
		ids = append(ids, id)
	}

	earliest, err := s.snapshots.FindEarliest(ids, since)
	if err != nil {
		return nil, err
	}

	latest, err := s.snapshots.FindLatest(ids, since)
	if err != nil {
		return nil, err
	}

	trending := library.ComputeTrending(videos, earliest, latest, library.TRENDING_TOP)
	for _, v := range trending.Videos {
		if v.Snapshots, err = s.snapshots.Find(v.Video.YoutubeID, since); err != nil {
			return nil, err
		}
	}

	return trending, nil
}

// Stats はプレイリストの保存済みの中身を現時点で集計する。
func (s *LibraryService) Stats(playlistID library.PlaylistID) (*library.Stats, error) {
	contents, err := s.catalog.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeStats(contents, s.now(), s.location), nil
}

// Wrapped はその年にプレイリストへ追加されたアイテムを集計する。
func (s *LibraryService) Wrapped(playlistID library.PlaylistID, year int) (*library.Wrapped, error) {
	contents, err := s.catalog.Contents([]library.PlaylistID{playlistID})
	if err != nil {
		return nil, err
	}

	return library.ComputeWrapped(contents, year, s.location), nil
}

// sync は指定したプレイリストの保存済みの中身を最新にして、同期の結果を返す。
// 同期に失敗したプレイリストはログに出して結果から除く。今回の同期で非公開・削除に
// なった動画のイベントも返す。イベントは動画を保存できたプレイリストの分だけ。
// panic してもロックが残らないように、defer で外す。
func (s *LibraryService) sync(playlistIDs []library.PlaylistID) ([]*library.PlaylistSynced, []*library.VideoHidden, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := unique(playlistIDs)

	metas, err := s.source.FetchPlaylistMetas(ids...)
	if err != nil {
		return nil, nil, err
	}

	var (
		synced []*library.PlaylistSynced
		hidden []*library.VideoHidden
	)

	for _, id := range ids {
		meta, ok := metas[id]
		if !ok {
			synced = append(synced, &library.PlaylistSynced{
				Snapshot: &library.Snapshot{Playlist: &library.Playlist{YoutubeID: id}, Deleted: true},
			})

			continue
		}

		snapshot, h, err := s.syncPlaylist(meta)
		if err != nil {
			log.Println("Could not sync playlist:", id, "cause:", err)

			continue
		}

		synced = append(synced, &library.PlaylistSynced{Snapshot: snapshot})

		hidden = append(hidden, h...)
	}

	return synced, hidden, nil
}

// refreshVideos は全動画の詳細を更新して保存し、新しい節目に達した動画のイベントを
// 返す。panic してもロックが残らないように、defer で外す。
func (s *LibraryService) refreshVideos() ([]*library.MilestoneReached, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.videos.FindListed()
	if err != nil {
		return nil, err
	}

	var targets []library.VideoID

	for id, v := range stored {
		if v.Available() {
			targets = append(targets, id)
		}
	}

	fetched, err := s.source.FetchVideoDetails(targets)
	if err != nil {
		return nil, err
	}
	// チャンネルは動画の検索にしか使わない
	if err = s.channels.Save(library.ChannelsOf(fetched)); err != nil {
		log.Println("Could not save channels cause:", err)
	}

	now := s.now()

	var (
		toSave    []*library.Video
		snapshots []*library.VideoSnapshot
		reached   []*library.MilestoneReached
	)

	for _, f := range fetched {
		video := stored[f.VideoID]
		video.ApplyDetails(f)

		if e := video.UpdateMilestone(); e != nil {
			reached = append(reached, e)
		}

		toSave = append(toSave, video)
		snapshots = append(snapshots, library.NewVideoSnapshot(video, now))
	}

	if err = s.videos.Save(toSave); err != nil {
		return nil, err
	}
	// 履歴は節目の判定に使わないので、記録に失敗しても節目は通知する
	if err = s.snapshots.Add(snapshots); err != nil {
		log.Println("Could not record video snapshots cause:", err)
	}

	log.Println("Refreshed videos:", len(toSave), "reached milestones:", len(reached))

	return reached, nil
}

// syncPlaylist は、全件取得の時期が来ているとき（library.FullSync.Due）だけ
// アイテムを取得し、動画の詳細はまだ持っていない動画の分だけ取得する。
//
// プレイリストと動画は別の集約なので、別々に保存する。プレイリストを先に保存し、
// 動画の保存に失敗したらエラーを返す。その回は通知せず、全件取得の記録も更新しない
// ので、次の同期で公開状態の変化と新しい動画をもう一度検知できる。逆の順だと、
// 動画の公開状態を保存した後にプレイリストの保存に失敗したとき、次の同期では
// 公開状態がもう変わっていて、非公開になったことを通知できない。
//
// 同じ回に同期した別のプレイリストにも同じ動画があり、そちらで公開状態を保存
// できたときは、次の同期ではもう変化を検知しないが、そちらの VideoHidden で通知
// する。通知先はハンドラが DB に保存したアイテムで決めるので、このプレイリストの
// チャンネルにも届く。
func (s *LibraryService) syncPlaylist(meta *library.PlaylistMeta) (*library.Snapshot, []*library.VideoHidden, error) {
	playlist, err := s.playlists.Find(meta.YoutubeID)
	if err != nil {
		return nil, nil, err
	}

	now := s.now()
	if !s.fullSyncs[meta.YoutubeID].Due(meta, now) {
		var snapshot *library.Snapshot

		snapshot, err = s.snapshot(meta, playlist)

		return snapshot, nil, err
	}

	fetched, err := s.source.FetchPlaylistItems(meta.YoutubeID)
	if err != nil {
		return nil, nil, err
	}

	videos, err := s.videos.FindInPlaylists(meta.YoutubeID)
	if err != nil {
		return nil, nil, err
	}
	// このプレイリストで初めて見る動画でも、別のプレイリストの分として保存済みのことがある
	var missing []library.VideoID

	for _, f := range fetched {
		if _, ok := videos[f.VideoID]; !ok {
			missing = append(missing, f.VideoID)
		}
	}

	others, err := s.videos.Find(unique(missing))
	if err != nil {
		return nil, nil, err
	}

	maps.Copy(videos, others)

	added, removed := playlist.Reconcile(fetched)

	changes := library.SyncVideos(fetched, videos)
	if len(changes.NeedDetails) > 0 {
		var details []*library.VideoDetails

		details, err = s.source.FetchVideoDetails(changes.NeedDetails)
		if err != nil {
			return nil, nil, err
		}

		changes.ApplyDetails(details)
		// チャンネルは動画の検索にしか使わない
		if err = s.channels.Save(library.ChannelsOf(details)); err != nil {
			log.Println("Could not save channels cause:", err)
		}
	}

	if err = s.playlists.Save(playlist); err != nil {
		return nil, nil, err
	}

	if err = s.videos.Save(changes.ChangedVideos()); err != nil {
		return nil, nil, err
	}

	s.fullSyncs[meta.YoutubeID] = library.NewFullSync(meta, now)
	if len(added) > 0 || len(removed) > 0 {
		log.Println("Synced playlist:", meta.YoutubeID, "added:", len(added), "removed:", len(removed))
	}

	return &library.Snapshot{Playlist: playlist, Title: meta.Title, Videos: changes.Videos}, changes.Hidden, nil
}

func (s *LibraryService) snapshot(meta *library.PlaylistMeta, playlist *library.Playlist) (*library.Snapshot, error) {
	videos, err := s.videos.FindInPlaylists(meta.YoutubeID)
	if err != nil {
		return nil, err
	}

	return &library.Snapshot{Playlist: playlist, Title: meta.Title, Videos: videos}, nil
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
