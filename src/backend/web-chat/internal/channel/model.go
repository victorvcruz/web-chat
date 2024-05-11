package channel

import (
	"gorm.io/gorm"
	"time"
	"web-chat/internal/user"
)

type Channel struct {
	ID        string         `json:"id" gorm:"column:id;primarykey;type:uuid;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"column:name;type:varchar(255)"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;type:timestamp"`
	// many to many
	Users []user.User `json:"users" gorm:"many2many:channel_users"`
}

type ChannelUser struct {
	ChannelID string `json:"channel_id" gorm:"column:channel_id;type:uuid"`
	UserID    string `json:"user_id" gorm:"column:user_id;type:uuid"`
}
