package playlist_notifier

import (
	"github.com/bwmarrin/discordgo"
)

var updateSubCommand = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "update",
	Description: "プレイリストの通知設定を更新します。",
	Options: []*discordgo.ApplicationCommandOption{
		playlistIdOption,
		mentionOption,
	},
}

func (c *playlistNotifier) update(playlistId string, needMention bool) string {
	return "update"
}
