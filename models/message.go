package models

import "time"

// Message 私信模型
type Message struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SenderID   uint      `json:"sender_id" gorm:"index;not null"`
	ReceiverID uint      `json:"receiver_id" gorm:"index;not null"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Sender   *User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Receiver *User `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
}

// TableName 表名
func (Message) TableName() string {
	return "messages"
}

// Conversation 会话概览
type Conversation struct {
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name"`
	UserAvatar   string    `json:"user_avatar"`
	LastMessage  string    `json:"last_message"`
	LastTime     time.Time `json:"last_time"`
	UnreadCount  int64     `json:"unread_count"`
}
