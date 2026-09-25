package playlist_notifier

import (
	"errors"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/service"
	"github.com/tabo-syu/discord-playlist-notifier/internal/view"
)

var (
	// YouTube playlists did not exist before this
	minWrappedYear = 2005.0

	wrappedYearOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "year",
		Description: "まとめる年（省略すると今年）",
		MinValue:    &minWrappedYear,
	}

	wrappedSubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "wrapped",
		Description: "通知登録されているプレイリストの 1 年間のまとめを表示します。",
		Options: []*discordgo.ApplicationCommandOption{
			wrappedYearOption,
		},
	}
)

func (c *PlaylistNotifier) wrapped(guildId string, year int) string {
	playlists, err := c.playlist.FindByGuild(guildId)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return "通知登録されているプレイリストが存在しません。"
	}
	if err != nil {
		return "エラー！システムに問題があります！"
	}
	if year == 0 {
		year = time.Now().In(c.location).Year()
	}

	var sections []string
	for _, playlist := range playlists {
		contents, err := c.library.Contents([]string{playlist.YoutubeID})
		if err != nil {
			return "エラー！システムに問題があります！"
		}
		w := service.ComputeWrapped(contents, year, c.location)
		sections = append(sections, "**"+view.WrappedTitle(playlist.Title, year)+"**\n"+view.Wrapped(w, c.location))
	}

	return view.Truncate(strings.Join(sections, "\n"), view.MAX_MESSAGE_LENGTH)
}
