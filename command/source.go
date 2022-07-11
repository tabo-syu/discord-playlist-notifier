package command

import (
	"github.com/bwmarrin/discordgo"
)

var sourceSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "source",
	Description: "この Bot のリポジトリへのリンクを表示します。",
}

func source() string {
	return "https://github.com/tabo-syu/discord-playlist-notifier"
}
