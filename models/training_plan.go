package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TrainingPlanStatus 训练计划状态
type TrainingPlanStatus string

const (
	TrainingPlanStatusDraft     TrainingPlanStatus = "draft"
	TrainingPlanStatusPublished TrainingPlanStatus = "published"
	TrainingPlanStatusCompleted TrainingPlanStatus = "completed"
	TrainingPlanStatusCancelled TrainingPlanStatus = "cancelled"
)

// TrainingPlan 训练计划模型
type TrainingPlan struct {
	ID           uint               `json:"id" gorm:"primaryKey"`
	ClubID       uint               `json:"club_id" gorm:"index;not null"`
	TeamID       uint               `json:"team_id" gorm:"index;not null"`
	Title        string             `json:"title" gorm:"size:200;not null"`
	Theme        string             `json:"theme" gorm:"size:100"` // 训练主题
	Location     string             `json:"location" gorm:"size:200"`
	StartTime    time.Time          `json:"start_time" gorm:"not null"`
	EndTime      *time.Time         `json:"end_time"`
	PlayerIDs    string             `json:"player_ids" gorm:"type:text"`    // JSON数组，参与球员ID列表
	Content      string             `json:"content" gorm:"type:text"`       // 训练内容
	VideoURLs    string             `json:"video_urls" gorm:"type:text"`    // JSON数组，视频URL列表
	Summary      string             `json:"summary" gorm:"type:text"`       // 训练总结
	CoachID          uint               `json:"coach_id" gorm:"index"`          // 负责教练
	WeeklyReportID   *uint              `json:"weekly_report_id" gorm:"index"`  // 关联周报周期
	PhysicalTestID   *uint              `json:"physical_test_id" gorm:"index"`  // 关联体测活动
	Status           TrainingPlanStatus `json:"status" gorm:"size:20;default:'draft'"`
	RemindSent       bool               `json:"remind_sent" gorm:"default:false"` // 是否已发送提醒
	CreatedBy        uint               `json:"created_by" gorm:"not null"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	DeletedAt        gorm.DeletedAt     `json:"-" gorm:"index"`

	// 关联
	Club         *Club         `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Team         *Team         `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	Coach        *User         `json:"coach,omitempty" gorm:"foreignKey:CoachID"`
	Player       *User         `json:"player,omitempty" gorm:"foreignKey:CreatedBy"`
	WeeklyReport *WeeklyReport `json:"weekly_report,omitempty" gorm:"foreignKey:WeeklyReportID"`
	PhysicalTest *PhysicalTestActivity `json:"physical_test,omitempty" gorm:"foreignKey:PhysicalTestID"`
}

// TableName 表名
func (TrainingPlan) TableName() string {
	return "training_plans"
}

// GetPlayerIDs 获取参与球员ID列表
func (t *TrainingPlan) GetPlayerIDs() []uint {
	if t.PlayerIDs == "" {
		return []uint{}
	}
	var ids []uint
	json.Unmarshal([]byte(t.PlayerIDs), &ids)
	return ids
}

// GetVideoURLs 获取视频URL列表
func (t *TrainingPlan) GetVideoURLs() []string {
	if t.VideoURLs == "" {
		return []string{}
	}
	var urls []string
	json.Unmarshal([]byte(t.VideoURLs), &urls)
	return urls
}
