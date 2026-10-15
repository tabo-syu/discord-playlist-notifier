package playlist_notifier

import (
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"

	"github.com/bwmarrin/discordgo"
)

type PlaylistNotifier struct {
	command       *discordgo.ApplicationCommand
	subscriptions *application.SubscriptionService
	library       *application.LibraryService
	// Time zone used to show dates and to group them by month or year
	location *time.Location
}

var (
	playlistIdOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "playlist-id",
		Description: "YouTube のプレイリストページの URL 末尾に付く ID を入力します。",
		Required:    true,
	}

	optionalPlaylistIdOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        playlistIdOption.Name,
		Description: "見るプレイリストの ID（省略するとすべて）",
	}
)

func NewPlaylistNotifier(s *application.SubscriptionService, l *application.LibraryService, loc *time.Location) *PlaylistNotifier {
	return &PlaylistNotifier{
		&discordgo.ApplicationCommand{
			Name:        "playlist-notifier",
			Description: "テキストチャンネルに YouTube のプレイリストの更新を通知します。",
			Options: []*discordgo.ApplicationCommandOption{
				listSubCommand,
				addSubCommand,
				deleteSubCommand,
				sourceSubCommand,
				summarySubCommand,
				trendingSubCommand,
				wrappedSubCommand,
				randomSubCommand,
				pickSubCommand,
				historySubCommand,
			},
		},
		s,
		l,
		loc,
	}
}

func (c *PlaylistNotifier) GetCommand() *discordgo.ApplicationCommand {
	return c.command
}

func (c *PlaylistNotifier) SetCommand(cmd *discordgo.ApplicationCommand) {
	c.command = cmd
}

func (c *PlaylistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) *command.Response {
	// Check if Options array is empty
	if len(data.Options) == 0 {
		return command.Text("Error: No subcommand provided. Please use one of the available subcommands.")
	}

	subcommand := data.Options[0]
	// Verify that the option is a subcommand
	if subcommand.Type != discordgo.ApplicationCommandOptionSubCommand {
		return command.Text("Error: Invalid command format. Please use one of the available subcommands.")
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
			return command.Text("Error: Playlist ID is required.")
		}

		playlistId := playlistOption.StringValue()
		if playlistId == "" {
			return command.Text("Error: Playlist ID cannot be empty.")
		}

		message = c.add(guildId, channelId, playlistId)
	case deleteSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		// Check if the required option exists
		playlistOption, exists := options[playlistIdOption.Name]
		if !exists {
			return command.Text("Error: Playlist ID is required.")
		}

		playlistId := playlistOption.StringValue()
		if playlistId == "" {
			return command.Text("Error: Playlist ID cannot be empty.")
		}

		message = c.delete(guildId, playlistId)
	case sourceSubCommand.Name:
		message = c.source()
	case summarySubCommand.Name:
		playlistId := ""
		options := command.ParseArguments(subcommand.Options)
		if playlistOption, exists := options[optionalPlaylistIdOption.Name]; exists {
			playlistId = playlistOption.StringValue()
		}

		message = c.summary(guildId, playlistId)
	case trendingSubCommand.Name:
		playlistId := ""
		options := command.ParseArguments(subcommand.Options)
		if playlistOption, exists := options[optionalPlaylistIdOption.Name]; exists {
			playlistId = playlistOption.StringValue()
		}
		days := DEFAULT_TRENDING_DAYS
		if daysOption, exists := options[trendingDaysOption.Name]; exists {
			days = int(daysOption.IntValue())
		}
		if days < 1 || days > MAX_TRENDING_DAYS {
			return command.Text("Error: days must be between 1 and 365.")
		}

		return c.trending(guildId, playlistId, days)
	case wrappedSubCommand.Name:
		year := 0
		options := command.ParseArguments(subcommand.Options)
		if yearOption, exists := options[wrappedYearOption.Name]; exists {
			year = int(yearOption.IntValue())
		}

		message = c.wrapped(guildId, year)
	case randomSubCommand.Name:
		count := 1
		options := command.ParseArguments(subcommand.Options)
		if countOption, exists := options[randomCountOption.Name]; exists {
			count = int(countOption.IntValue())
		}
		if count < 1 || count > MAX_RANDOM_COUNT {
			return command.Text("Error: count must be between 1 and 5.")
		}

		message = c.random(guildId, count)
	case historySubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		videoOption, exists := options[historyVideoOption.Name]
		if !exists || videoOption.StringValue() == "" {
			return command.Text("Error: Video is required.")
		}
		days := DEFAULT_HISTORY_DAYS
		if daysOption, exists := options[historyDaysOption.Name]; exists {
			days = int(daysOption.IntValue())
		}
		if days < 1 || days > MAX_HISTORY_DAYS {
			return command.Text("Error: days must be between 1 and 365.")
		}

		return c.history(guildId, videoOption.StringValue(), days)
	case pickSubCommand.Name:
		options := command.ParseArguments(subcommand.Options)
		playlistOption, exists := options[playlistIdOption.Name]
		if !exists || playlistOption.StringValue() == "" {
			return command.Text("Error: Playlist ID is required.")
		}
		intervalOption, exists := options[pickIntervalOption.Name]
		if !exists {
			return command.Text("Error: interval is required.")
		}

		message = c.pick(guildId, playlistOption.StringValue(), intervalOption.StringValue())
	default:
		message = "Error: Unknown subcommand. Please use one of the available subcommands."
	}

	return command.Text(message)
}
