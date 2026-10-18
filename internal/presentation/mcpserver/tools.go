package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// タグが数十個ある動画もある。全て検索するが、結果を小さく保つため
	// 返すのは先頭のいくつかだけにする。
	MAX_TAGS_SHOWN        = 10
	DEFAULT_VIDEO_LIMIT   = 20
	MAX_VIDEO_LIMIT       = 100
	MAX_RANDOM_COUNT      = 10
	DATE_FORMAT           = "2006-01-02"
	MONTH_FORMAT          = "2006-01"
	ERR_NOT_WATCHED       = "登録されていないプレイリストです。list_playlists で ID を確認してください。"
	ERR_INTERNAL          = "データベースの読み込みに失敗しました。時間をおいてもう一度試してください。"
	YOUTUBE_WATCH_URL     = "https://www.youtube.com/watch?v=%s&list=%s"
	YOUTUBE_PLAYLIST_URL  = "https://www.youtube.com/playlist?list=%s"
	YOUTUBE_VIDEO_URL     = "https://www.youtube.com/watch?v=%s"
	ERR_VIDEO_NOT_WATCHED = "登録されているプレイリストに、その動画は見つかりません。find_videos で video_id を確認してください。"
)

// クライアントの接続時に伝える
const INSTRUCTIONS = `Discord の playlist-notifier ボットが通知登録している YouTube プレイリストの中身を調べるためのツールです。
- 見られるのは通知登録されているプレイリストだけで、読み取り専用です。
- まず list_playlists でプレイリストの ID と名前を確認し、ほかのツールにはその ID を渡してください。
- データはボットのデータベースの内容です。プレイリストの中身は 5 分ごと、再生数は 6 時間ごとに更新されるので、YouTube の最新の状態と少しずれることがあります。
- 曲のチャンネル名・長さ・タグ・カテゴリは 6 時間ごとの更新で取得します。取得前の曲では空になります。
- 日時はボットのタイムゾーン（RFC 3339 形式）で返します。`

type tools struct {
	query    *application.QueryService
	location *time.Location
	// 現在時刻。テストで固定できるように外から渡す
	now func() time.Time
}

func readOnly(name, title, description string) *mcp.Tool {
	closed := false

	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: &closed},
	}
}

func addTools(s *mcp.Server, t *tools) {
	mcp.AddTool(s, readOnly("list_playlists", "プレイリストの一覧",
		"通知登録されている YouTube プレイリストの一覧を、曲数や最後に曲が追加された日時と一緒に返します。"),
		logged("list_playlists", t.listPlaylists))
	mcp.AddTool(s, readOnly("find_videos", "曲を探す",
		"プレイリストの曲を、曲名・チャンネル名・タグの一部、カテゴリ、曲の長さ、追加された期間で絞り込んで返します。"+
			"「最近追加された曲」「〇〇が入っているか」「再生数の多い曲」などを調べるのに使います。"+
			"同じ曲が複数のプレイリストにある場合や、同じプレイリストに 2 回追加されている場合は、追加ごとに 1 件として返します。"),
		logged("find_videos", t.findVideos))
	mcp.AddTool(s, readOnly("playlist_stats", "プレイリストの統計",
		"プレイリストの曲数（観られる・非公開・削除の内訳）、追加のペース、再生数ランキング、直近 6 か月の月別の追加数を返します。"),
		logged("playlist_stats", t.playlistStats))
	mcp.AddTool(s, readOnly("playlist_wrapped", "年間まとめ",
		"プレイリストに 1 年間で追加された曲のまとめ（曲数、前年の曲数、一番追加した月、最初と最後の一曲、再生数ランキング）を返します。"),
		logged("playlist_wrapped", t.playlistWrapped))
	mcp.AddTool(s, readOnly("random_videos", "ランダムな曲",
		"プレイリストから観られる曲をランダムに選んで返します。同じ曲は 2 回選ばれません。"),
		logged("random_videos", t.randomVideos))
	mcp.AddTool(s, readOnly("video_history", "再生数などの推移",
		"曲のタイトル・公開状態・再生数・高評価数・コメント数を、6 時間ごとに記録した履歴として古い順に返します。"+
			"video_id は find_videos などで確認してください。"),
		logged("video_history", t.videoHistory))
}

