package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestChannelsOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		details []*library.VideoDetails
		want    []*library.Channel
	}{
		{
			name: "動画の詳細から、投稿したチャンネルを最初に現れた順に返す",
			details: []*library.VideoDetails{
				{VideoID: "a", ChannelID: "UC2", ChannelTitle: "チャンネル 2"},
				{VideoID: "b", ChannelID: "UC1", ChannelTitle: "チャンネル 1"},
			},
			want: []*library.Channel{
				{YoutubeID: "UC2", Title: "チャンネル 2"},
				{YoutubeID: "UC1", Title: "チャンネル 1"},
			},
		},
		{
			name: "同じチャンネルの動画が複数あれば 1 つにまとめ、後の動画の名前を使う",
			details: []*library.VideoDetails{
				{VideoID: "a", ChannelID: "UC", ChannelTitle: "古い名前"},
				{VideoID: "b", ChannelID: "UC", ChannelTitle: "新しい名前"},
			},
			want: []*library.Channel{
				{YoutubeID: "UC", Title: "新しい名前"},
			},
		},
		{
			name: "チャンネルがわからない動画は除く",
			details: []*library.VideoDetails{
				{VideoID: "a"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := library.ChannelsOf(tt.details)

			assert.Equal(t, tt.want, got)
		})
	}
}
