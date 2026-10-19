package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
)

// guilds は通知設定が参照するギルドを返す。ギルド 1 は「guild」、ギルド 2 は「other」。
func guilds() []any {
	return []any{guildRow("guild"), guildRow("other")}
}

func TestSubscriptionRepositoryFindByGuildAndPlaylist(t *testing.T) {
	tests := []struct {
		name       string
		stored     []any
		guildID    guild.DiscordID
		playlistID library.PlaylistID
		want       *subscription.Subscription
		wantErr    error
	}{
		{
			name:       "ギルドとプレイリストの組で通知設定を返す",
			stored:     []any{subscriptionRow(1, "PL1", "PL1"), subscriptionRow(1, "PL2", "PL2")},
			guildID:    "guild",
			playlistID: "PL2",
			want:       storedSubscription(2, 1, "PL2", "PL2"),
			wantErr:    nil,
		},
		{
			name:       "別のギルドだけが登録していれば ErrNotSubscribed を返す",
			stored:     []any{subscriptionRow(2, "PL", "PL")},
			guildID:    "guild",
			playlistID: "PL",
			want:       nil,
			wantErr:    subscription.ErrNotSubscribed,
		},
		{
			name:       "論理削除した通知設定は返さない",
			stored:     []any{deletedSubscriptionRow(1, "PL", "PL")},
			guildID:    "guild",
			playlistID: "PL",
			want:       nil,
			wantErr:    subscription.ErrNotSubscribed,
		},
		{
			name:       "ギルドを保存していなければ ErrNotSubscribed を返す",
			stored:     []any{subscriptionRow(1, "PL", "PL")},
			guildID:    "unknown",
			playlistID: "PL",
			want:       nil,
			wantErr:    subscription.ErrNotSubscribed,
		},
		{
			name:       "論理削除したギルドの通知設定は返さない",
			stored:     []any{&persistence.GuildRecord{Model: deleted(3), DiscordID: "gone"}, subscriptionRow(3, "PL", "PL")},
			guildID:    "gone",
			playlistID: "PL",
			want:       nil,
			wantErr:    subscription.ErrNotSubscribed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			got, err := persistence.NewSubscriptionRepository(db).FindByGuildAndPlaylist(tt.guildID, tt.playlistID)

			assert.ErrorIs(t, err, tt.wantErr)
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
			stored: []any{subscriptionRow(1, "PL1", "PL1"), subscriptionRow(2, "PL2", "PL2")},
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				storedSubscription(2, 2, "PL2", "PL2"),
			},
		},
		{
			name:   "論理削除した通知設定は返さない",
			stored: []any{deletedSubscriptionRow(1, "PL1", "PL1"), subscriptionRow(2, "PL2", "PL2")},
			want: []*subscription.Subscription{
				storedSubscription(2, 2, "PL2", "PL2"),
			},
		},
		{
			// 既知の挙動: 空のスライスではなくエラーを返す
			name:    "通知設定がなければ ErrDBRecordNotFound を返す",
			want:    nil,
			wantErr: subscription.ErrNotSubscribed,
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
			stored:  []any{subscriptionRow(1, "PL1", "PL1"), subscriptionRow(2, "PL2", "PL2"), subscriptionRow(1, "PL3", "PL3")},
			guildID: "guild",
			want: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				storedSubscription(3, 1, "PL3", "PL3"),
			},
		},
		{
			name:    "ギルドに通知設定がなければ ErrDBRecordNotFound を返す",
			stored:  []any{subscriptionRow(2, "PL2", "PL2")},
			guildID: "guild",
			want:    nil,
			wantErr: subscription.ErrNotSubscribed,
		},
		{
			name:    "ギルドを保存していなければ ErrDBRecordNotFound を返す",
			stored:  []any{subscriptionRow(1, "PL1", "PL1")},
			guildID: "unknown",
			want:    nil,
			wantErr: subscription.ErrNotSubscribed,
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
		name             string
		stored           []any
		subscription     *subscription.Subscription
		want             []*persistence.SubscriptionRecord
		wantSubscription *subscription.Subscription
		wantErr          error
	}{
		{
			name:         "通知設定を保存し、振った ID を書き戻す",
			stored:       []any{},
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1, NotifiedUntil: hoursAgo(6)},
			want: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 1, "PL", "PL"),
			},
			wantSubscription: storedSubscription(1, 1, "PL", "PL"),
		},
		{
			name:             "ID のある通知設定は保存しない",
			stored:           []any{},
			subscription:     &subscription.Subscription{ID: 5, YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1},
			want:             []*persistence.SubscriptionRecord{},
			wantSubscription: &subscription.Subscription{ID: 5, YoutubeID: "PL", SendChannelID: "ch", Title: "PL", GuildID: 1},
			wantErr:          subscription.ErrAlreadySubscribed,
		},
		{
			// 一意のインデックスが断る。失敗した INSERT も ID の連番を 1 つ進める
			name:         "同じギルドに同じプレイリストがあれば保存しない",
			stored:       []any{subscriptionRow(1, "PL", "PL")},
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "other", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1, NotifiedUntil: hoursAgo(6)},
			want: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 1, "PL", "PL"),
			},
			wantSubscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "other", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1, NotifiedUntil: hoursAgo(6)},
			wantErr:          subscription.ErrAlreadySubscribed,
		},
		{
			name:         "別のギルドが同じプレイリストを登録していても保存する",
			stored:       []any{subscriptionRow(2, "PL", "PL")},
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1, NotifiedUntil: hoursAgo(6)},
			want: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 2, "PL", "PL"),
				storedSubscriptionRow(2, 1, "PL", "PL"),
			},
			wantSubscription: storedSubscription(2, 1, "PL", "PL"),
		},
		{
			name:         "解除した（論理削除した）プレイリストは登録し直せる",
			stored:       []any{deletedSubscriptionRow(1, "PL", "PL")},
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: subscription.PickWeekly, GuildID: 1, NotifiedUntil: hoursAgo(6)},
			want: []*persistence.SubscriptionRecord{
				{Model: deleted(1), YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: "weekly", GuildID: 1, NotifiedUntil: hoursAgo(6)},
				storedSubscriptionRow(2, 1, "PL", "PL"),
			},
			wantSubscription: storedSubscription(2, 1, "PL", "PL"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			err := persistence.NewSubscriptionRepository(db).Add(tt.subscription)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[persistence.SubscriptionRecord](t, db))
			assert.Equal(t, tt.wantSubscription, tt.subscription)
		})
	}
}

