package models

import (
	"time"

	"gorm.io/gorm"
)

// HelpGuide 站内使用指南模型
type HelpGuide struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Role      string         `json:"role" gorm:"size:50;not null;index"`
	Title     string         `json:"title" gorm:"size:200;not null"`
	Summary   string         `json:"summary" gorm:"type:text"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	ViewCount int            `json:"view_count" gorm:"default:0"`
	CreatedBy uint           `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (HelpGuide) TableName() string {
	return "help_guides"
}

type HelpGuideRepository struct {
	db *gorm.DB
}

func NewHelpGuideRepository(db *gorm.DB) *HelpGuideRepository {
	return &HelpGuideRepository{db: db}
}

func (r *HelpGuideRepository) Create(guide *HelpGuide) error {
	return r.db.Create(guide).Error
}

func (r *HelpGuideRepository) FindAll(page, pageSize int, role string, enabled *bool) ([]HelpGuide, int64, error) {
	var list []HelpGuide
	var total int64

	query := r.db.Model(&HelpGuide{}).Order("sort_order ASC, created_at DESC")
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *HelpGuideRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&HelpGuide{}).Where("id = ?", id).Updates(updates).Error
}

func (r *HelpGuideRepository) Delete(id uint) error {
	return r.db.Delete(&HelpGuide{}, id).Error
}
