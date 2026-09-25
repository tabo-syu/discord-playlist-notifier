package service

import (
	"errors"
	"log"
	"sort"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/repository"
)

type PlaylistService struct {
	youtube  repository.YouTubeRepository
	playlist repository.PlaylistRepository
	guild    repository.GuildRepository
}

func NewPlaylistService(y repository.YouTubeRepository, p repository.PlaylistRepository, g repository.GuildRepository) *PlaylistService {
	return &PlaylistService{y, p, g}
}

func (s *PlaylistService) FindAll() ([]*domain.Playlist, error) {
	return s.playlist.FindAll()
}

func (s *PlaylistService) FindByGuild(guildId string) ([]*domain.Playlist, error) {
	return s.playlist.FindByDiscordId(guildId)
}

func (s *PlaylistService) UpdateUpdatedAt(playlist *domain.Playlist, time time.Time) error {
	playlist.UpdatedAt = time

	return s.playlist.Update(playlist)
}

func (s *PlaylistService) Register(guildId string, channelId string, playlistId string) error {
	playlists, err := s.youtube.FindPlaylists(playlistId)
	if errors.Is(err, domain.ErrYouTubePlaylistNotFound) {
		return err
	}
	if err != nil {
		return domain.ErrYouTubeGeneralError
	}

	playlist := playlists[0]

	playlistExist, err := s.playlist.Exist(guildId, playlist.YoutubeID)
	if err != nil {
		return err
	}
	if playlistExist {
		return domain.ErrDBRecordAlreadyCreated
	}

	guild, err := s.guild.GetByDiscordId(guildId)
	if err != nil {
		return err
	}
	playlist.Guild = *guild
	playlist.SendChannelID = channelId

	return s.playlist.Add(playlist)
}

func (s *PlaylistService) Unregister(guildId string, playlistId string) error {
	playlistExist, err := s.playlist.Exist(guildId, playlistId)
	if err != nil {
		return err
	}
	if !playlistExist {
		return domain.ErrDBRecordNotFound
	}

	playlists, err := s.playlist.FindByDiscordId(guildId)
	if err != nil {
		return err
	}

	var target *domain.Playlist
	for _, playlist := range playlists {
		if playlist.YoutubeID == playlistId {
			target = playlist
		}
	}

	err = s.playlist.DeleteAll([]*domain.Playlist{target})
	if err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) SetPickInterval(guildId string, playlistId string, interval string) error {
	playlists, err := s.playlist.FindByDiscordId(guildId)
	if err != nil {
		return err
	}

	for _, playlist := range playlists {
		if playlist.YoutubeID == playlistId {
			return s.playlist.SetPickInterval(playlist, interval)
		}
	}

	return domain.ErrDBRecordNotFound
}

// NewVideos are the videos added to a registered playlist since its last notification.
type NewVideos struct {
	Playlist *domain.Playlist
	Videos   []*PlaylistVideo
}

// GetDiffFromLatest picks, for each registered playlist, the videos added to
// the playlist since its last notification.
func (s *PlaylistService) GetDiffFromLatest(lastPlaylists []*domain.Playlist, snapshots map[string]*PlaylistSnapshot) []*NewVideos {
	var updates []*NewVideos
	for _, last := range lastPlaylists {
		latest, ok := snapshots[last.YoutubeID]
		if !ok {
			// Failed to sync; it has already been logged
			continue
		}
		if latest.Deleted {
			log.Println("Playlist(ID:", last.YoutubeID, ") may have been deleted from YouTube")
			continue
		}

		var added []*PlaylistVideo
		for _, item := range latest.Items {
			// Use !Before instead of After to include videos added at exactly the same time
			if item.AddedAt.Before(last.UpdatedAt) {
				continue
			}
			video, ok := latest.Videos[item.VideoYoutubeID]
			if !ok || !video.Available() {
				continue
			}
			added = append(added, &PlaylistVideo{PlaylistID: last.YoutubeID, Item: item, Video: video})
		}
		if len(added) != 0 {
			// Notify in the order the videos were added, regardless of the playlist order
			sort.SliceStable(added, func(i, j int) bool {
				return added[i].Item.AddedAt.Before(added[j].Item.AddedAt)
			})
			last.Title = latest.Title
			updates = append(updates, &NewVideos{Playlist: last, Videos: added})
		}
	}

	return updates
}
