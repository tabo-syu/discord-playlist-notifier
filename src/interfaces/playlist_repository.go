package interfaces

import (
	"discord-playlist-notifier/src/constant"
	"discord-playlist-notifier/src/domain"
	"discord-playlist-notifier/src/errors"
	"encoding/json"
	"time"
)

type PlaylistRepository struct {
	RedisHandler   RedisHandler
	YouTubeHandler YouTubeHandler
}

func (r *PlaylistRepository) Insert(id string) (*domain.Playlist, error) {
	// プレイリストの情報を取得
	ps := r.YouTubeHandler.Playlists()
	part := []string{constant.PART_SNIPPET}
	listRes, err := ps.List(part).Id(id).Do()
	if err != nil {
		return &domain.Playlist{}, err
	}
	if len(listRes.Items) == 0 {
		return &domain.Playlist{}, errors.ErrNotFoundAtYouTube
	}

	// プレイリストに登録されている動画の情報を取得
	is := r.YouTubeHandler.PlaylistItems()
	part = []string{constant.PART_SNIPPET, constant.PART_CONTENT_DETAILS}
	itemRes, err := is.List(part).PlaylistId(id).MaxResults(10).Do()
	if err != nil {
		return &domain.Playlist{}, err
	}

	var videos []domain.Video
	for _, item := range itemRes.Items {
		t, _ := time.Parse(constant.YOUTUBE_TIME_FORMAT, item.Snippet.PublishedAt)
		v := domain.Video{
			Title:       item.Snippet.Title,
			PublishedAt: t,
			Id:          item.ContentDetails.VideoId,
		}
		videos = append(videos, v)
	}
	lastInsertedStr := itemRes.Items[0].Snippet.PublishedAt
	date, _ := time.Parse(constant.YOUTUBE_TIME_FORMAT, lastInsertedStr)
	playlist := domain.Playlist{
		Id:        id,
		Title:     listRes.Items[0].Snippet.Title,
		Videos:    videos,
		UpdatedAt: date,
	}

	b, _ := json.Marshal(playlist)
	err = r.RedisHandler.Set(playlist.Id, b, 0).Err()
	if err != nil {
		return &domain.Playlist{}, err
	}

	return &playlist, nil
}

func (r *PlaylistRepository) FindById(id string) (*domain.Playlist, error) {
	response, err := r.RedisHandler.Get(id).Bytes()
	if err != nil {
		return &domain.Playlist{}, err
	}

	result := &domain.Playlist{}
	json.Unmarshal(response, result)

	return result, nil
}

func (r *PlaylistRepository) Delete(id string) error {
	return r.RedisHandler.Del(id).Err()
}

func (r *PlaylistRepository) Exists(id string) (bool, error) {
	res, err := r.RedisHandler.Exists(id).Result()
	if err != nil {
		return true, err
	}

	if res == 1 {
		return true, nil
	}

	return false, nil
}
