package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

// guilds は通知設定が参照するギルドを返す。ギルド 1 は「guild」、ギルド 2 は「other」。
func guilds() []any {
	return []any{guild.New("guild"), guild.New("other")}
}

func TestSubscriptionRepositoryExists(t *testing.T) {
	tests := []struct {
		name       string
		stored     []any
		guildID    guild.DiscordID
		playlistID library.PlaylistID
		want       bool
	}{
		{
			name:       "ギルドが登録している",
			stored:     []any{newSubscription(1, "PL", "PL")},
			guildID:    "guild",
			playlistID: "PL",
			want:       true,
		},
		{
			name:       "別のギルドだけが登録している",
			stored:     []any{newSubscription(2, "PL", "PL")},
			guildID:    "guild",
			playlistID: "PL",
			want:       false,
		},
		{
			name:       "ギルドを保存していない",
			stored:     []any{newSubscription(1, "PL", "PL")},
			guildID:    "unknown",
			playlistID: "PL",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			got, err := persistence.NewSubscriptionRepository(db).Exists(tt.guildID, tt.playlistID)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubscriptionRepositoryFindAll(t *testing.T) {
	tests := []struct {
		name    string
		stored  []any
		want    []*subscription.Subscription
		wantErr error
	}{
		{
			name:   "全てのギルドの通知設定を返す",
			stored: []any{newSubscription(1, "PL1", "PL1"), newSubscription(2, "PL2", "PL2")},
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				storedSubscription(2, 2, "PL2", "PL2"),
			},
		},
		{
			name:   "論理削除した通知設定は返さない",
			stored: []any{deletedSubscription(1, "PL1", "PL1"), newSubscription(2, "PL2", "PL2")},
			want: []*subscription.Subscription{
				storedSubscription(2, 2, "PL2", "PL2"),
			},
		},
		{
			// 既知の挙動: 空のスライスではなくエラーを返す
			name:    "通知設定がなければ ErrDBRecordNotFound を返す",
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			got, err := persistence.NewSubscriptionRepository(db).FindAll()

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubscriptionRepositoryFindByGuild(t *testing.T) {
	tests := []struct {
		name    string
		stored  []any
		guildID guild.DiscordID
		want    []*subscription.Subscription
		wantErr error
	}{
		{
			name:    "ギルドの通知設定だけを返す",
			stored:  []any{newSubscription(1, "PL1", "PL1"), newSubscription(2, "PL2", "PL2"), newSubscription(1, "PL3", "PL3")},
			guildID: "guild",
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				storedSubscription(3, 1, "PL3", "PL3"),
			},
		},
		{
			name:    "ギルドに通知設定がなければ ErrDBRecordNotFound を返す",
			stored:  []any{newSubscription(2, "PL2", "PL2")},
			guildID: "guild",
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
		{
			name:    "ギルドを保存していなければ ErrDBRecordNotFound を返す",
			stored:  []any{newSubscription(1, "PL1", "PL1")},
			guildID: "unknown",
			want:    nil,
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			got, err := persistence.NewSubscriptionRepository(db).FindByGuild(tt.guildID)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubscriptionRepositoryAdd(t *testing.T) {
	tests := []struct {
		name         string
		subscription *subscription.Subscription
		want         []*subscription.Subscription
		wantErr      error
	}{
		{
			name:         "通知設定を保存する",
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1, NotifiedUntil: hoursAgo(6)},
			want: []*subscription.Subscription{
				{Model: model(1), YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1, NotifiedUntil: hoursAgo(6)},
			},
		},
		{
			name:         "ID のある通知設定は保存しない",
			subscription: &subscription.Subscription{ID: 5, YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1},
			want:         []*subscription.Subscription{},
			wantErr:      domain.ErrDBRecordAlreadyCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)

			err := persistence.NewSubscriptionRepository(db).Add(tt.subscription)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[subscription.Subscription](t, db))
		})
	}
}

func TestSubscriptionRepositoryUpdate(t *testing.T) {
	tests := []struct {
		name         string
		subscription *subscription.Subscription
		want         []*subscription.Subscription
		wantErr      error
	}{
		{
			// GORM の Save は UpdatedAt を保存した時刻で上書きするが、新着動画の
			// 通知の基準時刻（NotifiedUntil）には触れない
			name: "全てのフィールドを保存し、UpdatedAt だけは保存した時刻にする",
			subscription: &subscription.Subscription{
				ID: 1, CreatedAt: now, UpdatedAt: hoursAgo(1),
				YoutubeID:     "PL",
				SendChannelID: "ch",
				Title:         "新しいタイトル",
				PickInterval:  subscription.PickDaily,
				GuildID:       1,
				NotifiedUntil: hoursAgo(1),
			},
			want: []*subscription.Subscription{
				{Model: model(1), YoutubeID: "PL", SendChannelID: "ch", Title: "新しいタイトル", PickInterval: subscription.PickDaily, GuildID: 1, NotifiedUntil: hoursAgo(1)},
			},
		},
		{
			name:         "ID のない通知設定は保存しない",
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "新しいタイトル", GuildID: 1},
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL", "PL"),
			},
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, newSubscription(1, "PL", "PL"))

			err := persistence.NewSubscriptionRepository(db).Update(tt.subscription)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[subscription.Subscription](t, db))
		})
	}
}

func TestSubscriptionRepositoryDeleteAll(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		want          []*subscription.Subscription
		wantErr       error
	}{
		{
			name: "指定した通知設定を論理削除する",
			subscriptions: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
			},
			want: []*subscription.Subscription{
				{Model: deleted(1), YoutubeID: "PL1", SendChannelID: "ch", Title: "PL1", GuildID: 1, NotifiedUntil: hoursAgo(6)},
				storedSubscription(2, 1, "PL2", "PL2"),
			},
		},
		{
			name: "ID のない通知設定があれば、どれも削除しない",
			subscriptions: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				{YoutubeID: "PL2"},
			},
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				storedSubscription(2, 1, "PL2", "PL2"),
			},
			wantErr: domain.ErrDBRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, newSubscription(1, "PL1", "PL1"), newSubscription(1, "PL2", "PL2"))

			err := persistence.NewSubscriptionRepository(db).DeleteAll(tt.subscriptions)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[subscription.Subscription](t, db))
		})
	}
}
