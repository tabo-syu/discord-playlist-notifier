package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// received はハンドラが受け取ったイベントを、どのハンドラが受け取ったかと一緒に記録する。
type received struct {
	handler string
	events  any
}

func TestPublish(t *testing.T) {
	t.Parallel()

	a := &library.VideoHidden{Video: &library.Video{YoutubeID: "a"}}
	b := &library.VideoHidden{Video: &library.Video{YoutubeID: "b"}}

	tests := []struct {
		name   string
		events []*library.VideoHidden
		want   []received
	}{
		{
			name:   "その型を購読したハンドラだけに、登録した順にまとめて配る",
			events: []*library.VideoHidden{a, b},
			want: []received{
				{handler: "hidden1", events: []*library.VideoHidden{a, b}},
				{handler: "hidden2", events: []*library.VideoHidden{a, b}},
			},
		},
		{
			name:   "イベントがなければ、どのハンドラも呼ばない",
			events: nil,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []received

			p := application.NewEventPublisher()
			application.Subscribe(p, func(events []*library.VideoHidden) { got = append(got, received{"hidden1", events}) })
			application.Subscribe(p, func(events []*library.MilestoneReached) { got = append(got, received{"milestone", events}) })
			application.Subscribe(p, func(events []*library.VideoHidden) { got = append(got, received{"hidden2", events}) })

			application.Publish(p, tt.events)

			assert.Equal(t, tt.want, got)
		})
	}
}
