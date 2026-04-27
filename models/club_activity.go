package models

import (
	"encoding/json"
	"time"
)

// ClubActivity 俱乐部活动
type ClubActivity struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	ClubID          uint      `json:"club_id" gorm:"index;not null"`
	Title           string    `json:"title" gorm:"size:200;not null"`
	Type            string    `json:"type" gorm:"size:50"`      // external: 外部活动(足球嘉年华), internal: 内部活动(球员团建)
	Status          string    `json:"status" gorm:"size:50"`    // upcoming: 即将开始, ongoing: 进行中, ended: 已结束
	Description     string    `json:"description" gorm:"type:text"`
	CoverImage      string    `json:"coverImage" gorm:"size:500"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Location        string    `json:"location" gorm:"size:300"`
	MaxParticipants int       `json:"maxParticipants"`
	ContactPhone    string    `json:"contactPhone" gorm:"size:50"`
	ContactWechat   string    `json:"contactWechat" gorm:"size:100"`
	PublishStatus   string    `json:"publishStatus" gorm:"size:50;default:'published'"` // draft, published, unpublished
	IsReview        bool      `json:"isReview" gorm:"default:false"`           // 是否作为回顾展示
	ReviewContent   string    `json:"reviewContent" gorm:"type:text"`          // 回顾内容
	ReviewImages    string    `json:"reviewImages" gorm:"type:text"`           // JSON 图片数组字符串
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ClubActivity) TableName() string {
	return "club_activities"
}

// GetReviewImagesArray 获取回顾图片数组
func (a *ClubActivity) GetReviewImagesArray() []string {
	if a.ReviewImages == "" {
		return []string{}
	}
	var images []string
	json.Unmarshal([]byte(a.ReviewImages), &images)
	return images
}

// SetReviewImagesArray 设置回顾图片数组
func (a *ClubActivity) SetReviewImagesArray(images []string) {
	b, _ := json.Marshal(images)
	a.ReviewImages = string(b)
}

// ClubActivityRegistration 活动报名记录
type ClubActivityRegistration struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ActivityID uint      `json:"activity_id" gorm:"index;not null"`
	UserID     *uint     `json:"user_id" gorm:"index"` // 注册用户ID，可选
	Name       string    `json:"name" gorm:"size:100;not null"`
	Phone      string    `json:"phone" gorm:"size:50"`
	Wechat     string    `json:"wechat" gorm:"size:100"`
	Remark     string    `json:"remark" gorm:"size:300"`
	Status     string    `json:"status" gorm:"size:50;default:'pending'"` // pending, confirmed, cancelled
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (ClubActivityRegistration) TableName() string {
	return "club_activity_registrations"
}