func logged[In, Out any](name string, h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		res, out, err := h(ctx, req, in)
		if err != nil {
			log.Println("MCP tool failed:", name, "cause:", err)
		} else {
			log.Println("MCP tool called:", name)
		}

		return res, out, err
	}
}

// clientError はユースケースのエラーを、クライアントに見せるメッセージに変換する。
// 内部エラーの原因はログに出すだけにする。
func clientError(err error) error {
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return errors.New(ERR_NOT_WATCHED)
	}

	log.Println("MCP query failed cause:", err)

	return errors.New(ERR_INTERNAL)
}

// ---------- 出力 ----------

type playlistOut struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type videoOut struct {
	VideoID       string   `json:"video_id"`
	Title         string   `json:"title" jsonschema:"曲名。非公開・削除の曲では空のことがある"`
	URL           string   `json:"url"`
	PlaylistID    string   `json:"playlist_id"`
	PlaylistTitle string   `json:"playlist_title"`
	AddedAt       string   `json:"added_at" jsonschema:"プレイリストに追加された日時"`
	PublishedAt   string   `json:"published_at,omitempty" jsonschema:"動画が YouTube に投稿された日時"`
	Views         uint64   `json:"views"`
	Status        string   `json:"status" jsonschema:"public（公開）、unlisted（限定公開）、private（非公開）、deleted（削除）のいずれか"`
	Channel       string   `json:"channel,omitempty" jsonschema:"動画を投稿したチャンネルの名前"`
	ChannelID     string   `json:"channel_id,omitempty" jsonschema:"チャンネルの ID。同じ名前の別のチャンネルを見分けるときに使う"`
	DurationSec   int      `json:"duration_seconds,omitempty" jsonschema:"曲の長さ（秒）。不明なときは省略"`
	Tags          []string `json:"tags" jsonschema:"投稿者が付けたタグ（最初の 10 個まで）。付いていないことも多い"`
	Category      string   `json:"category,omitempty" jsonschema:"YouTube のカテゴリ（音楽、ゲームなど）"`
}

type rankedVideoOut struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Views   uint64 `json:"views"`
}

type monthCountOut struct {
	Month string `json:"month" jsonschema:"YYYY-MM"`
	Count int    `json:"count"`
}

func (t *tools) timeOut(at time.Time) string {
	if at.IsZero() {
		return ""
	}

	return at.In(t.location).Format(time.RFC3339)
}

func toPlaylist(p *application.WatchedPlaylist) playlistOut {
	return playlistOut{ID: string(p.ID), Title: p.Title, URL: fmt.Sprintf(YOUTUBE_PLAYLIST_URL, p.ID)}
}

// videoNames は動画が ID で参照する名前を持つ。
type videoNames struct {
	playlists  map[library.PlaylistID]string
	channels   map[library.ChannelID]*library.Channel
	categories map[library.CategoryID]*library.Category
}

func (t *tools) toVideo(c *library.PlaylistVideo, names *videoNames) videoOut {
	v := c.Video

	out := videoOut{
		VideoID:       string(v.YoutubeID),
		Title:         v.Title,
		URL:           fmt.Sprintf(YOUTUBE_WATCH_URL, v.YoutubeID, c.PlaylistID),
		PlaylistID:    string(c.PlaylistID),
		PlaylistTitle: names.playlists[c.PlaylistID],
		AddedAt:       t.timeOut(c.Item.AddedAt),
		Views:         uint64(v.Views),
		Status:        string(v.PrivacyStatus),
		ChannelID:     string(v.ChannelYoutubeID),
		DurationSec:   int(v.Duration.Seconds()),
		Tags:          []string{},
	}
	if v.HasDetails() {
		out.PublishedAt = t.timeOut(v.PublishedAt)
	}

	if ch, ok := names.channels[v.ChannelYoutubeID]; ok {
		out.Channel = ch.Title
	}

	if v.Tags != nil {
		out.Tags = v.Tags[:min(len(v.Tags), MAX_TAGS_SHOWN)]
	}

	if cat, ok := names.categories[v.CategoryID]; ok {
		out.Category = cat.Title
	}

	return out
}

