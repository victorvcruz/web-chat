package channel

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"web-chat/internal/platform"
)

type Repository interface {
	Create(ctx context.Context, channel *Channel) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, channel *Channel) (err error)
	FindById(ctx context.Context, id string) (*Channel, error)
	List(ctx context.Context) ([]Channel, error)
	RemoverUser(ctx context.Context, userId string, channelId string) (err error)
}

type repository struct {
	db *gorm.DB
}

func (r *repository) Create(ctx context.Context, channel *Channel) error {
	err := r.db.WithContext(ctx).Preload(clause.Associations).Create(channel).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) Update(ctx context.Context, channel *Channel) (err error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		err = platform.CommitOrRollback(tx, err)
	}()

	err = tx.WithContext(ctx).Preload(clause.Associations).Save(channel).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) RemoverUser(ctx context.Context, userId string, channelId string) (err error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		err = platform.CommitOrRollback(tx, err)
	}()

	err = tx.WithContext(ctx).Where("channel_id = ? AND user_id = ?", channelId, userId).
		Delete(&ChannelUser{}).Error
	if err != nil {
		return
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) (err error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		err = platform.CommitOrRollback(tx, err)
	}()

	err = tx.WithContext(ctx).Where("channel_id = ?", id).Preload(clause.Associations).Delete(&ChannelUser{}).Error
	if err != nil {
		return
	}

	err = tx.WithContext(ctx).Where("id = ?", id).Preload(clause.Associations).Delete(&Channel{}).Error
	if err != nil {
		return
	}

	return
}

func (r *repository) FindById(ctx context.Context, id string) (*Channel, error) {
	var channel Channel
	err := r.db.WithContext(ctx).Preload(clause.Associations).First(&channel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

func (r *repository) List(ctx context.Context) ([]Channel, error) {
	var channels []Channel
	err := r.db.WithContext(ctx).Preload(clause.Associations).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
