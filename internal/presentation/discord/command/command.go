package command

import (
	"github.com/bwmarrin/discordgo"
)

// Response is what a command replies with. Handlers only build it; sending
// it is the job of the server.
type Response struct {
	Content string
	Files   []*discordgo.File
}

// Text is a response with a message only.
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
