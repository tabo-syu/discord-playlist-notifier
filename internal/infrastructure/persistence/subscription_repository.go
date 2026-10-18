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
	var subscriptions []*subscription.Subscription

	err := r.ofGuild(guildID).Where(subscription.Subscription{YoutubeID: playlistID}).Find(&subscriptions).Error
	if err != nil {
		return false, err
	}

	return len(subscriptions) > 0, nil
}

func (r *subscriptionRepository) FindAll() ([]*subscription.Subscription, error) {
	var subscriptions []*subscription.Subscription
	if err := r.db.Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	if len(subscriptions) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) FindByGuild(guildID guild.DiscordID) ([]*subscription.Subscription, error) {
	var subscriptions []*subscription.Subscription
	if err := r.ofGuild(guildID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	if len(subscriptions) == 0 {
		return nil, domain.ErrDBRecordNotFound
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) Add(s *subscription.Subscription) error {
	if s.ID != 0 {
		return domain.ErrDBRecordAlreadyCreated
	}

	result := r.db.Save(s)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *subscriptionRepository) Update(s *subscription.Subscription) error {
	if s.ID == 0 {
		return domain.ErrDBRecordNotFound
	}

	result := r.db.Save(s)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

// UpdatedAt に触れないように、Save ではなく UpdateColumn を使う。
func (r *subscriptionRepository) UpdatePickInterval(s *subscription.Subscription) error {
	if s.ID == 0 {
		return domain.ErrDBRecordNotFound
	}

	return r.db.Model(s).UpdateColumn("pick_interval", s.PickInterval).Error
}

func (r *subscriptionRepository) DeleteAll(subscriptions []*subscription.Subscription) error {
	var ids []uint

	for _, s := range subscriptions {
		if s.ID == 0 {
			return domain.ErrDBRecordNotFound
		}

		ids = append(ids, s.ID)
	}

	err := r.db.Delete(subscriptions, ids).Error
	if err != nil {
		return err
	}

	return nil
}

// ofGuild はクエリをギルドの通知設定に絞り込む。ギルドが見つからないときは
// ID が 0 になり、何にも一致しない。
func (r *subscriptionRepository) ofGuild(guildID guild.DiscordID) *gorm.DB {
	var g guild.Guild
	r.db.Where(guild.Guild{DiscordID: guildID}).Find(&g)

	return r.db.Where("guild_id = ?", g.ID)
}
