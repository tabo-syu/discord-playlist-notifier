package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

func TestVideoAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		video *library.Video
		want  bool
	}{
		{
			name:  "公開",
			video: video("a", library.PrivacyPublic, 0),
			want:  true,
		},
		{
			name:  "限定公開",
			video: video("a", library.PrivacyUnlisted, 0),
			want:  true,
		},
		{
			name:  "非公開",
			video: video("a", library.PrivacyPrivate, 0),
			want:  false,
		},
		{
			name:  "削除",
			video: video("a", library.PrivacyDeleted, 0),
			want:  false,
		},
		{
			name: "詳細をまだ取得していない",
			video: &library.Video{
				YoutubeID:     "a",
				PrivacyStatus: library.PrivacyPublic,
			},
			want: false,
		},
		{
			name: "タイトルが空でも、詳細を取得済みなら見せる",
			video: &library.Video{
				YoutubeID:      "a",
				DetailsFetched: true,
				PrivacyStatus:  library.PrivacyPublic,
			},
			want: true,
		},
		{
			name: "タイトルがあっても、詳細を取得済みでなければ見せない",
			video: &library.Video{
				YoutubeID:     "a",
				Title:         "a",
				PrivacyStatus: library.PrivacyPublic,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.video.Available()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVideoChangePrivacy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		video       *library.Video
		status      library.PrivacyStatus
		want        *library.Video
		wantChanged bool
		wantHidden  *library.VideoHidden
	}{
		{
			name:        "変わらない",
			video:       video("a", library.PrivacyPublic, 0),
			status:      library.PrivacyPublic,
			want:        video("a", library.PrivacyPublic, 0),
			wantChanged: false,
			wantHidden:  nil,
		},
		{
			name:        "公開から非公開になった",
			video:       video("a", library.PrivacyPublic, 0),
			status:      library.PrivacyPrivate,
			want:        video("a", library.PrivacyPrivate, 0),
			wantChanged: true,
			wantHidden:  &library.VideoHidden{Video: video("a", library.PrivacyPrivate, 0)},
		},
		{
			name:        "限定公開から削除になった",
			video:       video("a", library.PrivacyUnlisted, 0),
			status:      library.PrivacyDeleted,
			want:        video("a", library.PrivacyDeleted, 0),
			wantChanged: true,
			wantHidden:  &library.VideoHidden{Video: video("a", library.PrivacyDeleted, 0)},
		},
		{
			name:        "非公開から公開に戻った",
			video:       video("a", library.PrivacyPrivate, 0),
			status:      library.PrivacyPublic,
			want:        video("a", library.PrivacyPublic, 0),
			wantChanged: true,
			wantHidden:  nil,
		},
		{
			name:        "非公開から削除になった",
			video:       video("a", library.PrivacyPrivate, 0),
			status:      library.PrivacyDeleted,
			want:        video("a", library.PrivacyDeleted, 0),
			wantChanged: true,
			wantHidden:  nil,
		},
		{
			name:   "初めて見る動画が非公開",
			video:  &library.Video{YoutubeID: "a"},
			status: library.PrivacyPrivate,
			want: &library.Video{
				YoutubeID:     "a",
				PrivacyStatus: library.PrivacyPrivate,
			},
			wantChanged: true,
			wantHidden:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			changed, hidden := tt.video.ChangePrivacy(tt.status)

			assert.Equal(t, tt.want, tt.video)
			assert.Equal(t, tt.wantChanged, changed)
			assert.Equal(t, tt.wantHidden, hidden)
		})
	}
}

func TestVideoApplyDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		video   *library.Video
		details *library.Video
		want    *library.Video
	}{
		{
			name: "公開状態と ID 以外をコピーし、詳細を取得済みにする",
			video: &library.Video{
				YoutubeID:     "a",
				PrivacyStatus: library.PrivacyUnlisted,
			},
			details: &library.Video{
				YoutubeID:        "b",
				Title:            "曲",
				Views:            100,
				Likes:            10,
				Comments:         1,
				PrivacyStatus:    library.PrivacyPublic,
				ChannelYoutubeID: "UC",
				Tags:             []string{"tag"},
				CategoryID:       "10",
			},
			want: &library.Video{
				YoutubeID:        "a",
				DetailsFetched:   true,
				Title:            "曲",
				Views:            100,
				Likes:            10,
				Comments:         1,
				PrivacyStatus:    library.PrivacyUnlisted,
				ChannelYoutubeID: "UC",
				Tags:             []string{"tag"},
				CategoryID:       "10",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.video.ApplyDetails(tt.details)

			assert.Equal(t, tt.want, tt.video)
		})
	}
}

func TestVideoInitMilestone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		video *library.Video
		want  *library.Video
	}{
		{
			name:  "今の再生数の節目を記録する",
			video: &library.Video{Views: 2_500_000},
			want: &library.Video{
				Views:         2_500_000,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
		},
		{
			name: "記録済みの節目は変えない",
			video: &library.Video{
				Views:         2_500_000,
				ViewMilestone: new(library.ViewCount(100_000)),
			},
			want: &library.Video{
				Views:         2_500_000,
				ViewMilestone: new(library.ViewCount(100_000)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.video.InitMilestone()

			assert.Equal(t, tt.want, tt.video)
		})
	}
}

func TestVideoUpdateMilestone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		video       *library.Video
		want        *library.Video
		wantReached *library.MilestoneReached
	}{
		{
			name:  "初めて再生数がわかったときは通知しない",
			video: &library.Video{Views: 2_500_000},
			want: &library.Video{
				Views:         2_500_000,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			wantReached: nil,
		},
		{
			name: "新しい節目に達した",
			video: &library.Video{
				Views:         1_000_000,
				ViewMilestone: new(library.ViewCount(100_000)),
			},
			want: &library.Video{
				Views:         1_000_000,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			wantReached: &library.MilestoneReached{
				Video: &library.Video{
					Views:         1_000_000,
					ViewMilestone: new(library.ViewCount(1_000_000)),
				},
				Milestone: 1_000_000,
			},
		},
		{
			name: "同じ節目のまま",
			video: &library.Video{
				Views:         9_999_999,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			want: &library.Video{
				Views:         9_999_999,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			wantReached: nil,
		},
		{
			name: "再生数が減っても節目は下げない",
			video: &library.Video{
				Views:         999_999,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			want: &library.Video{
				Views:         999_999,
				ViewMilestone: new(library.ViewCount(1_000_000)),
			},
			wantReached: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reached := tt.video.UpdateMilestone()

			assert.Equal(t, tt.want, tt.video)
			assert.Equal(t, tt.wantReached, reached)
		})
	}
}
