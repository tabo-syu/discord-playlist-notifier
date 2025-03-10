package playlist_notifier

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
)

var deleteSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "delete",
	Description: "Delete a playlist from notifications.",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
	},
}

func (c *PlaylistNotifier) delete(guildId string, playlistId string) string {
	var message string
	switch c.playlist.Unregister(guildId, playlistId) {
	case nil:
		message = "The specified playlist has been deleted."
	case domain.ErrDBRecordNotFound:
		message = "This playlist is not registered for notifications."
	default:
		message = "Error! There is a problem with the system."
	}

	return message
}
