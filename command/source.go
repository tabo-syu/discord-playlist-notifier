package command

import "github.com/bwmarrin/discordgo"

var source = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "source",
	Description: "この Bot のリポジトリへのリンクを表示します。",
}

func sourceFunc() string {
	return "source"
}
