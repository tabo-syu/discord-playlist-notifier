package command

import (
	"github.com/bwmarrin/discordgo"
)

// Response はコマンドが返す応答。ハンドラは組み立てるだけで、
// 送るのはサーバーの役目。
type Response struct {
	Content string
	Files   []*discordgo.File
}

// Text は本文だけの応答。
func Text(content string) *Response {
	return &Response{Content: content}
}

type HandleType func(request *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) *Response

type Command interface {
	Handle(request *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) *Response
	GetCommand() *discordgo.ApplicationCommand
	SetCommand(*discordgo.ApplicationCommand)
}

type Options map[string]*discordgo.ApplicationCommandInteractionDataOption

func ParseArguments(args []*discordgo.ApplicationCommandInteractionDataOption) Options {
	options := make(Options, len(args))
	for _, arg := range args {
		options[arg.Name] = arg
	}

	return options
}
