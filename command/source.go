package command

import (
	"discord-playlist-notifier/repository"

	"github.com/bwmarrin/discordgo"
)

var sourceSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "source",
	Description: "この Bot のリポジトリへのリンクを表示します。",
}

func source(repository repository.YouTubeRepository) string {
	return "https://github.com/tabo-syu/discord-playlist-notifier"
}
