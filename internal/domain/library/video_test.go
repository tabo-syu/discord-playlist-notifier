package library_test

import (
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestVideoAvailable(t *testing.T) {
	tests := map[string]struct {
		video *library.Video
		want  bool
	}{
		"公開":           {&library.Video{Title: "曲", PrivacyStatus: library.PrivacyPublic}, true},
		"限定公開":         {&library.Video{Title: "曲", PrivacyStatus: library.PrivacyUnlisted}, true},
		"非公開":          {&library.Video{Title: "曲", PrivacyStatus: library.PrivacyPrivate}, false},
		"削除":           {&library.Video{Title: "曲", PrivacyStatus: library.PrivacyDeleted}, false},
		"詳細をまだ取得していない": {&library.Video{PrivacyStatus: library.PrivacyPublic}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.video.Available(); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVideoChangePrivacy(t *testing.T) {
	tests := map[string]struct {
		title       string
		from, to    library.PrivacyStatus
		wantChanged bool
		wantHidden  bool
	}{
		"変わらない":           {"曲", library.PrivacyPublic, library.PrivacyPublic, false, false},
		"非公開になった":         {"曲", library.PrivacyPublic, library.PrivacyPrivate, true, true},
		"限定公開から削除になった":    {"曲", library.PrivacyUnlisted, library.PrivacyDeleted, true, true},
		"限定公開になった":        {"曲", library.PrivacyPublic, library.PrivacyUnlisted, true, false},
		"非公開から公開に戻った":     {"曲", library.PrivacyPrivate, library.PrivacyPublic, true, false},
		"非公開から削除になった":     {"曲", library.PrivacyPrivate, library.PrivacyDeleted, true, false},
		"初めて見る動画":         {"", "", library.PrivacyPublic, true, false},
		"初めて見る動画が非公開":     {"", "", library.PrivacyPrivate, true, false},
		"詳細のない動画が非公開になった": {"", library.PrivacyPublic, library.PrivacyPrivate, true, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v := &library.Video{Title: tt.title, PrivacyStatus: tt.from}
			changed, hidden := v.ChangePrivacy(tt.to)
			if changed != tt.wantChanged || hidden != tt.wantHidden {
				t.Errorf("ChangePrivacy(%q) = (%v, %v), want (%v, %v)", tt.to, changed, hidden, tt.wantChanged, tt.wantHidden)
			}
			if v.PrivacyStatus != tt.to {
				t.Errorf("PrivacyStatus = %q, want %q", v.PrivacyStatus, tt.to)
			}
		})
	}
}

func TestVideoApplyDetails(t *testing.T) {
	v := &library.Video{YoutubeID: "aaaaaaaaaaa", PrivacyStatus: library.PrivacyUnlisted}
	v.ApplyDetails(&library.Video{
		YoutubeID:        "bbbbbbbbbbb",
		Title:            "曲",
		Views:            123,
		Likes:            4,
		Comments:         5,
		PrivacyStatus:    library.PrivacyPublic,
		ChannelYoutubeID: "UCxxxx",
		Tags:             []string{"tag"},
		CategoryID:       "10",
	})

	if v.Title != "曲" || v.Views != 123 || v.Likes != 4 || v.Comments != 5 || v.ChannelYoutubeID != "UCxxxx" || len(v.Tags) != 1 || v.CategoryID != "10" {
		t.Errorf("details were not copied: %+v", v)
	}
	// 公開状態と ID はプレイリストのアイテムが持つものなので変えない
	if v.PrivacyStatus != library.PrivacyUnlisted {
		t.Errorf("PrivacyStatus = %q, want %q", v.PrivacyStatus, library.PrivacyUnlisted)
	}
	if v.YoutubeID != "aaaaaaaaaaa" {
		t.Errorf("YoutubeID = %q, want %q", v.YoutubeID, "aaaaaaaaaaa")
	}
}

func milestone(c library.ViewCount) *library.ViewCount {
	return &c
}

func TestVideoInitMilestone(t *testing.T) {
	t.Run("初めて再生数がわかった", func(t *testing.T) {
		v := &library.Video{Views: 2_500_000}
		v.InitMilestone()
		if v.ViewMilestone == nil || *v.ViewMilestone != 1_000_000 {
			t.Errorf("ViewMilestone = %v, want 1000000", v.ViewMilestone)
		}
	})
	t.Run("記録済みの節目は変えない", func(t *testing.T) {
		v := &library.Video{Views: 2_500_000, ViewMilestone: milestone(100_000)}
		v.InitMilestone()
		if *v.ViewMilestone != 100_000 {
			t.Errorf("ViewMilestone = %d, want 100000", *v.ViewMilestone)
		}
	})
}

func TestVideoUpdateMilestone(t *testing.T) {
	tests := map[string]struct {
		recorded *library.ViewCount
		views    library.ViewCount
		want     bool
		wantNext library.ViewCount
	}{
		// 既存の再生数は通知しない
		"初めて再生数がわかった": {nil, 2_500_000, false, 1_000_000},
		"新しい節目に達した":   {milestone(100_000), 1_000_000, true, 1_000_000},
		"節目を 2 つ飛ばした": {milestone(0), 10_000_000, true, 10_000_000},
		"同じ節目のまま":     {milestone(1_000_000), 9_999_999, false, 1_000_000},
		// 再生数が減っても、記録した節目は下げない（同じ節目を再び通知しない）
		"再生数が減った": {milestone(1_000_000), 999_999, false, 1_000_000},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v := &library.Video{Views: tt.views, ViewMilestone: tt.recorded}
			if got := v.UpdateMilestone(); got != tt.want {
				t.Errorf("UpdateMilestone() = %v, want %v", got, tt.want)
			}
			if v.ViewMilestone == nil || *v.ViewMilestone != tt.wantNext {
				t.Errorf("ViewMilestone = %v, want %d", v.ViewMilestone, tt.wantNext)
			}
		})
	}
}
