package usecases

import (
	"discord-playlist-notifier/src/domain"
	"time"
)

type PlaylistInteractor struct {
	PlaylistRepository PlaylistRepository
}

func (i *PlaylistInteractor) Register(playlistId string) (*domain.Playlist, error) {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil {
		return &domain.Playlist{}, err
	}

	if exists {
		return &domain.Playlist{}, nil
	}

	res, err := i.PlaylistRepository.Insert(playlistId)
	if err != nil {
		return &domain.Playlist{}, err
	}

	return res, nil
}

func (i *PlaylistInteractor) AddedVideosSince(playlistId string, since time.Time) *[]domain.Video {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil || !exists {
		return &[]domain.Video{}
	}

	playlist, err := i.PlaylistRepository.FindById(playlistId)
	if err != nil {
		return &[]domain.Video{}
	}

	var addedVideos []domain.Video
	for _, v := range playlist.Videos {
		if v.PublishedAt.After(since) {
			addedVideos = append(addedVideos, v)
		}
	}

	return &addedVideos
}

func (i *PlaylistInteractor) Update(playlistId string) (*domain.Playlist, error) {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil {
		return &domain.Playlist{}, err
	}

	if !exists {
		return &domain.Playlist{}, nil
	}

	res, err := i.PlaylistRepository.Insert(playlistId)
	if err != nil {
		return &domain.Playlist{}, err
	}

	return res, nil
}

func (i *PlaylistInteractor) Delete(playlistId string) error {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return i.PlaylistRepository.Delete(playlistId)
}
