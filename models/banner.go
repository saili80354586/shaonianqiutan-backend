package models

import (
	"time"

	"gorm.io/gorm"
)

// Banner 轮播图模型
type Banner struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"size:200"`
	ImageURL  string         `json:"image_url" gorm:"size:500;not null"`
	LinkURL   string         `json:"link_url" gorm:"size:500"`
	Position  string         `json:"position" gorm:"size:50;default:'home'"` // home/dashboard/popup
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	StartAt   *time.Time     `json:"start_at"`
	EndAt     *time.Time     `json:"end_at"`
	CreatedBy uint           `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (Banner) TableName() string {
	return "banners"
}

// BannerRepository 轮播图数据访问层
type BannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

// Create 创建轮播图
func (r *BannerRepository) Create(banner *Banner) error {
	return r.db.Create(banner).Error
}

// FindByID 根据ID查询
func (r *BannerRepository) FindByID(id uint) (*Banner, error) {
	var b Banner
	result := r.db.First(&b, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &b, nil
}

// FindAll 获取轮播图列表
func (r *BannerRepository) FindAll(page, pageSize int, position string, enabled *bool) ([]Banner, int64, error) {
	var list []Banner
	var total int64

	query := r.db.Model(&Banner{}).Order("sort_order ASC, created_at DESC")
	if position != "" {
		query = query.Where("position = ?", position)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新轮播图
func (r *BannerRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Banner{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除轮播图
func (r *BannerRepository) Delete(id uint) error {
	return r.db.Delete(&Banner{}, id).Error
}