func (t *tools) toVideos(contents []*library.PlaylistVideo) ([]videoOut, error) {
	names := &videoNames{playlists: map[library.PlaylistID]string{}}

	playlists, err := t.query.Playlists()
	if err != nil {
		return nil, err
	}

	for _, p := range playlists {
		names.playlists[p.ID] = p.Title
	}

	var channelIDs []library.ChannelID
	for _, c := range contents {
		channelIDs = append(channelIDs, c.Video.ChannelYoutubeID)
	}

	if names.channels, err = t.query.Channels(channelIDs); err != nil {
		return nil, err
	}

	if names.categories, err = t.query.Categories(); err != nil {
		return nil, err
	}

	videos := []videoOut{}
	for _, c := range contents {
		videos = append(videos, t.toVideo(c, names))
	}

	return videos, nil
}

func toRanked(videos []*library.Video, playlistID library.PlaylistID) []rankedVideoOut {
	ranked := []rankedVideoOut{}
	for _, v := range videos {
		ranked = append(ranked, rankedVideoOut{
			VideoID: string(v.YoutubeID),
			Title:   v.Title,
			URL:     fmt.Sprintf(YOUTUBE_WATCH_URL, v.YoutubeID, playlistID),
			Views:   uint64(v.Views),
		})
	}

	return ranked
}

// parseDate は YYYY-MM-DD を、ボットのタイムゾーンでのその日の始まりとして読む。
func (t *tools) parseDate(name, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	date, err := time.ParseInLocation(DATE_FORMAT, value, t.location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s は YYYY-MM-DD の形式で指定してください: %q", name, value)
	}

	return date, nil
}

// ---------- list_playlists ----------

type listPlaylistsIn struct{}

type playlistOverviewOut struct {
	playlistOut
	Items        int    `json:"items" jsonschema:"曲数（同じ曲の重複も数える）"`
	Available    int    `json:"available" jsonschema:"観られる曲の数"`
	Private      int    `json:"private"`
	Deleted      int    `json:"deleted"`
	FirstAddedAt string `json:"first_added_at,omitempty"`
	LastAddedAt  string `json:"last_added_at,omitempty" jsonschema:"最後に曲が追加された日時"`
}

type listPlaylistsOut struct {
	Playlists []playlistOverviewOut `json:"playlists"`
}

func (t *tools) listPlaylists(_ context.Context, _ *mcp.CallToolRequest, _ listPlaylistsIn) (*mcp.CallToolResult, listPlaylistsOut, error) {
	overviews, err := t.query.Overviews()
	if err != nil {
		return nil, listPlaylistsOut{}, clientError(err)
	}

	out := listPlaylistsOut{Playlists: []playlistOverviewOut{}}
	for _, o := range overviews {
		out.Playlists = append(out.Playlists, playlistOverviewOut{
			playlistOut:  toPlaylist(&o.WatchedPlaylist),
			Items:        o.Stats.Total,
			Available:    o.Stats.Available,
			Private:      o.Stats.Private,
			Deleted:      o.Stats.Deleted,
			FirstAddedAt: t.timeOut(o.Stats.FirstAddedAt),
			LastAddedAt:  t.timeOut(o.Stats.LastAddedAt),
		})
	}

	return nil, out, nil
}

// ---------- find_videos ----------

type findVideosIn struct {
	PlaylistID         string `json:"playlist_id,omitempty" jsonschema:"プレイリストの ID。省略するとすべてのプレイリストから探す"`
	Title              string `json:"title,omitempty" jsonschema:"曲名に含まれる文字列（大文字・小文字は区別しない）"`
	Channel            string `json:"channel,omitempty" jsonschema:"チャンネル名に含まれる文字列（大文字・小文字は区別しない）"`
	ChannelID          string `json:"channel_id,omitempty" jsonschema:"チャンネルの ID（videos の channel_id）。名前より確実に絞り込める"`
	Tag                string `json:"tag,omitempty" jsonschema:"どれかのタグに含まれる文字列（大文字・小文字は区別しない）"`
	Category           string `json:"category,omitempty" jsonschema:"カテゴリ名（音楽、ゲーム、エンターテイメントなど。videos の category と同じ表記）"`
	MinDurationSec     int    `json:"min_duration_seconds,omitempty" jsonschema:"この秒数以上の長さの曲。長さが不明な曲は除く"`
	MaxDurationSec     int    `json:"max_duration_seconds,omitempty" jsonschema:"この秒数以下の長さの曲。長さが不明な曲は除く"`
	AddedFrom          string `json:"added_from,omitempty" jsonschema:"この日以降に追加された曲（YYYY-MM-DD、この日を含む）"`
	AddedUntil         string `json:"added_until,omitempty" jsonschema:"この日より前に追加された曲（YYYY-MM-DD、この日を含まない）"`
	IncludeUnavailable bool   `json:"include_unavailable,omitempty" jsonschema:"非公開・削除になった曲も含める。省略すると観られる曲だけ"`
	Sort               string `json:"sort,omitempty" jsonschema:"並び順。added_desc（新しく追加された順、既定）、added_asc（古い順）、views_desc（再生数の多い順）"`
	Limit              int    `json:"limit,omitempty" jsonschema:"返す件数（1〜100、既定 20）"`
	Offset             int    `json:"offset,omitempty" jsonschema:"先頭から飛ばす件数（続きを取るときに使う）"`
}

