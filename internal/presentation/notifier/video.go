package notifier

import (
	"fmt"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"github.com/bwmarrin/discordgo"
)

// videoNotice は動画を 1 本知らせる通知のうち、通知の種類ごとに違う部分。
// タイトル、URL、サムネイル、投稿者は sendVideo が動画から組み立てる。
type videoNotice struct {
	Color  int
	Author string
	Fields []field
	Image  imageSize
	// ゼロなら出さない
	Timestamp time.Time
}

// field は Embed のフィールド。全て横に並べる。
type field struct{ Name, Value string }

// imageSize はサムネイルの出し方。
type imageSize int

const (
	// Embed の下に大きく出す
	LARGE imageSize = iota
	// Embed の右上に小さく出す
	SMALL
)

// sendVideo は動画を知らせる Embed を送る。投稿者は live に動画があるときだけ出す。
func (n *notifier) sendVideo(channelID subscription.ChannelID, v *library.Video, playlistID library.PlaylistID, live map[library.VideoID]*library.LiveVideo, notice videoNotice) error {
	embed := &discordgo.MessageEmbed{
		Color:  notice.Color,
		Author: &discordgo.MessageEmbedAuthor{Name: notice.Author},
		Title:  v.Title,
		URL:    videoURL(v.YoutubeID, playlistID),
		Footer: channelFooter(v, live),
	}
	for _, f := range notice.Fields {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: f.Name, Value: f.Value, Inline: true})
	}
	switch notice.Image {
	case LARGE:
		embed.Image = &discordgo.MessageEmbedImage{URL: v.Thumbnail()}
	case SMALL:
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: v.Thumbnail()}
	}
	if !notice.Timestamp.IsZero() {
		// UTC の時刻を渡すと、Discord がユーザーに合ったタイムゾーンに変換する
		embed.Timestamp = notice.Timestamp.Format(time.RFC3339)
	}

	_, err := n.session.ChannelMessageSendEmbed(string(channelID), embed)

	return err
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
