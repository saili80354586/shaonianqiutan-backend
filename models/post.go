package models

import (
	"encoding/json"
	"time"
)

// Post 社交动态帖子
type Post struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"index;not null"`
	Content       string    `json:"content" gorm:"type:text;not null"`
	Images        string    `json:"images" gorm:"type:text"`           // JSON 图片数组字符串
	TargetType    string    `json:"target_type" gorm:"size:50"`       // 关联目标类型：report, player, match 等
	TargetID      uint      `json:"target_id" gorm:"index;default:0"`
	RoleTag       string    `json:"role_tag" gorm:"size:50"`          // 发帖时用户角色标签：player, parent, coach, scout, analyst, club
	LikesCount    int       `json:"likes_count" gorm:"default:0"`
	CommentsCount int       `json:"comments_count" gorm:"default:0"`
	IsTop         bool      `json:"is_top" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Post) TableName() string {
	return "posts"
}

// GetImagesArray 获取图片数组
func (p *Post) GetImagesArray() []string {
	if p.Images == "" {
		return []string{}
	}
	var images []string
	json.Unmarshal([]byte(p.Images), &images)
	return images
}

// SetImagesArray 设置图片数组
func (p *Post) SetImagesArray(images []string) {
	b, _ := json.Marshal(images)
	p.Images = string(b)
}
