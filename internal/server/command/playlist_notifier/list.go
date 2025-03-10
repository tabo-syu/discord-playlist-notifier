package playlist_notifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"

	"github.com/bwmarrin/discordgo"
)

var listSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "list",
	Description: "List playlists that are being notified.",
}

func (c *PlaylistNotifier) list(guildId string) string {
	playlists, err := c.playlist.FindByGuild(guildId)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return "No playlists are registered for notifications."
	}
	if err != nil {
		return "Error! There is a problem with the system."
	}

	var sb strings.Builder
	sb.WriteString("List of playlists registered for notifications:\n")
	for _, playlist := range playlists {
		sb.WriteString(fmt.Sprintf("https://www.youtube.com/playlist?list=%s\n", playlist.YoutubeID))
	}

	return sb.String()
}
