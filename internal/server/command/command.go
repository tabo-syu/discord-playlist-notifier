package command

import (
	"github.com/bwmarrin/discordgo"
)

type HandleType func(request *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string

type Command interface {
	Handle(request *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string
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
