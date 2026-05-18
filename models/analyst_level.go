package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const DefaultAnalystLevelCode = "L1"

// AnalystLevel 分析师成长等级字典。
type AnalystLevel struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	Code                string    `json:"code" gorm:"uniqueIndex;size:20;not null"`
	Name                string    `json:"name" gorm:"size:50;not null"`
	Description         string    `json:"description" gorm:"type:text"`
	PriorityWeight      int       `json:"priority_weight" gorm:"default:0"`
	DailyTaskLimit      int       `json:"daily_task_limit" gorm:"default:1"`
	OfficialTaskVisible bool      `json:"official_task_visible" gorm:"default:true"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// AnalystLevelApplicationStatus 分析师等级申请状态。
type AnalystLevelApplicationStatus string

const (
	AnalystLevelApplicationPending  AnalystLevelApplicationStatus = "pending"
	AnalystLevelApplicationApproved AnalystLevelApplicationStatus = "approved"
	AnalystLevelApplicationAdjusted AnalystLevelApplicationStatus = "adjusted"
	AnalystLevelApplicationRejected AnalystLevelApplicationStatus = "rejected"
)

// AnalystLevelApplication 分析师初始/晋级等级申请。
type AnalystLevelApplication struct {
	ID                 uint                          `json:"id" gorm:"primaryKey"`
	AnalystID          uint                          `json:"analyst_id" gorm:"not null;index"`
	Analyst            *Analyst                      `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	RequestedLevelCode string                        `json:"requested_level_code" gorm:"size:20;not null;index"`
	ApplicationReason  string                        `json:"application_reason" gorm:"type:text"`
	ExperienceSummary  string                        `json:"experience_summary" gorm:"type:text"`
	CaseMaterials      string                        `json:"case_materials" gorm:"type:text"`
	Specialties        string                        `json:"specialties" gorm:"size:255"`
	SelfAssessment     string                        `json:"self_assessment" gorm:"type:text"`
	Status             AnalystLevelApplicationStatus `json:"status" gorm:"size:20;not null;default:'pending';index"`
	ReviewedLevelCode  string                        `json:"reviewed_level_code" gorm:"size:20"`
	ReviewNote         string                        `json:"review_note" gorm:"size:500"`
	ReviewedBy         uint                          `json:"reviewed_by" gorm:"default:0;index"`
	ReviewedAt         *time.Time                    `json:"reviewed_at"`
	CreatedAt          time.Time                     `json:"created_at"`
	UpdatedAt          time.Time                     `json:"updated_at"`
}

