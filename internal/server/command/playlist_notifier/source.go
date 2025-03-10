package playlist_notifier

import (
	"github.com/bwmarrin/discordgo"
)

var sourceSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "source",
	Description: "Display a link to this bot's repository.",
}

func (c *PlaylistNotifier) source() string {
	return "https://github.com/tabo-syu/discord-playlist-notifier"
}
