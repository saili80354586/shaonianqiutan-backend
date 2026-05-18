package models

import (
	"time"

	"gorm.io/gorm"
)

// AnalystStatus 分析师状态
type AnalystStatus string

const (
	AnalystStatusActive    AnalystStatus = "active"
	AnalystStatusInactive  AnalystStatus = "inactive"
	AnalystStatusSuspended AnalystStatus = "suspended"
)

// Analyst 分析师模型
type Analyst struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	UserID       uint          `json:"user_id" gorm:"not null;uniqueIndex"`
	User         User          `json:"user" gorm:"foreignKey:UserID"`
	Name         string        `json:"name" gorm:"size:50;not null"`
	Bio          string        `json:"bio" gorm:"type:text"`
	Specialty    string        `json:"specialty" gorm:"size:255"`          // 擅长领域
	Experience   int           `json:"experience"`                         // 从业年限
	Profession   string        `json:"profession" gorm:"size:50"`          // 职业背景
	IsProPlayer  bool          `json:"is_pro_player" gorm:"default:false"` // 是否有职业球员经历
	HasCase      bool          `json:"has_case" gorm:"default:false"`      // 是否有分析案例
	CaseDetail   string        `json:"case_detail" gorm:"type:text"`       // 案例说明
	ContactPhone string        `json:"contact_phone" gorm:"size:20"`       // 联系电话
	ContactEmail string        `json:"contact_email" gorm:"size:100"`      // 联系邮箱
	Rating       float64       `json:"rating" gorm:"default:0"`            // 评分
	ReviewCount  int           `json:"review_count" gorm:"default:0"`      // 评论数
	Status       AnalystStatus `json:"status" gorm:"size:20;default:'active'"`

	// 官方选题与等级治理字段
	LevelCode              string                         `json:"level_code" gorm:"size:20;default:'L1';index"`
	LevelUpdatedAt         *time.Time                     `json:"level_updated_at"`
	LevelUpdatedBy         uint                           `json:"level_updated_by" gorm:"default:0"`
	LevelNote              string                         `json:"level_note" gorm:"size:500"`
	OfficialAdoptionCount  int                            `json:"official_adoption_count" gorm:"default:0"`
	OfficialPublishCount   int                            `json:"official_publish_count" gorm:"default:0"`
	OfficialMaterialCount  int                            `json:"official_material_count" gorm:"default:0"`
	IsOfficialPartner      bool                           `json:"is_official_partner" gorm:"default:false;index"`
	PartnershipStartedAt   *time.Time                     `json:"partnership_started_at"`
	PartnershipUpdatedBy   uint                           `json:"partnership_updated_by" gorm:"default:0"`
	PartnershipNote        string                         `json:"partnership_note" gorm:"size:500"`
	PartnershipBenefits    string                         `json:"partnership_benefits" gorm:"type:text"`
	QualityScore           float64                        `json:"quality_score" gorm:"type:decimal(5,2);default:0"`
	DeliveryScore          float64                        `json:"delivery_score" gorm:"type:decimal(5,2);default:0"`
	ContentScore           float64                        `json:"content_score" gorm:"type:decimal(5,2);default:0"`
	BusinessScore          float64                        `json:"business_score" gorm:"type:decimal(5,2);default:0"`
	GrowthScore            float64                        `json:"growth_score" gorm:"type:decimal(5,2);default:0"`
	GrowthSnapshot         *AnalystGrowthSnapshot         `json:"growth_snapshot,omitempty" gorm:"foreignKey:AnalystID"`
	OfficialPublishMetrics *OfficialPublishMetricsSummary `json:"official_publish_metrics,omitempty" gorm:"-"`
	CreatedAt              time.Time                      `json:"created_at"`
	UpdatedAt              time.Time                      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt                 `json:"-" gorm:"index"`
}

// OfficialPublishMetricsSummary 是官方发布记录在分析师维度的只读运营汇总。
type OfficialPublishMetricsSummary struct {
	PublishRecordCount int64 `json:"publish_record_count"`
	TotalPlayCount     int64 `json:"total_play_count"`
	TotalLikeCount     int64 `json:"total_like_count"`
	TotalCommentCount  int64 `json:"total_comment_count"`
	TotalShareCount    int64 `json:"total_share_count"`
	TotalFavoriteCount int64 `json:"total_favorite_count"`
	MaxPlayCount       int64 `json:"max_play_count"`
}

// AnalystRepository 分析师仓库
type AnalystRepository struct {
	db *gorm.DB
}

// NewAnalystRepository 创建分析师仓库
func NewAnalystRepository(db *gorm.DB) *AnalystRepository {
	return &AnalystRepository{db: db}
}

// FindByID 根据ID查找分析师
func (r *AnalystRepository) FindByID(id uint) (*Analyst, error) {
	var analyst Analyst
	if err := r.db.Preload("User").First(&analyst, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &analyst, nil
}

// FindByUserID 根据用户ID查找分析师
func (r *AnalystRepository) FindByUserID(userID uint) (*Analyst, error) {
	var analyst Analyst
	if err := r.db.Preload("User").Where("user_id = ?", userID).First(&analyst).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &analyst, nil
}

// FindAll 查找所有活跃分析师
func (r *AnalystRepository) FindAll(page, pageSize int) ([]Analyst, int64, error) {
	var analysts []Analyst
	var total int64

	// 先统计总数
	if err := r.db.Model(&Analyst{}).Where("status = ?", AnalystStatusActive).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Where("status = ?", AnalystStatusActive).
		Order("rating DESC, created_at DESC").
		Preload("User").
		Offset(offset).
		Limit(pageSize).
		Find(&analysts).Error
	return analysts, total, err
}

// Create 创建分析师
func (r *AnalystRepository) Create(analyst *Analyst) error {
	return r.db.Create(analyst).Error
}

// Update 更新分析师
func (r *AnalystRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Analyst{}).Where("id = ?", id).Updates(updates).Error
}

// GetTopByOrders 获取Top分析师（按订单数）
func (r *AnalystRepository) GetTopByOrders(limit int) ([]Analyst, error) {
	var analysts []Analyst
	// 简化实现：按创建时间排序，实际应按订单数统计
	err := r.db.Order("created_at DESC").Limit(limit).Find(&analysts).Error
	return analysts, err
}
