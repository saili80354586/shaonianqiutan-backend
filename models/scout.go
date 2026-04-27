package models

import (
	"time"

	"gorm.io/gorm"
)

// Scout 球探模型
type Scout struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	UserID             uint           `json:"user_id" gorm:"index;not null"`
	ScoutingExperience string         `json:"scouting_experience" gorm:"size:20"` // 球探年限: 0-1/1-3/3-5/5-10/10+
	Specialties        string         `json:"specialties" gorm:"type:text"`        // JSON数组: ["前锋", "中场"]
	PreferredAgeGroups string         `json:"preferred_age_groups" gorm:"type:text"` // JSON数组: ["U12", "U14"]
	ScoutingRegions    string         `json:"scouting_regions" gorm:"type:text"`    // JSON数组: ["华东", "华北"]
	CurrentOrganization string         `json:"current_organization" gorm:"size:100"` // 所属机构
	Bio                string         `json:"bio" gorm:"type:text"`                 // 个人简介
	Verified           bool           `json:"verified" gorm:"default:false"`
	TotalDiscovered    int            `json:"total_discovered" gorm:"default:0"`  // 发掘球员总数
	TotalReports       int            `json:"total_reports" gorm:"default:0"`     // 报告总数
	TotalAdopted       int            `json:"total_adopted" gorm:"default:0"`     // 被采纳报告数
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (Scout) TableName() string {
	return "scouts"
}

// ScoutFollowPlayer 球探关注球员关联模型
type ScoutFollowPlayer struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ScoutID     uint           `json:"scout_id" gorm:"index;not null"`
	UserID      uint           `json:"user_id" gorm:"index;not null"` // 球员用户ID
	IsStarred   bool           `json:"is_starred" gorm:"default:false"`
	Notes       string         `json:"notes" gorm:"type:text"` // 备注
	FollowedAt  time.Time     `json:"followed_at" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Scout *Scout `json:"scout,omitempty" gorm:"foreignKey:ScoutID"`
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (ScoutFollowPlayer) TableName() string {
	return "scout_follow_players"
}

// ScoutReport 球探报告模型
type ScoutReport struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ScoutID         uint           `json:"scout_id" gorm:"index;not null"`
	PlayerID        uint           `json:"player_id" gorm:"index;not null"` // 被评估的球员ID
	OverallRating   int            `json:"overall_rating" gorm:"default:0"`    // 综合评分 1-100
	PotentialRating string         `json:"potential_rating" gorm:"size:5"`     // 潜力评级: S/A/B/C/D
	Status          string         `json:"status" gorm:"size:20;default:'draft'"` // 状态: draft/published/adopted
	Strengths       string         `json:"strengths" gorm:"type:text"`          // JSON数组: 优势
	Weaknesses      string         `json:"weaknesses" gorm:"type:text"`         // JSON数组: 劣势
	TechnicalSkills string         `json:"technical_skills" gorm:"type:text"`    // JSON: {shooting,passing,dribbling,defending,physical,mentality}
	Summary         string         `json:"summary" gorm:"type:text"`            // 总评
	Recommendation  string         `json:"recommendation" gorm:"type:text"`     // 发展建议
	TargetClub      string         `json:"target_club" gorm:"size:100"`         // 目标俱乐部
	Views           int            `json:"views" gorm:"default:0"`               // 浏览次数
	Likes           int            `json:"likes" gorm:"default:0"`              // 点赞次数
	PublishedAt     *time.Time     `json:"published_at"`                         // 发布时间
	AdoptedAt       *time.Time     `json:"adopted_at"`                         // 被采纳时间
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Scout  *Scout  `json:"scout,omitempty" gorm:"foreignKey:ScoutID"`
	Player *Player  `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
}

// TableName 表名
func (ScoutReport) TableName() string {
	return "scout_reports"
}

// ScoutTask 球探任务模型
type ScoutTask struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`     // 任务标题
	Description string         `json:"description" gorm:"type:text"`         // 任务描述
	Region      string         `json:"region" gorm:"size:50"`               // 区域
	AgeGroup    string         `json:"age_group" gorm:"size:20"`           // 目标年龄段
	Reward      int            `json:"reward" gorm:"default:0"`            // 奖励金额
	Status      string         `json:"status" gorm:"size:20;default:'open'"` // 状态: open/accepted/completed/closed
	Deadline    time.Time      `json:"deadline"`                            // 截止时间
	AcceptedBy  *uint          `json:"accepted_by"`                         // 接取者ID
	CompletedAt *time.Time     `json:"completed_at"`                         // 完成时间
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	AcceptedByScout *Scout `json:"accepted_by_scout,omitempty" gorm:"foreignKey:AcceptedBy"`
}

// TableName 表名
func (ScoutTask) TableName() string {
	return "scout_tasks"
}
