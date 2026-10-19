package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/guild"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"

	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *subscriptionRepository {
	return &subscriptionRepository{db}
}

func (r *subscriptionRepository) Exists(guildID guild.DiscordID, playlistID library.PlaylistID) (bool, error) {
	var recs []*subscriptionRecord

	err := r.ofGuild(guildID).Where(subscriptionRecord{YoutubeID: string(playlistID)}).Find(&recs).Error
	if err != nil {
		return false, err
	}

	return len(recs) > 0, nil
}

func (r *subscriptionRepository) FindAll() ([]*subscription.Subscription, error) {
	var recs []*subscriptionRecord
	if err := r.db.Find(&recs).Error; err != nil {
		return nil, err
	}

	if len(recs) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return mapSlice(recs, (*subscriptionRecord).toDomain), nil
}

func (r *subscriptionRepository) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	var recs []*subscriptionRecord
	if err := r.ofGuild(guildID).Find(&recs).Error; err != nil {
		return nil, err
	}

	if len(recs) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return mapSlice(recs, (*subscriptionRecord).toDomain), nil
}

// Add は保存した通知設定に DB が振った ID を s に書き戻す。
func (r *subscriptionRepository) Add(s *subscription.Subscription) error {
	if s.ID != 0 {
		return domain.ErrDBRecordAlreadyCreated
	}

	rec := toSubscriptionRecord(s)

	result := r.db.Save(rec)
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
		return domain.ErrDBRecordNotFound
	}

	rec := toSubscriptionRecord(s)

	result := r.db.Model(rec).Select("*").Omit("created_at", "deleted_at").Updates(rec)
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return domain.ErrDBRecordNotFound
	}

	return nil
}

func (r *subscriptionRepository) DeleteAll(subscriptions []*subscription.Subscription) error {
	var ids []uint

	for _, s := range subscriptions {
		if s.ID == 0 {
			return domain.ErrDBRecordNotFound
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
