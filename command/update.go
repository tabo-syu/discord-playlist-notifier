package command

import "github.com/bwmarrin/discordgo"

var update = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "update",
	Description: "プレイリストの通知設定を更新します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
		mentionOption,
	},
}

func updateFunc(playlistId string, needMention bool) string {
	return "update"
}
