package usecases

import (
	"discord-playlist-notifier/src/domain"
	"discord-playlist-notifier/src/errors"
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
		return &domain.Playlist{}, errors.ErrAlreadyRegisteredAtDatabase
	}

	res, err := i.PlaylistRepository.Insert(playlistId)
	if err != nil {
		return &domain.Playlist{}, err
	}

	return res, nil
}

func (i *PlaylistInteractor) AddedVideosSince(playlistId string) (*[]domain.Video, error) {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil {
		return &[]domain.Video{}, err
	}
	if !exists {
		return &[]domain.Video{}, errors.ErrNotFoundAtDatabase
	}

	pp, err := i.PlaylistRepository.FindById(playlistId)
	if err != nil {
		return &[]domain.Video{}, err
	}

	cp, err := i.PlaylistRepository.Insert(playlistId)
	if err != nil {
		return &[]domain.Video{}, err
	}

	var addedVideos []domain.Video
	for _, v := range cp.Videos {
		if v.PublishedAt.After(pp.UpdatedAt) {
			addedVideos = append(addedVideos, v)
		}
	}

	return &addedVideos, nil
}

func (i *PlaylistInteractor) Delete(playlistId string) error {
	exists, err := i.PlaylistRepository.Exists(playlistId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrNotFoundAtDatabase
	}

	return i.PlaylistRepository.Delete(playlistId)
}