type findVideosOut struct {
	Total  int        `json:"total" jsonschema:"条件に合う件数（limit で切る前）"`
	Offset int        `json:"offset"`
	Videos []videoOut `json:"videos"`
}

func (t *tools) findVideos(_ context.Context, _ *mcp.CallToolRequest, in findVideosIn) (*mcp.CallToolResult, findVideosOut, error) {
	sort := application.VideoSort(in.Sort)
	switch sort {
	case "":
		sort = application.SortAddedDesc
	case application.SortAddedDesc, application.SortAddedAsc, application.SortViewsDesc:
	default:
		return nil, findVideosOut{}, fmt.Errorf("sort は added_desc、added_asc、views_desc のいずれかを指定してください: %q", in.Sort)
	}

	limit := in.Limit
	if limit == 0 {
		limit = DEFAULT_VIDEO_LIMIT
	}

	if limit < 1 || limit > MAX_VIDEO_LIMIT || in.Offset < 0 {
		return nil, findVideosOut{}, fmt.Errorf("limit は 1〜%d、offset は 0 以上で指定してください", MAX_VIDEO_LIMIT)
	}

	from, err := t.parseDate("added_from", in.AddedFrom)
	if err != nil {
		return nil, findVideosOut{}, err
	}

	until, err := t.parseDate("added_until", in.AddedUntil)
	if err != nil {
		return nil, findVideosOut{}, err
	}

	var category library.CategoryID
	if in.Category != "" {
		if category, err = t.categoryByName(in.Category); err != nil {
			return nil, findVideosOut{}, err
		}
	}

	if in.MinDurationSec < 0 || in.MaxDurationSec < 0 {
		return nil, findVideosOut{}, errors.New("min_duration_seconds と max_duration_seconds は 0 以上で指定してください")
	}

	page, err := t.query.FindVideos(application.VideoQuery{
		PlaylistID:         library.PlaylistID(in.PlaylistID),
		Title:              in.Title,
		Channel:            in.Channel,
		ChannelID:          library.ChannelID(in.ChannelID),
		Tag:                in.Tag,
		CategoryID:         category,
		MinDuration:        time.Duration(in.MinDurationSec) * time.Second,
		MaxDuration:        time.Duration(in.MaxDurationSec) * time.Second,
		AddedFrom:          from,
		AddedUntil:         until,
		IncludeUnavailable: in.IncludeUnavailable,
		Sort:               sort,
		Offset:             in.Offset,
		Limit:              limit,
	})
	if err != nil {
		return nil, findVideosOut{}, clientError(err)
	}

	videos, err := t.toVideos(page.Videos)
	if err != nil {
		return nil, findVideosOut{}, clientError(err)
	}

	return nil, findVideosOut{Total: page.Total, Offset: in.Offset, Videos: videos}, nil
}

// categoryByName は指定した名前の保存済みのカテゴリの ID を返す。
// ない場合、エラーには設定可能なカテゴリの名前を並べる。
func (t *tools) categoryByName(name string) (library.CategoryID, error) {
	categories, err := t.query.Categories()
	if err != nil {
		return "", clientError(err)
	}

	var names []string

	for id, c := range categories {
		if c.Title == name {
			return id, nil
		}

		if c.Assignable {
			names = append(names, c.Title)
		}
	}

	slices.Sort(names)

	return "", fmt.Errorf("category は次のどれかを指定してください: %s（指定された値: %q）", strings.Join(names, "、"), name)
}

// ---------- playlist_stats ----------

