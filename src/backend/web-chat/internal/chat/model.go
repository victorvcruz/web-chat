package chat

import (
	"gorm.io/gorm"
	"time"
)

type Chat struct {
	ID         string         `json:"id" gorm:"column:id;primarykey;type:uuid;default:gen_random_uuid()"`
	ChannelID  string         `json:"channel_id" gorm:"column:channel_id;type:uuid;"`
	SenderId   string         `json:"sender_id" gorm:"column:sender_id;type:uuid;"`
	Message    string         `json:"message" gorm:"column:message;type:text;"`
	CreatedAt  time.Time      `json:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"column:deleted_at;type:timestamp"`
	SenderName string         `json:"sender_name" `
}
