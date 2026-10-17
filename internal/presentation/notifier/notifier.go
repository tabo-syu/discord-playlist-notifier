// Package notifier は Discord のチャンネルに Embed を投稿することで
// application.Notifier を実装する。
package notifier

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"github.com/bwmarrin/discordgo"
)

type notifier struct {
	session *discordgo.Session
	// 日付の表示に使うタイムゾーン
	location *time.Location
}

func New(s *discordgo.Session, loc *time.Location) *notifier {
	return &notifier{s, loc}
}

func (n *notifier) send(channelID subscription.ChannelID, embed *discordgo.MessageEmbed) error {
	_, err := n.session.ChannelMessageSendEmbed(string(channelID), embed)

	return err
}

func (n *notifier) NotifyNewVideos(news *notification.NewVideos, live map[library.VideoID]*library.LiveVideo) error {
	red := color("ff0000")
	sub := news.Subscription

	var errs []error
	for _, added := range news.Videos {
		video := added.Video
		embed := &discordgo.MessageEmbed{
			Color: red,
			Author: &discordgo.MessageEmbedAuthor{
				Name: sub.Title + " に追加されました！",
			},
			Title: video.Title,
			URL:   videoURL(video.YoutubeID, sub.YoutubeID),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "追加日時",
					Value:  added.Item.AddedAt.In(n.location).Format("2006/01/02 15:04:05"),
					Inline: true,
				},
				{
					Name:   "再生回数",
					Value:  separator(views(video, live)),
					Inline: true,
				},
			},
			Image:  &discordgo.MessageEmbedImage{URL: video.Thumbnail()},
			Footer: channelFooter(video, live),
			// UTC の時刻を渡すと、Discord がユーザーに合ったタイムゾーンに変換する
			Timestamp: video.PublishedAt.Format(time.RFC3339),
		}

		// 動画 1 件につき 1 通送り、1 件失敗しても残りを送り続ける
		if err := n.send(sub.SendChannelID, embed); err != nil {
			errs = append(errs, fmt.Errorf("video %s: %w", video.YoutubeID, err))
		}
	}

	return errors.Join(errs...)
}

func videoURL(videoID library.VideoID, playlistID library.PlaylistID) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", videoID, playlistID)
}

// views は送る直前に取得した再生数を優先する。
func views(video *library.Video, live map[library.VideoID]*library.LiveVideo) library.ViewCount {
	if l, ok := live[video.YoutubeID]; ok {
		return l.Views
	}
	return video.Views
}

// channelFooter は動画のチャンネルを表示する。チャンネルは送る直前に
// 取得したときにしかわからない。
func channelFooter(video *library.Video, live map[library.VideoID]*library.LiveVideo) *discordgo.MessageEmbedFooter {
	l, ok := live[video.YoutubeID]
	if !ok {
		return nil
	}
	return &discordgo.MessageEmbedFooter{Text: l.ChannelName, IconURL: l.ChannelIcon}
}

func color(hex string) int {
	color, err := strconv.ParseInt(hex, 16, 0)
	if err != nil {
		// エラーのときは赤にする
		return 0xff0000
	}

	return int(color)
}

func separator(integer library.ViewCount) string {
	// strings.Builder を使って効率よく組み立てる
	var sb strings.Builder
	str := fmt.Sprintf("%d", integer)

	// 必要なカンマの数を計算する
	commas := (len(str) - 1) / 3

	// 再割り当てを避けるためにバッファを事前に確保する
	sb.Grow(len(str) + commas)

	// カンマを入れながら数字を追加する
	for i, char := range str {
		// 先頭を除き、右から 3 桁ごとにカンマを入れる
		if i > 0 && (len(str)-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(char)
	}

	return sb.String()
}
