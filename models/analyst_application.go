package models

import (
	"time"

	"gorm.io/gorm"
)

// ApplicationStatus 申请状态
type ApplicationStatus string

const (
	// ApplicationStatusPending 待审核
	ApplicationStatusPending ApplicationStatus = "pending"
	// ApplicationStatusApproved 已批准
	ApplicationStatusApproved ApplicationStatus = "approved"
	// ApplicationStatusRejected 已拒绝
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

// AnalystApplication 分析师申请模型
type AnalystApplication struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	UserID     uint              `json:"user_id" gorm:"not null;uniqueIndex"`
	Name       string            `json:"name" gorm:"not null"`
	Phone      string            `json:"phone" gorm:"not null"`
	Email      string            `json:"email"`
	Experience string            `json:"experience" gorm:"not null"`
	Resume     string            `json:"resume"` // 简历URL
	Status     ApplicationStatus `json:"status" gorm:"default:pending"`
	Remark     string            `json:"remark"` // 审核备注
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// AnalystApplicationRepository 分析师申请仓库
type AnalystApplicationRepository struct {
	db *gorm.DB
}

func NewAnalystApplicationRepository(db *gorm.DB) *AnalystApplicationRepository {
	return &AnalystApplicationRepository{db: db}
}

// Create 创建申请
func (r *AnalystApplicationRepository) Create(app *AnalystApplication) error {
	return r.db.Create(app).Error
}

// FindByID 根据ID查找申请
func (r *AnalystApplicationRepository) FindByID(id uint) (*AnalystApplication, error) {
	var app AnalystApplication
	if err := r.db.First(&app, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

// FindByUserID 根据用户ID查找申请
func (r *AnalystApplicationRepository) FindByUserID(userID uint) (*AnalystApplication, error) {
	var app AnalystApplication
	if err := r.db.Where("user_id = ?", userID).First(&app).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

// FindAll 获取所有申请（分页）
func (r *AnalystApplicationRepository) FindAll(page, pageSize int, status *ApplicationStatus) ([]AnalystApplication, int64, error) {
	var apps []AnalystApplication
	var total int64

	query := r.db.Model(&AnalystApplication{}).Order("created_at DESC")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&apps).Error
	return apps, total, err
}

// Update 更新申请
func (r *AnalystApplicationRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&AnalystApplication{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateStatus 更新申请状态
func (r *AnalystApplicationRepository) UpdateStatus(id uint, status ApplicationStatus, remark string) error {
	updates := map[string]interface{}{
		"status": status,
		"remark": remark,
	}
	return r.db.Model(&AnalystApplication{}).Where("id = ?", id).Updates(updates).Error
}

// CountByStatus 按状态统计申请数量
func (r *AnalystApplicationRepository) CountByStatus(status ApplicationStatus) (int64, error) {
	var count int64
	err := r.db.Model(&AnalystApplication{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