func TestSubscriptionRepositoryAddConcurrently(t *testing.T) {
	tests := []struct {
		name string
		// 同時に Add する回数
		n int
		// 保存された行の数。どの呼び出しが勝つかは決まらないので、行の中身ではなく数を比べる
		wantSaved int
		wantErrs  []error
	}{
		{
			name:      "同じギルドに同じプレイリストを同時に登録しても、1 つだけ保存する",
			n:         5,
			wantSaved: 1,
			wantErrs: []error{
				nil,
				subscription.ErrAlreadySubscribed,
				subscription.ErrAlreadySubscribed,
				subscription.ErrAlreadySubscribed,
				subscription.ErrAlreadySubscribed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)

			errs := addConcurrently(persistence.NewSubscriptionRepository(db), tt.n)

			assert.ElementsMatch(t, tt.wantErrs, errs)
			assert.Len(t, rows[persistence.SubscriptionRecord](t, db), tt.wantSaved)
		})
	}
}

func TestSubscriptionRepositoryUpdate(t *testing.T) {
	tests := []struct {
		name         string
		stored       []any
		subscription *subscription.Subscription
		want         []*persistence.SubscriptionRecord
		wantErr      error
	}{
		{
			// 作った時刻は残し、updated_at は GORM が保存した時刻にする
			name: "全てのフィールドを保存し、created_at は変えない",
			stored: []any{
				&persistence.SubscriptionRecord{CreatedAt: hoursAgo(12), UpdatedAt: hoursAgo(12), YoutubeID: "PL", SendChannelID: "old", Title: "PL", GuildID: 1, NotifiedUntil: hoursAgo(6)},
			},
			subscription: &subscription.Subscription{
				ID:            1,
				YoutubeID:     "PL",
				SendChannelID: "ch",
				Title:         "新しいタイトル",
				PickInterval:  subscription.PickDaily,
				GuildID:       1,
				NotifiedUntil: hoursAgo(1),
			},
			want: []*persistence.SubscriptionRecord{
				{ID: 1, CreatedAt: hoursAgo(12), UpdatedAt: now, YoutubeID: "PL", SendChannelID: "ch", Title: "新しいタイトル", PickInterval: "daily", GuildID: 1, NotifiedUntil: hoursAgo(1)},
			},
		},
		{
			name:         "ID のない通知設定は保存しない",
			stored:       []any{subscriptionRow(1, "PL", "PL")},
			subscription: &subscription.Subscription{YoutubeID: "PL", SendChannelID: "ch", Title: "新しいタイトル", GuildID: 1},
			want: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 1, "PL", "PL"),
			},
			wantErr: subscription.ErrNotSubscribed,
		},
		{
			// 読んだ後に解除された通知設定を、保存し直して元に戻さない
			name:         "論理削除した通知設定は保存しない",
			stored:       []any{deletedSubscriptionRow(1, "PL", "PL")},
			subscription: storedSubscription(1, 1, "PL", "新しいタイトル"),
			want: []*persistence.SubscriptionRecord{
				{Model: deleted(1), YoutubeID: "PL", SendChannelID: "ch", Title: "PL", PickInterval: "weekly", GuildID: 1, NotifiedUntil: hoursAgo(6)},
			},
			wantErr: subscription.ErrNotSubscribed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, tt.stored...)

			err := persistence.NewSubscriptionRepository(db).Update(tt.subscription)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[persistence.SubscriptionRecord](t, db))
		})
	}
}

func TestSubscriptionRepositoryDeleteAll(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []*subscription.Subscription
		want          []*persistence.SubscriptionRecord
		wantErr       error
	}{
		{
			name: "指定した通知設定を論理削除する",
			subscriptions: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
			},
			want: []*persistence.SubscriptionRecord{
				{Model: deleted(1), YoutubeID: "PL1", SendChannelID: "ch", Title: "PL1", PickInterval: "weekly", GuildID: 1, NotifiedUntil: hoursAgo(6)},
				storedSubscriptionRow(2, 1, "PL2", "PL2"),
			},
		},
		{
			name: "ID のない通知設定があれば、どれも削除しない",
			subscriptions: []*subscription.Subscription{
				storedSubscription(1, 1, "PL1", "PL1"),
				{YoutubeID: "PL2"},
			},
			want: []*persistence.SubscriptionRecord{
				storedSubscriptionRow(1, 1, "PL1", "PL1"),
				storedSubscriptionRow(2, 1, "PL2", "PL2"),
			},
			wantErr: subscription.ErrNotSubscribed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newDB(t)
			seed(t, db, guilds()...)
			seed(t, db, subscriptionRow(1, "PL1", "PL1"), subscriptionRow(1, "PL2", "PL2"))

			err := persistence.NewSubscriptionRepository(db).DeleteAll(tt.subscriptions)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, rows[persistence.SubscriptionRecord](t, db))
		})
	}
}
