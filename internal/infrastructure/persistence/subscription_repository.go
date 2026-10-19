package persistence

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"gorm.io/gorm"
)

// subscriptionUniqueIndex は、同じギルドに同じプレイリストの通知設定を 1 つだけにする
// 一意のインデックス（subscriptionRecord のタグ）の名前。
const subscriptionUniqueIndex = "idx_playlists_guild_youtube"

// uniqueViolation は PostgreSQL の一意制約の違反のエラーコード。
const uniqueViolation = "23505"

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *subscriptionRepository {
	return &subscriptionRepository{db}
}

// FindByGuildAndPlaylist は、ギルドとプレイリストの組で通知設定を引く。一意の
// インデックスがあるので、論理削除していない行は多くても 1 つ。
func (r *subscriptionRepository) FindByGuildAndPlaylist(guildID guild.DiscordID, playlistID library.PlaylistID) (*subscription.Subscription, error) {
	var rec subscriptionRecord

	result := r.ofGuild(guildID).Where(subscriptionRecord{YoutubeID: string(playlistID)}).Take(&rec)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, subscription.ErrNotSubscribed
	}

	if err := result.Error; err != nil {
		return nil, err
	}

	return rec.toDomain(), nil
}

func (r *subscriptionRepository) FindAll() ([]*subscription.Subscription, error) {
	var recs []*subscriptionRecord
	if err := r.db.Find(&recs).Error; err != nil {
		return nil, err
	}

	if len(recs) == 0 {
		return nil, subscription.ErrNotSubscribed
	}

	return mapSlice(recs, (*subscriptionRecord).toDomain), nil
}

func (r *subscriptionRepository) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	var recs []*subscriptionRecord
	if err := r.ofGuild(guildID).Find(&recs).Error; err != nil {
		return nil, err
	}

	if len(recs) == 0 {
		return nil, subscription.ErrNotSubscribed
	}

	return mapSlice(recs, (*subscriptionRecord).toDomain), nil
}

// Add は保存した通知設定に DB が振った ID を s に書き戻す。同じギルドに同じプレイリストが
// あるかは、事前に確かめると同時に登録したときに重複するので、一意のインデックスに任せ、
// その違反を ErrAlreadySubscribed にする。
func (r *subscriptionRepository) Add(s *subscription.Subscription) error {
	if s.ID != 0 {
		return subscription.ErrAlreadySubscribed
	}

	rec := toSubscriptionRecord(s)

	result := r.db.Create(rec)
	if isUniqueViolation(result.Error, subscriptionUniqueIndex) {
		return subscription.ErrAlreadySubscribed
	}

	if err := result.Error; err != nil {
		return err
	}

	s.ID = rec.ID

	return nil
}

// Save は対象の行がない（論理削除されている）と INSERT し直し、解除された通知設定を
// 元に戻してしまうので、カラムを Select して Updates する。ドメインのモデルは
// created_at を持たないので、レコードのゼロ値で上書きしないように除く。
func (r *subscriptionRepository) Update(s *subscription.Subscription) error {
	if s.ID == 0 {
		return subscription.ErrNotSubscribed
	}

	rec := toSubscriptionRecord(s)

	result := r.db.Model(rec).Select("*").Omit("created_at", "deleted_at").Updates(rec)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return subscription.ErrNotSubscribed
	}

	return nil
}

func (r *subscriptionRepository) DeleteAll(subscriptions []*subscription.Subscription) error {
	var ids []uint

	for _, s := range subscriptions {
		if s.ID == 0 {
			return subscription.ErrNotSubscribed
		}

		ids = append(ids, s.ID)
	}

	err := r.db.Delete(&subscriptionRecord{}, ids).Error
	if err != nil {
		return err
	}

	return nil
}

// ofGuild はクエリをギルドの通知設定に絞り込む。ギルドが見つからないときは
// ID が 0 になり、何にも一致しない。
func (r *subscriptionRepository) ofGuild(guildID guild.DiscordID) *gorm.DB {
	var g guildRecord
	r.db.Where(guildRecord{DiscordID: string(guildID)}).Find(&g)

	return r.db.Where("guild_id = ?", g.ID)
}

// isUniqueViolation は err が一意のインデックス index の違反かを返す。
func isUniqueViolation(err error, index string) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == index
}
