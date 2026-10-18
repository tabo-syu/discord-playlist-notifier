package notification_test

import (
	"slices"
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindNewVideos(t *testing.T) {
	snapshots := map[library.PlaylistID]*library.Snapshot{
		"PL1": {
			YoutubeID: "PL1",
			Title:     "現在のタイトル",
			// プレイリストの並びは追加された順とは限らない
			Items: []*library.PlaylistItem{
				item("latest", at(3)),
				item("exact", at(0)),
				item("old", at(-1)),
				item("private", at(1)),
				item("no-details", at(1)),
				item("missing", at(1)),
			},
			Videos: map[library.VideoID]*library.Video{
				"latest":     video("latest", library.PrivacyPublic),
				"exact":      video("exact", library.PrivacyUnlisted),
				"old":        video("old", library.PrivacyPublic),
				"private":    video("private", library.PrivacyPrivate),
				"no-details": {YoutubeID: "no-details", PrivacyStatus: library.PrivacyPublic},
			},
		},
		"PL-deleted": {
			YoutubeID: "PL-deleted",
			Deleted:   true,
			Items:     []*library.PlaylistItem{item("latest", at(3))},
			Videos:    map[library.VideoID]*library.Video{"latest": video("latest", library.PrivacyPublic)},
		},
	}
	subscriptions := []*subscription.Subscription{
		sub("PL1", "ch1", at(0)),
		// 同期に失敗したプレイリスト
		sub("PL-failed", "ch1", at(0)),
		sub("PL-deleted", "ch1", at(0)),
		// 同じプレイリストでも、通知設定ごとの基準時刻で判定する
		sub("PL1", "ch2", at(2)),
		sub("PL1", "ch3", at(4)),
	}

	got := notification.FindNewVideos(subscriptions, snapshots)

	type result struct {
		channel subscription.ChannelID
		title   string
		videos  []library.VideoID
	}
	var results []result
	for _, n := range got {
		r := result{channel: n.Subscription.SendChannelID, title: n.PlaylistTitle}
		for _, v := range n.Videos {
			r.videos = append(r.videos, v.Video.YoutubeID)
			if v.PlaylistID != n.Subscription.YoutubeID || v.Item.VideoYoutubeID != v.Video.YoutubeID {
				t.Errorf("%s: PlaylistVideo = %+v, want the item and video of %s", r.channel, v, n.Subscription.YoutubeID)
			}
		}
		results = append(results, r)
	}
	want := []result{
		// 基準時刻ちょうどに追加された動画を含め、追加された順に並べる
		{"ch1", "現在のタイトル", []library.VideoID{"exact", "latest"}},
		{"ch2", "現在のタイトル", []library.VideoID{"latest"}},
	}
	if !slices.EqualFunc(results, want, func(a, b result) bool {
		return a.channel == b.channel && a.title == b.title && slices.Equal(a.videos, b.videos)
	}) {
		t.Errorf("FindNewVideos() = %+v, want %+v", results, want)
	}
}

func TestFindNewVideosNothing(t *testing.T) {
	if got := notification.FindNewVideos(nil, nil); len(got) != 0 {
		t.Errorf("FindNewVideos(nil, nil) = %v, want empty", got)
	}
}
