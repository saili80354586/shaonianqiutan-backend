package models

import (
	"time"

	"gorm.io/gorm"
)

// Announcement 俱乐部公告模型
type Announcement struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ClubID    uint           `json:"club_id" gorm:"index;not null"`
	Title     string         `json:"title" gorm:"size:200;not null"`
	Content   string         `json:"content" gorm:"type:text"`
	IsPinned  bool           `json:"is_pinned" gorm:"default:false"`
	CreatedBy uint           `json:"created_by" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (Announcement) TableName() string {
	return "announcements"
}
