package fake

import (
	"slices"

	"github.com/bwmarrin/discordgo"
)

// Discord は notifier.Sender の偽物。送った Embed を記録する。
type Discord struct {
	// 送信に失敗させる Embed のタイトル
	FailTitles []string

	Sent []*SentEmbed
}

// SentEmbed は送った Embed とその宛先。
type SentEmbed struct {
	ChannelID string
	Embed     *discordgo.MessageEmbed
}

func (d *Discord) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	if slices.Contains(d.FailTitles, embed.Title) {
		return nil, Err
	}
	d.Sent = append(d.Sent, &SentEmbed{ChannelID: channelID, Embed: embed})
	return &discordgo.Message{}, nil
}