type playlistIn struct {
	PlaylistID string `json:"playlist_id" jsonschema:"プレイリストの ID（list_playlists で確認できる）"`
}

type playlistStatsOut struct {
	Playlist        playlistOut      `json:"playlist"`
	Items           int              `json:"items" jsonschema:"曲数（同じ曲の重複も数える）"`
	Available       int              `json:"available"`
	Private         int              `json:"private"`
	Deleted         int              `json:"deleted"`
	AddedThisMonth  int              `json:"added_this_month"`
	AddedLast30Days int              `json:"added_last_30_days"`
	FirstAddedAt    string           `json:"first_added_at,omitempty"`
	LastAddedAt     string           `json:"last_added_at,omitempty"`
	TopViewed       []rankedVideoOut `json:"top_viewed" jsonschema:"再生数の多い観られる曲（上位 3 曲）"`
	Monthly         []monthCountOut  `json:"monthly" jsonschema:"今月を含む直近 6 か月の月別の追加数（古い順）"`
}

func (t *tools) playlistStats(_ context.Context, _ *mcp.CallToolRequest, in playlistIn) (*mcp.CallToolResult, playlistStatsOut, error) {
	id := library.PlaylistID(in.PlaylistID)

	playlist, err := t.query.Playlist(id)
	if err != nil {
		return nil, playlistStatsOut{}, clientError(err)
	}

	stats, err := t.query.Stats(id)
	if err != nil {
		return nil, playlistStatsOut{}, clientError(err)
	}

	out := playlistStatsOut{
		Playlist:        toPlaylist(playlist),
		Items:           stats.Total,
		Available:       stats.Available,
		Private:         stats.Private,
		Deleted:         stats.Deleted,
		AddedThisMonth:  stats.AddedThisMonth,
		AddedLast30Days: stats.AddedLast30Days,
		FirstAddedAt:    t.timeOut(stats.FirstAddedAt),
		LastAddedAt:     t.timeOut(stats.LastAddedAt),
		TopViewed:       toRanked(stats.TopViewed, id),
		Monthly:         []monthCountOut{},
	}
	for _, m := range stats.Monthly {
		out.Monthly = append(out.Monthly, monthCountOut{Month: m.Month.Format(MONTH_FORMAT), Count: m.Count})
	}

	return nil, out, nil
}

// ---------- playlist_wrapped ----------

type playlistWrappedIn struct {
	PlaylistID string `json:"playlist_id" jsonschema:"プレイリストの ID（list_playlists で確認できる）"`
	Year       int    `json:"year,omitempty" jsonschema:"まとめる年。省略すると今年"`
}

type playlistWrappedOut struct {
	Playlist          playlistOut      `json:"playlist"`
	Year              int              `json:"year"`
	Added             int              `json:"added" jsonschema:"この年に追加された曲数"`
	PreviousYearAdded int              `json:"previous_year_added"`
	BusiestMonth      *monthCountOut   `json:"busiest_month,omitempty" jsonschema:"一番多く追加された月"`
	First             *videoOut        `json:"first,omitempty" jsonschema:"この年に最初に追加された観られる曲"`
	Last              *videoOut        `json:"last,omitempty" jsonschema:"この年に最後に追加された観られる曲"`
	TopViewed         []rankedVideoOut `json:"top_viewed" jsonschema:"この年に追加された観られる曲の再生数ランキング（上位 5 曲）"`
}

func (t *tools) playlistWrapped(_ context.Context, _ *mcp.CallToolRequest, in playlistWrappedIn) (*mcp.CallToolResult, playlistWrappedOut, error) {
	year := in.Year
	if year == 0 {
		year = t.now().In(t.location).Year()
	}

	id := library.PlaylistID(in.PlaylistID)

	playlist, err := t.query.Playlist(id)
	if err != nil {
		return nil, playlistWrappedOut{}, clientError(err)
	}

	w, err := t.query.Wrapped(id, year)
	if err != nil {
		return nil, playlistWrappedOut{}, clientError(err)
	}

	out := playlistWrappedOut{
		Playlist:          toPlaylist(playlist),
		Year:              w.Year,
		Added:             w.Added,
		PreviousYearAdded: w.PreviousYearAdded,
		TopViewed:         toRanked(w.TopViewed, id),
	}
	if w.BusiestMonth.Count > 0 {
		out.BusiestMonth = &monthCountOut{Month: w.BusiestMonth.Month.Format(MONTH_FORMAT), Count: w.BusiestMonth.Count}
	}

	if w.First != nil {
		videos, err := t.toVideos([]*library.PlaylistVideo{w.First, w.Last})
		if err != nil {
			return nil, playlistWrappedOut{}, clientError(err)
		}

		out.First, out.Last = &videos[0], &videos[1]
	}

	return nil, out, nil
}

