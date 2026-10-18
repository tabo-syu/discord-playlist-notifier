package notification_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/notification"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
)

func TestFindHiddenVideos(t *testing.T) {
	shared := video("shared", library.PrivacyPrivate)
	own := video("own", library.PrivacyDeleted)
	snapshots := map[library.PlaylistID]*library.Snapshot{
		// 動画は 1 つしか保存しないので、変化は最初に同期したプレイリストでしか報告されない
		"PL1": {
			Title:  "PL1 のタイトル",
			Items:  []*library.PlaylistItem{item("shared", at(0)), item("own", at(0))},
			Hidden: []*library.Video{shared},
		},
		"PL2": {
			Title: "PL2 のタイトル",
			Items: []*library.PlaylistItem{item("shared", at(0))},
		},
		"PL3": {
			Title:  "PL3 のタイトル",
			Items:  []*library.PlaylistItem{item("own", at(0))},
			Hidden: []*library.Video{own},
		},
		"PL4": {
			Title: "PL4 のタイトル",
			Items: []*library.PlaylistItem{item("other", at(0))},
		},
	}
	subscriptions := []*subscription.Subscription{
		sub("PL1", "ch1", at(0)),
		sub("PL2", "ch2", at(0)),
		// ch1 には PL1 で通知するので、同じ動画を PL2 でもう一度通知しない
		sub("PL2", "ch1", at(0)),
		sub("PL3", "ch3", at(0)),
		// 動画を含まないプレイリスト
		sub("PL4", "ch4", at(0)),
		// 同期に失敗したプレイリスト
		sub("PL-failed", "ch5", at(0)),
	}

	got := notification.FindHiddenVideos(subscriptions, snapshots)

	var notices []string
	for _, n := range got {
		notices = append(notices, fmt.Sprintf("%s %s %s %s", n.ChannelID, n.PlaylistID, n.PlaylistTitle, n.Video.YoutubeID))
	}
	// 動画の順はプレイリストのマップの順に依存するので、並べ替えて比べる
	slices.Sort(notices)
	want := []string{
		"ch1 PL1 PL1 のタイトル own",
		"ch1 PL1 PL1 のタイトル shared",
		"ch2 PL2 PL2 のタイトル shared",
		"ch3 PL3 PL3 のタイトル own",
	}
	if !slices.Equal(notices, want) {
		t.Errorf("FindHiddenVideos() =\n%v\nwant\n%v", notices, want)
	}
}

func TestFindHiddenVideosNothingHidden(t *testing.T) {
	snapshots := map[library.PlaylistID]*library.Snapshot{
		"PL1": {Items: []*library.PlaylistItem{item("a", at(0))}},
	}
	subscriptions := []*subscription.Subscription{sub("PL1", "ch1", at(0))}

	if got := notification.FindHiddenVideos(subscriptions, snapshots); got != nil {
		t.Errorf("FindHiddenVideos() = %v, want nil", got)
	}
}
