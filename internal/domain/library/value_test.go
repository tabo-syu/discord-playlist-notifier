package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestViewCountMilestone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		views library.ViewCount
		want  library.ViewCount
	}{
		{
			name:  "最初の節目に届かない",
			views: 99_999,
			want:  0,
		},
		{
			name:  "節目ちょうど",
			views: 100_000,
			want:  100_000,
		},
		{
			name:  "節目と節目の間",
			views: 999_999,
			want:  100_000,
		},
		{
			name:  "いくつかの節目を超えた",
			views: 12_345_678,
			want:  10_000_000,
		},
		{
			name:  "最も高い節目を超えた",
			views: 5_000_000_000,
			want:  100_000_000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.views.Milestone()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrivacyStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		status        library.PrivacyStatus
		wantWatchable bool
		wantHidden    bool
	}{
		{
			name:          "公開",
			status:        library.PrivacyPublic,
			wantWatchable: true,
			wantHidden:    false,
		},
		{
			name:          "限定公開",
			status:        library.PrivacyUnlisted,
			wantWatchable: true,
			wantHidden:    false,
		},
		{
			name:          "非公開",
			status:        library.PrivacyPrivate,
			wantWatchable: false,
			wantHidden:    true,
		},
		{
			name:          "削除",
			status:        library.PrivacyDeleted,
			wantWatchable: false,
			wantHidden:    true,
		},
		{
			name:          "わからない",
			status:        "",
			wantWatchable: false,
			wantHidden:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantWatchable, tt.status.Watchable())
			assert.Equal(t, tt.wantHidden, tt.status.Hidden())
		})
	}
}

func TestParseVideoID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       string
		want    library.VideoID
		wantErr error
	}{
		{
			name:    "動画 ID",
			s:       "dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "前後に空白がある",
			s:       " dQw4w9WgXcQ\n",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "watch の URL",
			s:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "他のクエリがある URL",
			s:       "https://www.youtube.com/watch?list=PL&v=dQw4w9WgXcQ&t=42s",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "短縮 URL",
			s:       "https://youtu.be/dQw4w9WgXcQ?si=abc",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "ショートの URL",
			s:       "https://www.youtube.com/shorts/dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "空文字列",
			s:       "",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
		{
			name:    "短すぎる ID",
			s:       "dQw4w9WgXc",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
		{
			name:    "使えない文字がある ID",
			s:       "dQw4w9WgXc!",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
		{
			name:    "動画を指さない URL",
			s:       "https://www.youtube.com/",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
		{
			name:    "プレイリストの URL",
			s:       "https://www.youtube.com/playlist?list=PLxxxxxxxxxxxxxxxx",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
		{
			name:    "URL として読めない",
			s:       "%zz",
			want:    "",
			wantErr: library.ErrInvalidVideoID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := library.ParseVideoID(tt.s)

			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParsePlaylistID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       string
		want    library.PlaylistID
		wantErr error
	}{
		{
			name:    "プレイリスト ID",
			s:       "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			want:    "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			wantErr: nil,
		},
		{
			name:    "接頭辞や長さの違うプレイリスト ID",
			s:       "OLAK5uy_kXo3W2vZ-Lr0pHkkVQ1lgg5tYyDmKj6aI",
			want:    "OLAK5uy_kXo3W2vZ-Lr0pHkkVQ1lgg5tYyDmKj6aI",
			wantErr: nil,
		},
		{
			name:    "前後に空白がある",
			s:       " PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf\n",
			want:    "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			wantErr: nil,
		},
		{
			name:    "空文字列",
			s:       "",
			want:    "",
			wantErr: library.ErrInvalidPlaylistID,
		},
		{
			name:    "空白だけ",
			s:       "  ",
			want:    "",
			wantErr: library.ErrInvalidPlaylistID,
		},
		{
			name:    "プレイリストの URL",
			s:       "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			want:    "",
			wantErr: library.ErrInvalidPlaylistID,
		},
		{
			name:    "全角文字がある",
			s:       "ＰＬrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			want:    "",
			wantErr: library.ErrInvalidPlaylistID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := library.ParsePlaylistID(tt.s)

			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseChannelID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       string
		want    library.ChannelID
		wantErr error
	}{
		{
			name:    "チャンネル ID",
			s:       "UCuAXFkgsw1L7xaCfnd5JJOw",
			want:    "UCuAXFkgsw1L7xaCfnd5JJOw",
			wantErr: nil,
		},
		{
			name:    "前後に空白がある",
			s:       " UCuAXFkgsw1L7xaCfnd5JJOw ",
			want:    "UCuAXFkgsw1L7xaCfnd5JJOw",
			wantErr: nil,
		},
		{
			name:    "空文字列",
			s:       "",
			want:    "",
			wantErr: library.ErrInvalidChannelID,
		},
		{
			name:    "チャンネルの URL",
			s:       "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
			want:    "",
			wantErr: library.ErrInvalidChannelID,
		},
		{
			name:    "ハンドル",
			s:       "@YouTube",
			want:    "",
			wantErr: library.ErrInvalidChannelID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := library.ParseChannelID(tt.s)

			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