// ---------- random_videos ----------

type randomVideosIn struct {
	PlaylistID string `json:"playlist_id,omitempty" jsonschema:"プレイリストの ID。省略するとすべてのプレイリストから選ぶ"`
	Count      int    `json:"count,omitempty" jsonschema:"選ぶ曲数（1〜10、既定 1）"`
}

type randomVideosOut struct {
	Videos []videoOut `json:"videos"`
}

func (t *tools) randomVideos(_ context.Context, _ *mcp.CallToolRequest, in randomVideosIn) (*mcp.CallToolResult, randomVideosOut, error) {
	count := in.Count
	if count == 0 {
		count = 1
	}

	if count < 1 || count > MAX_RANDOM_COUNT {
		return nil, randomVideosOut{}, fmt.Errorf("count は 1〜%d で指定してください", MAX_RANDOM_COUNT)
	}

	picked, err := t.query.RandomVideos(library.PlaylistID(in.PlaylistID), count)
	if err != nil {
		return nil, randomVideosOut{}, clientError(err)
	}

	videos, err := t.toVideos(picked)
	if err != nil {
		return nil, randomVideosOut{}, clientError(err)
	}

	return nil, randomVideosOut{Videos: videos}, nil
}

// ---------- video_history ----------

const (
	DEFAULT_HISTORY_DAYS = 30
	MAX_HISTORY_DAYS     = 365
)

type videoHistoryIn struct {
	VideoID string `json:"video_id" jsonschema:"動画の ID"`
	Days    int    `json:"days,omitempty" jsonschema:"さかのぼる日数（1〜365、既定 30）"`
}

type snapshotOut struct {
	RecordedAt string `json:"recorded_at"`
	Title      string `json:"title"`
	Status     string `json:"status" jsonschema:"public（公開）または unlisted（限定公開）"`
	Views      uint64 `json:"views"`
	Likes      uint64 `json:"likes" jsonschema:"高評価数。投稿者が非表示にしていると 0"`
	Comments   uint64 `json:"comments" jsonschema:"コメント数。コメントが無効だと 0"`
}

type videoHistoryOut struct {
	VideoID   string        `json:"video_id"`
	Title     string        `json:"title" jsonschema:"現在の曲名"`
	URL       string        `json:"url"`
	Snapshots []snapshotOut `json:"snapshots" jsonschema:"古い順。記録は 6 時間ごとで、観られない間は記録しない"`
}

func (t *tools) videoHistory(_ context.Context, _ *mcp.CallToolRequest, in videoHistoryIn) (*mcp.CallToolResult, videoHistoryOut, error) {
	days := in.Days
	if days == 0 {
		days = DEFAULT_HISTORY_DAYS
	}

	if days < 1 || days > MAX_HISTORY_DAYS {
		return nil, videoHistoryOut{}, fmt.Errorf("days は 1〜%d で指定してください", MAX_HISTORY_DAYS)
	}

	h, err := t.query.History(library.VideoID(in.VideoID), t.now().AddDate(0, 0, -days))
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return nil, videoHistoryOut{}, errors.New(ERR_VIDEO_NOT_WATCHED)
	}

	if err != nil {
		return nil, videoHistoryOut{}, clientError(err)
	}

	out := videoHistoryOut{
		VideoID:   string(h.Video.YoutubeID),
		Title:     h.Video.Title,
		URL:       fmt.Sprintf(YOUTUBE_VIDEO_URL, h.Video.YoutubeID),
		Snapshots: []snapshotOut{},
	}
	for _, s := range h.Snapshots {
		out.Snapshots = append(out.Snapshots, snapshotOut{
			RecordedAt: t.timeOut(s.RecordedAt),
			Title:      s.Title,
			Status:     string(s.PrivacyStatus),
			Views:      uint64(s.Views),
			Likes:      s.Likes,
			Comments:   s.Comments,
		})
	}

	return nil, out, nil
}