// AnalystGrowthSnapshot 分析师成长分当前快照，用于系统建议等级。
type AnalystGrowthSnapshot struct {
	ID                    uint       `json:"id" gorm:"primaryKey"`
	AnalystID             uint       `json:"analyst_id" gorm:"not null;uniqueIndex"`
	Analyst               *Analyst   `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	QualityScore          float64    `json:"quality_score" gorm:"type:decimal(5,2);default:0"`
	DeliveryScore         float64    `json:"delivery_score" gorm:"type:decimal(5,2);default:0"`
	ContentScore          float64    `json:"content_score" gorm:"type:decimal(5,2);default:0"`
	BusinessScore         float64    `json:"business_score" gorm:"type:decimal(5,2);default:0"`
	GrowthScore           float64    `json:"growth_score" gorm:"type:decimal(5,2);default:0"`
	SuggestedLevelCode    string     `json:"suggested_level_code" gorm:"size:20;index"`
	SuggestionReason      string     `json:"suggestion_reason" gorm:"size:500"`
	NextLevelCode         string     `json:"next_level_code" gorm:"size:20"`
	NextLevelGap          float64    `json:"next_level_gap" gorm:"type:decimal(5,2);default:0"`
	OfficialSubmissionNum int        `json:"official_submission_num" gorm:"default:0"`
	OfficialApprovedNum   int        `json:"official_approved_num" gorm:"default:0"`
	OfficialAdoptionNum   int        `json:"official_adoption_num" gorm:"default:0"`
	PaidCompletedNum      int        `json:"paid_completed_num" gorm:"default:0"`
	SuggestionStatus      string     `json:"suggestion_status" gorm:"size:20;default:'pending';index"`
	SuggestionReviewedBy  uint       `json:"suggestion_reviewed_by" gorm:"default:0"`
	SuggestionReviewedAt  *time.Time `json:"suggestion_reviewed_at"`
	SuggestionReviewNote  string     `json:"suggestion_review_note" gorm:"size:500"`
	CalculatedAt          time.Time  `json:"calculated_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// AnalystLevelHistory 记录等级变更与系统建议采纳/忽略历史。
type AnalystLevelHistory struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnalystID          uint      `json:"analyst_id" gorm:"not null;index"`
	Analyst            *Analyst  `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	FromLevelCode      string    `json:"from_level_code" gorm:"size:20"`
	ToLevelCode        string    `json:"to_level_code" gorm:"size:20"`
	SuggestedLevelCode string    `json:"suggested_level_code" gorm:"size:20"`
	Source             string    `json:"source" gorm:"size:30;not null;index"`
	Action             string    `json:"action" gorm:"size:20;not null;index"`
	Note               string    `json:"note" gorm:"size:500"`
	OperatorID         uint      `json:"operator_id" gorm:"default:0;index"`
	CreatedAt          time.Time `json:"created_at"`
}

type DefaultAnalystLevel struct {
	Code                string
	Name                string
	Description         string
	PriorityWeight      int
	DailyTaskLimit      int
	OfficialTaskVisible bool
}

func DefaultAnalystLevels() []DefaultAnalystLevel {
	return []DefaultAnalystLevel{
		{Code: "L1", Name: "见习分析师", Description: "新入驻或样例单验证阶段", PriorityWeight: 10, DailyTaskLimit: 1, OfficialTaskVisible: true},
		{Code: "L2", Name: "认证分析师", Description: "通过平台基础审核，能稳定完成报告", PriorityWeight: 20, DailyTaskLimit: 2, OfficialTaskVisible: true},
		{Code: "L3", Name: "优选分析师", Description: "交付稳定、报告质量较好", PriorityWeight: 30, DailyTaskLimit: 3, OfficialTaskVisible: true},
		{Code: "L4", Name: "专家分析师", Description: "高质量、高准时率、有采用记录", PriorityWeight: 40, DailyTaskLimit: 5, OfficialTaskVisible: true},
		{Code: "L5", Name: "官方合作分析师", Description: "长期稳定合作，内容被官方多次采用", PriorityWeight: 50, DailyTaskLimit: 8, OfficialTaskVisible: true},
	}
}

// SeedDefaultAnalystLevels 幂等补齐默认等级，不覆盖已存在等级配置。
func SeedDefaultAnalystLevels(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	for _, item := range DefaultAnalystLevels() {
		level := AnalystLevel{
			Code:                item.Code,
			Name:                item.Name,
			Description:         item.Description,
			PriorityWeight:      item.PriorityWeight,
			DailyTaskLimit:      item.DailyTaskLimit,
			OfficialTaskVisible: item.OfficialTaskVisible,
		}
		if err := db.Where("code = ?", item.Code).FirstOrCreate(&level).Error; err != nil {
			return err
		}
	}

	return nil
}

func AnalystLevelRank(code string) int {
	switch code {
	case "L1":
		return 1
	case "L2":
		return 2
	case "L3":
		return 3
	case "L4":
		return 4
	case "L5":
		return 5
	default:
		return 0
	}
}

func AnalystLevelMeets(current, required string) bool {
	if required == "" {
		required = DefaultAnalystLevelCode
	}
	return AnalystLevelRank(current) >= AnalystLevelRank(required)
}
