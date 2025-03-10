package playlist_notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/server/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"

	"github.com/bwmarrin/discordgo"
)

type PlaylistNotifier struct {
	command  *discordgo.ApplicationCommand
	playlist *service.PlaylistService
}

var (
	playlistIdOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "playlist-id",
		Description: "Enter the ID that appears at the end of the YouTube playlist page URL.",
		Required:    true,
	}
)

func NewPlaylistNotifier(p *service.PlaylistService) *PlaylistNotifier {
	return &PlaylistNotifier{
		&discordgo.ApplicationCommand{
			Name:        "playlist-notifier",
			Description: "Sends notifications to text channels when YouTube playlists are updated.",
			Options: []*discordgo.ApplicationCommandOption{
				listSubCommand,
				addSubCommand,
				deleteSubCommand,
				sourceSubCommand,
			},
		},
		p,
	}
}

func (c *PlaylistNotifier) GetCommand() *discordgo.ApplicationCommand {
	return c.command
}

func (c *PlaylistNotifier) SetCommand(cmd *discordgo.ApplicationCommand) {
	c.command = cmd
}

func (c *PlaylistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string {
	// Check if Options array is empty
	if len(data.Options) == 0 {
		return "Error: No subcommand provided. Please use one of the available subcommands."
	}

	subcommand := data.Options[0]
	// Verify that the option is a subcommand
	if subcommand.Type != discordgo.ApplicationCommandOptionSubCommand {
		return "Error: Invalid command format. Please use one of the available subcommands."
	}

	var message string
	switch subcommand.Name {
	case listSubCommand.Name:
		message = c.list(guildId)
	case addSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		// Check if the required option exists
		playlistOption, exists := options[playlistIdOption.Name]
		if !exists {
			return "Error: Playlist ID is required."
		}
		
		playlistId := playlistOption.StringValue()
		if playlistId == "" {
			return "Error: Playlist ID cannot be empty."
		}
		
		message = c.add(guildId, channelId, playlistId)
	case deleteSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		// Check if the required option exists
		playlistOption, exists := options[playlistIdOption.Name]
		if !exists {
			return "Error: Playlist ID is required."
		}
		
		playlistId := playlistOption.StringValue()
		if playlistId == "" {
			return "Error: Playlist ID cannot be empty."
		}
		
		message = c.delete(guildId, playlistId)
	case sourceSubCommand.Name:
		message = c.source()
	default:
		message = "Error: Unknown subcommand. Please use one of the available subcommands."
	}

	return message
}
