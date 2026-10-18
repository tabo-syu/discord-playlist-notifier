package notification_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindMilestones(t *testing.T) {
	million := &library.Video{YoutubeID: "million", Views: 1_234_567}
	hundredK := &library.Video{YoutubeID: "hundred-k", Views: 150_000}
	videos := []*library.Video{million, hundredK}
	containing := map[library.VideoID][]library.PlaylistID{
		"million":   {"PL1", "PL2"},
		"hundred-k": {"PL2"},
	}
	subscriptions := []*subscription.Subscription{
		sub("PL1", "ch1", at(0)),
		// ch1 には PL1 で通知するので、PL2 でもう一度通知しない
		sub("PL2", "ch1", at(0)),
		sub("PL2", "ch2", at(0)),
		// 動画を含まないプレイリスト
		sub("PL3", "ch3", at(0)),
	}

	got := notification.FindMilestones(subscriptions, videos, containing)

	var notices []string
	for _, n := range got {
		notices = append(notices, fmt.Sprintf("%s %s %s %d", n.ChannelID, n.PlaylistID, n.Video.YoutubeID, n.Milestone))
	}
	want := []string{
		"ch1 PL1 million 1000000",
		"ch2 PL2 million 1000000",
		"ch1 PL2 hundred-k 100000",
		"ch2 PL2 hundred-k 100000",
	}
	if !slices.Equal(notices, want) {
		t.Errorf("FindMilestones() =\n%v\nwant\n%v", notices, want)
	}
}
