package persistence

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *channelRepository {
	return &channelRepository{db}
}

func (r *channelRepository) Find(ids []library.ChannelID) (map[library.ChannelID]*library.Channel, error) {
	channels := map[library.ChannelID]*library.Channel{}
	if len(ids) == 0 {
		return channels, nil
	}

	var recs []*channelRecord
	if err := r.db.Where("youtube_id IN ?", ids).Find(&recs).Error; err != nil {
		return nil, err
	}

	for _, rec := range recs {
		channels[library.ChannelID(rec.YoutubeID)] = rec.toDomain()
	}

	return channels, nil
}

func (r *channelRepository) Save(channels []*library.Channel) error {
	if len(channels) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "updated_at"}),
	}).CreateInBatches(mapSlice(channels, toChannelRecord), 500).Error
}
