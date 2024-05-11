package chat

import (
	"context"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, chat *Chat) error
	Update(ctx context.Context, chat *Chat) error
	FindById(ctx context.Context, id string) (*Chat, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, id string) ([]Chat, error)
}

type repository struct {
	db *gorm.DB
}

func (r *repository) Create(ctx context.Context, chat *Chat) error {
	err := r.db.WithContext(ctx).Create(chat).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) Update(ctx context.Context, chat *Chat) error {
	err := r.db.WithContext(ctx).Model(chat).UpdateColumns(Chat{
		Message: chat.Message,
	}).Error
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).First(chat, "id = ?", chat.ID).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) FindById(ctx context.Context, id string) (*Chat, error) {
	var chat Chat
	err := r.db.WithContext(ctx).Raw(`SELECT chats.*, users.username as sender_name
FROM chats
INNER JOIN users ON chats.sender_id = users.id
WHERE chats.id = ?`, id).Scan(&chat).Error
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Chat{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) List(ctx context.Context, id string) ([]Chat, error) {
	var chats []Chat
	err := r.db.WithContext(ctx).Raw(`SELECT chats.*, users.username as sender_name
	FROM chats
	INNER JOIN users ON chats.sender_id = users.id
	WHERE chats.channel_id = ?
	ORDER BY chats.created_at ASC`, id).Scan(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
