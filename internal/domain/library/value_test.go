package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestViewCountMilestone(t *testing.T) {
	tests := []struct {
		name  string
		views library.ViewCount
		want  library.ViewCount
	}{
		{name: "最初の節目に届かない", views: 99_999, want: 0},
		{name: "節目ちょうど", views: 100_000, want: 100_000},
		{name: "節目と節目の間", views: 999_999, want: 100_000},
		{name: "いくつかの節目を超えた", views: 12_345_678, want: 10_000_000},
		{name: "最も高い節目を超えた", views: 5_000_000_000, want: 100_000_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.views.Milestone()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrivacyStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        library.PrivacyStatus
		wantWatchable bool
		wantHidden    bool
	}{
		{name: "公開", status: library.PrivacyPublic, wantWatchable: true, wantHidden: false},
		{name: "限定公開", status: library.PrivacyUnlisted, wantWatchable: true, wantHidden: false},
		{name: "非公開", status: library.PrivacyPrivate, wantWatchable: false, wantHidden: true},
		{name: "削除", status: library.PrivacyDeleted, wantWatchable: false, wantHidden: true},
		{name: "わからない", status: "", wantWatchable: false, wantHidden: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantWatchable, tt.status.Watchable())
			assert.Equal(t, tt.wantHidden, tt.status.Hidden())
		})
	}
}

func TestParseVideoID(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		want   library.VideoID
		wantOK bool
	}{
		{name: "動画 ID", s: "dQw4w9WgXcQ", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "前後に空白がある", s: " dQw4w9WgXcQ\n", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "watch の URL", s: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "他のクエリがある URL", s: "https://www.youtube.com/watch?list=PL&v=dQw4w9WgXcQ&t=42s", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "短縮 URL", s: "https://youtu.be/dQw4w9WgXcQ?si=abc", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "ショートの URL", s: "https://www.youtube.com/shorts/dQw4w9WgXcQ", want: "dQw4w9WgXcQ", wantOK: true},
		{name: "空文字列", s: "", want: "", wantOK: false},
		{name: "短すぎる ID", s: "dQw4w9WgXc", want: "", wantOK: false},
		{name: "使えない文字がある ID", s: "dQw4w9WgXc!", want: "", wantOK: false},
		{name: "動画を指さない URL", s: "https://www.youtube.com/", want: "", wantOK: false},
		{name: "プレイリストの URL", s: "https://www.youtube.com/playlist?list=PLxxxxxxxxxxxxxxxx", want: "", wantOK: false},
		{name: "URL として読めない", s: "%zz", want: "", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := library.ParseVideoID(tt.s)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
