package usecases

import "discord-playlist-notifier/src/domain"

type PlaylistInteractor struct {
	PlaylistRepository PlaylistRepository
}

func (i *PlaylistInteractor) FindById(id string) (domain.Playlist, error) {
	playlist, err := i.PlaylistRepository.FindById(id)
	if err != nil {
		return domain.Playlist{}, err
	}

	return playlist, nil
}

func (i *PlaylistInteractor) Save(playlist domain.Playlist) error {
	err := i.PlaylistRepository.Save(playlist)
	if err != nil {
		return err
	}

	return nil
}
