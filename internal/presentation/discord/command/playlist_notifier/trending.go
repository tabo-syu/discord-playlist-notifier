package playlist_notifier

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

const (
	DEFAULT_TRENDING_DAYS = 30
	MAX_TRENDING_DAYS     = 365
	// Discord はこれより多い添付ファイルを持つメッセージを拒否する
	MAX_ATTACHMENTS = 10
)

var (
	minTrendingDays = 1.0

	trendingDaysOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "days",
		Description: "さかのぼる日数（1〜365、省略すると 30）",
		MinValue:    &minTrendingDays,
		MaxValue:    MAX_TRENDING_DAYS,
	}

	trendingSubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "trending",
		Description: "通知登録されているプレイリストで、再生数が伸びている動画とその推移をグラフで表示します。",
		Options: []*discordgo.ApplicationCommandOption{
			optionalPlaylistIdOption,
			trendingDaysOption,
		},
	}
)

func (c *PlaylistNotifier) trending(guildId string, playlistId string, days int) *command.Response {
	subscriptions, err := c.subscriptions.FindByGuild(guild.DiscordID(guildId))
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return command.Text("通知登録されているプレイリストが存在しません。")
	}

	if err != nil {
		return command.Text("エラー！システムに問題があります！")
	}

	since := c.now().AddDate(0, 0, -days)

	var (
		sections []string
		files    []*discordgo.File
	)

	for _, sub := range subscriptions {
		if playlistId != "" && sub.YoutubeID != library.PlaylistID(playlistId) {
			continue
		}

		trending, err := c.library.Trending(sub.YoutubeID, since)
		if err != nil {
			return command.Text("エラー！システムに問題があります！")
		}

		heading := fmt.Sprintf("🔥 **%s** の急上昇", sub.Title)
		if len(trending.Videos) > 0 && len(files) < MAX_ATTACHMENTS {
			png, err := view.TrendingChart(sub.Title+" の急上昇", trending.Videos, c.location)
			if err != nil {
				log.Println("Could not draw trending chart:", sub.YoutubeID, "cause:", err)
			} else {
				files = append(files, &discordgo.File{
					Name:        fmt.Sprintf("trending-%d.png", len(files)+1),
					ContentType: "image/png",
					Reader:      bytes.NewReader(png),
				})
				heading += fmt.Sprintf("（グラフ %d）", len(files))
			}
		}

		sections = append(sections, heading+"\n"+view.Trending(trending, c.location))
	}

	if len(sections) == 0 {
		return command.Text("通知登録されていないプレイリストです。")
	}

	return &command.Response{
		Content: view.Truncate(strings.Join(sections, "\n"), view.MAX_MESSAGE_LENGTH),
		Files:   files,
	}
}
