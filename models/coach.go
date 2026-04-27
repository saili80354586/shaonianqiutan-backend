package models

import (
	"time"

	"gorm.io/gorm"
)

// CoachLicenseType 教练执照类型
type CoachLicenseType string

const (
	CoachLicenseA  CoachLicenseType = "A级"
	CoachLicenseB  CoachLicenseType = "B级"
	CoachLicenseC  CoachLicenseType = "C级"
	CoachLicenseD  CoachLicenseType = "D级"
	CoachLicenseE  CoachLicenseType = "E级"
	CoachLicenseUE CoachLicenseType = "UEFA"
)

// Coach 教练模型
type Coach struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"index;not null"`
	LicenseType    string         `json:"license_type" gorm:"size:20"` // A级/B级/C级/D级/E级/UEFA
	LicenseNumber  string         `json:"license_number" gorm:"size:50"`
	Specialties    string         `json:"specialties" gorm:"type:text"`   // JSON数组: ["技术训练", "青少年培养"]
	Style          string         `json:"style" gorm:"type:text"`          // 执教风格: ["技术型", "战术型", "体能型", "心理型", "综合型", "青训专长型"]
	AgeGroups      string         `json:"age_groups" gorm:"type:text"`     // 擅长年龄段: ["U6", "U8", "U10", "U12", "U14", "U16", "U18", "成年队"]
	Bio            string         `json:"bio" gorm:"type:text"`
	CoachingYears  int            `json:"coaching_years" gorm:"default:0"`
	CurrentClub    string         `json:"current_club" gorm:"size:100"`
	City           string         `json:"city" gorm:"size:50"` // 常驻城市
	Verified       bool           `json:"verified" gorm:"default:false"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (Coach) TableName() string {
	return "coaches"
}

// FootballStage 足球经历阶段
type FootballStage string

const (
	StagePrimary      FootballStage = "primary"      // 小学
	StageMiddle       FootballStage = "middle"      // 初中
	StageHigh         FootballStage = "high"        // 高中
	StageUniversity   FootballStage = "university"   // 大学
	StageProfessional FootballStage = "professional"  // 职业队
)

// StageNameMap 阶段显示名称映射
var StageNameMap = map[FootballStage]string{
	StagePrimary:      "小学阶段",
	StageMiddle:       "初中阶段",
	StageHigh:         "高中阶段",
	StageUniversity:   "大学阶段",
	StageProfessional: "职业队阶段",
}

// FootballExperience 足球经历模型
type FootballExperience struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CoachID   uint           `json:"coach_id" gorm:"index;not null"`
	Stage     string         `json:"stage" gorm:"size:20;not null"`   // primary/middle/high/university/professional
	TeamName  string         `json:"team_name" gorm:"size:100"`      // 球队名称
	Position  string         `json:"position" gorm:"size:50"`       // 场上位置
	StartYear int            `json:"start_year" gorm:"not null"`     // 开始年份
	EndYear   int            `json:"end_year"`                       // 结束年份，0表示至今
	Level     string         `json:"level" gorm:"size:50"`           // 校队/区队/市队/省队/职业
	Honors    string         `json:"honors" gorm:"type:text"`        // 主要荣誉
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Coach *Coach `json:"coach,omitempty" gorm:"foreignKey:CoachID"`
}

// TableName 表名
func (FootballExperience) TableName() string {
	return "football_experiences"
}

// CoachFollowPlayer 教练关注球员关联模型
type CoachFollowPlayer struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CoachID    uint           `json:"coach_id" gorm:"index;not null"`
	UserID     uint           `json:"user_id" gorm:"index;not null"` // 球员用户ID
	IsStarred  bool           `json:"is_starred" gorm:"default:false"`
	Notes      string         `json:"notes" gorm:"type:text"`
	FollowedAt time.Time      `json:"followed_at" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Coach *Coach `json:"coach,omitempty" gorm:"foreignKey:CoachID"`
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (CoachFollowPlayer) TableName() string {
	return "coach_follow_players"
}

// TrainingNote 训练笔记模型
type TrainingNote struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CoachID     uint           `json:"coach_id" gorm:"index;not null"`
	PlayerID    uint           `json:"player_id" gorm:"index;not null"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Category    string         `json:"category" gorm:"size:20"` // 技术/战术/体能/心理
	Tags        string         `json:"tags" gorm:"type:text"`     // JSON数组
	Rating      int            `json:"rating" gorm:"default:0"`   // 1-5星
	IsPublic    bool           `json:"is_public" gorm:"default:false"`
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Coach *Coach `json:"coach,omitempty" gorm:"foreignKey:CoachID"`
	Player *User `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
}

// TableName 表名
func (TrainingNote) TableName() string {
	return "training_notes"
}

// GetSpecialtiesArray 获取专长列表
func (c *Coach) GetSpecialtiesArray() []string {
	var specialties []string
	if c.Specialties != "" {
		// JSON unmarshal would be done in service layer
	}
	return specialties
}