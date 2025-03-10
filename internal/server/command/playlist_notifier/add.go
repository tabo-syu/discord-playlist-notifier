package playlist_notifier

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

var addSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "Add a playlist to be notified.",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func (c *PlaylistNotifier) add(guildId string, channelId string, playlistId string) string {
	var message string
	switch c.playlist.Register(guildId, channelId, playlistId) {
	case nil:
		message = fmt.Sprintf("Notification registered!\nhttps://www.youtube.com/playlist?list=%s", playlistId)
	case domain.ErrYouTubePlaylistNotFound:
		message = "Playlist not found. Is it a private playlist?"
	case domain.ErrDBRecordAlreadyCreated:
		message = "This playlist is already registered for notifications."
	case domain.ErrYouTubeGeneralError:
		message = "Please check the YouTube API service status."
	default:
		message = "Error! There is a problem with the system."
	}

	return message
}
