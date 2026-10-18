package playlist_notifier

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"
)

const (
	DEFAULT_HISTORY_DAYS = 30
	MAX_HISTORY_DAYS     = 365
)

var (
	minHistoryDays = 1.0

	historyVideoOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "video",
		Description: "動画の URL か ID を入力します。",
		Required:    true,
	}

	historyDaysOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "days",
		Description: "さかのぼる日数（1〜365、省略すると 30）",
		MinValue:    &minHistoryDays,
		MaxValue:    MAX_HISTORY_DAYS,
	}

	historySubCommand = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "history",
		Description: "通知登録されているプレイリストの動画の、再生数と高評価数の推移をグラフで表示します。",
		Options: []*discordgo.ApplicationCommandOption{
			historyVideoOption,
			historyDaysOption,
		},
	}
)

func (c *PlaylistNotifier) history(guildId string, video string, days int) *command.Response {
	videoID, ok := library.ParseVideoID(video)
	if !ok {
		return command.Text("動画の URL か ID を入力してください。")
	}

	subscriptions, err := c.subscriptions.FindByGuild(guild.DiscordID(guildId))
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return command.Text("通知登録されているプレイリストが存在しません。")
	}
	if err != nil {
		return command.Text("エラー！システムに問題があります！")
	}

	var ids []library.PlaylistID
	for _, sub := range subscriptions {
		ids = append(ids, sub.YoutubeID)
	}
	since := time.Now().AddDate(0, 0, -days)
	history, err := c.library.History(ids, videoID, since)
	if errors.Is(err, domain.ErrDBRecordNotFound) {
		return command.Text("通知登録されているプレイリストに、その動画は見つかりませんでした。")
	}
	if err != nil {
		return command.Text("エラー！システムに問題があります！")
	}

	message := formatHistory(history, days)
	if len(history.Snapshots) < 2 {
		return command.Text(message)
	}

	png, err := view.HistoryChart(history.Video.Title, history.Snapshots, c.location)
	if err != nil {
		log.Println("Could not draw history chart:", videoID, "cause:", err)
		return command.Text(message)
	}

	return &command.Response{
		Content: message,
		Files: []*discordgo.File{
			{Name: "history.png", ContentType: "image/png", Reader: bytes.NewReader(png)},
		},
	}
}

func formatHistory(h *library.VideoHistory, days int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "📈 **%s** の推移（直近 %d 日）\n", h.Video.Title, days)

	if len(h.Snapshots) < 2 {
		sb.WriteString("まだ記録が足りません。再生数などは 6 時間ごとに記録しています。\n")
	} else {
		first, last := h.Snapshots[0], h.Snapshots[len(h.Snapshots)-1]
		fmt.Fprintf(&sb, "再生数: %s\n", change(uint64(first.Views), uint64(last.Views)))
		fmt.Fprintf(&sb, "高評価: %s\n", change(first.Likes, last.Likes))
		fmt.Fprintf(&sb, "コメント: %s\n", change(first.Comments, last.Comments))
	}
	fmt.Fprintf(&sb, "https://www.youtube.com/watch?v=%s", h.Video.YoutubeID)

	return sb.String()
}

// change は before から after への値の変化を整形する。例: 1,000 → 1,200（+200）
func change(before, after uint64) string {
	sign := "+"
	diff := after - before
	if after < before {
		sign = "-"
		diff = before - after
	}

	return fmt.Sprintf("%s → %s（%s%s）", view.Number(before), view.Number(after), sign, view.Number(diff))
}
