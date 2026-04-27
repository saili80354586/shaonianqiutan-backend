package models

import (
	"time"

	"gorm.io/gorm"
)

// MatchScheduleStatus 赛程状态
type MatchScheduleStatus string

const (
	MatchScheduleStatusUpcoming   MatchScheduleStatus = "upcoming"
	MatchScheduleStatusOngoing    MatchScheduleStatus = "ongoing"
	MatchScheduleStatusCompleted  MatchScheduleStatus = "completed"
	MatchScheduleStatusCancelled  MatchScheduleStatus = "cancelled"
)

// MatchScheduleType 赛事类型
type MatchScheduleType string

const (
	MatchScheduleTypeLeague     MatchScheduleType = "league"     // 联赛
	MatchScheduleTypeCup        MatchScheduleType = "cup"        // 杯赛
	MatchScheduleTypeFriendly   MatchScheduleType = "friendly"   // 友谊赛
	MatchScheduleTypeTraining   MatchScheduleType = "training_match" // 训练赛
)

// MatchSchedule 赛程日历模型
type MatchSchedule struct {
	ID          uint                `json:"id" gorm:"primaryKey"`
	ClubID      uint                `json:"club_id" gorm:"index;not null"`
	TeamID      uint                `json:"team_id" gorm:"index;not null"`
	Name        string              `json:"name" gorm:"size:200;not null"` // 赛事名称
	MatchType   MatchScheduleType   `json:"match_type" gorm:"size:20;not null"`
	Opponent    string              `json:"opponent" gorm:"size:100"`        // 对手
	MatchTime   time.Time           `json:"match_time" gorm:"not null"`
	Location    string              `json:"location" gorm:"size:200"`
	HomeScore   *int                `json:"home_score"`
	AwayScore   *int                `json:"away_score"`
	Remark      string              `json:"remark" gorm:"type:text"`
	Status      MatchScheduleStatus `json:"status" gorm:"size:20;default:'upcoming'"`
	PreRemindSent   bool            `json:"pre_remind_sent" gorm:"default:false"`    // 赛前提醒是否已发送
	PostNotifySent  bool            `json:"post_notify_sent" gorm:"default:false"`   // 赛后通知是否已发送
	MatchSummaryID  *uint           `json:"match_summary_id"`                        // 关联的比赛总结ID
	CreatedBy   uint                `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"-" gorm:"index"`

	// 关联
	Club         *Club          `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Team         *Team          `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	MatchSummary *MatchSummary  `json:"match_summary,omitempty" gorm:"foreignKey:MatchSummaryID"`
}

// TableName 表名
func (MatchSchedule) TableName() string {
	return "match_schedules"
}
