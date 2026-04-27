package models

import (
	"time"
)

// GrowthRecordType 成长记录类型
type GrowthRecordType string

const (
	GrowthRecordTypeMilestone GrowthRecordType = "milestone" // 里程碑
	GrowthRecordTypeAchievement GrowthRecordType = "achievement" // 成就
	GrowthRecordTypeTraining GrowthRecordType = "training" // 训练
	GrowthRecordTypeMatch GrowthRecordType = "match" // 比赛
	GrowthRecordTypePhysical GrowthRecordType = "physical" // 体测
)

// GrowthRecord 球员成长记录模型
type GrowthRecord struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	UserID     uint            `json:"user_id" gorm:"index;not null"` // 球员用户ID
	RecordDate time.Time       `json:"record_date" gorm:"not null"`    // 记录日期
	RecordType GrowthRecordType `json:"record_type" gorm:"size:20;not null"` // 记录类型
	Title      string          `json:"title" gorm:"size:200;not null"` // 标题
	Content    string          `json:"content" gorm:"type:text"`       // 内容
	StatsJSON  string          `json:"stats_json" gorm:"type:json"`   // 统计数据JSON
	CreatedAt  time.Time       `json:"created_at"`
	User       *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (GrowthRecord) TableName() string {
	return "growth_records"
}