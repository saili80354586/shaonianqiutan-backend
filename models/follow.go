package models

import (
	"time"
)

// Follow 用户关注关系模型
type Follow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FollowerID  uint      `json:"follower_id" gorm:"index;not null"`  // 关注者ID
	FollowingID uint      `json:"following_id" gorm:"index;not null"`  // 被关注者ID
	CreatedAt   time.Time `json:"created_at"`
	Follower    *User     `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	Following   *User     `json:"following,omitempty" gorm:"foreignKey:FollowingID"`
}

// TableName 表名
func (Follow) TableName() string {
	return "follows"
}