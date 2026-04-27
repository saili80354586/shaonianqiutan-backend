package models

import (
	"time"

	"gorm.io/gorm"
)

// FAQ 常见问题模型
type FAQ struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Question  string         `json:"question" gorm:"type:text;not null"`
	Answer    string         `json:"answer" gorm:"type:text;not null"`
	Category  string         `json:"category" gorm:"size:50;default:'general'"` // general/account/order/payment/report/club
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	ViewCount int            `json:"view_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (FAQ) TableName() string {
	return "faqs"
}

// FAQRepository FAQ数据访问层
type FAQRepository struct {
	db *gorm.DB
}

func NewFAQRepository(db *gorm.DB) *FAQRepository {
	return &FAQRepository{db: db}
}

// Create 创建FAQ
func (r *FAQRepository) Create(faq *FAQ) error {
	return r.db.Create(faq).Error
}

// FindByID 根据ID查询
func (r *FAQRepository) FindByID(id uint) (*FAQ, error) {
	var f FAQ
	result := r.db.First(&f, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &f, nil
}

// FindAll 获取FAQ列表
func (r *FAQRepository) FindAll(page, pageSize int, category string, enabled *bool) ([]FAQ, int64, error) {
	var list []FAQ
	var total int64

	query := r.db.Model(&FAQ{}).Order("sort_order ASC, created_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新FAQ
func (r *FAQRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&FAQ{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除FAQ
func (r *FAQRepository) Delete(id uint) error {
	return r.db.Delete(&FAQ{}, id).Error
}
