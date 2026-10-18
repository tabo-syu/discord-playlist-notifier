package notifier

import (
	"fmt"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"github.com/bwmarrin/discordgo"
)

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
