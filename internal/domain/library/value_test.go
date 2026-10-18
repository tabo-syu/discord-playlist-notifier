package library_test

import (
	"testing"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestViewCountMilestone(t *testing.T) {
	tests := []struct {
		views library.ViewCount
		want  library.ViewCount
	}{
		{0, 0},
		{99_999, 0},
		{100_000, 100_000},
		{999_999, 100_000},
		{1_000_000, 1_000_000},
		{12_345_678, 10_000_000},
		{100_000_000, 100_000_000},
		// 最も高い節目を超えても、それより上の節目はない
		{5_000_000_000, 100_000_000},
	}
	for _, tt := range tests {
		if got := tt.views.Milestone(); got != tt.want {
			t.Errorf("ViewCount(%d).Milestone() = %d, want %d", tt.views, got, tt.want)
		}
	}
}

func TestPrivacyStatus(t *testing.T) {
	tests := []struct {
		status    library.PrivacyStatus
		watchable bool
		hidden    bool
	}{
		{library.PrivacyPublic, true, false},
		{library.PrivacyUnlisted, true, false},
		{library.PrivacyPrivate, false, true},
		{library.PrivacyDeleted, false, true},
		// 初めて見る動画など、状態がわからないときはどちらでもない
		{"", false, false},
	}
	for _, tt := range tests {
		if got := tt.status.Watchable(); got != tt.watchable {
			t.Errorf("%q.Watchable() = %v, want %v", tt.status, got, tt.watchable)
		}
		if got := tt.status.Hidden(); got != tt.hidden {
			t.Errorf("%q.Hidden() = %v, want %v", tt.status, got, tt.hidden)
		}
	}
}

func TestParseVideoID(t *testing.T) {
	const id = "dQw4w9WgXcQ"
	tests := map[string]struct {
		in   string
		want library.VideoID
		ok   bool
	}{
		"動画 ID":         {id, id, true},
		"前後の空白":         {"  " + id + "\n", id, true},
		"ハイフンとアンダースコア":  {"a-b_c-d_e-f", "a-b_c-d_e-f", true},
		"watch の URL":   {"https://www.youtube.com/watch?v=" + id, id, true},
		"他のクエリを含む URL":  {"https://www.youtube.com/watch?list=PLxxxx&v=" + id + "&t=42s", id, true},
		"YouTube Music": {"https://music.youtube.com/watch?v=" + id, id, true},
		"短縮 URL":        {"https://youtu.be/" + id + "?si=abc", id, true},
		"ショートの URL":     {"https://www.youtube.com/shorts/" + id, id, true},
		"末尾のスラッシュ":      {"https://youtu.be/" + id + "/", id, true},
		"空文字列":          {"", "", false},
		"短すぎる ID":       {"dQw4w9WgXc", "", false},
		"長すぎる ID":       {"dQw4w9WgXcQQ", "", false},
		"使えない文字":        {"dQw4w9WgXc!", "", false},
		"v が不正な URL":    {"https://www.youtube.com/watch?v=short", "", false},
		"動画を指さない URL":   {"https://www.youtube.com/", "", false},
		"プレイリストの URL":   {"https://www.youtube.com/playlist?list=PLxxxxxxxxxxxxxxxx", "", false},
		"パースできない URL":   {"%zz", "", false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := library.ParseVideoID(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Errorf("ParseVideoID(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}
