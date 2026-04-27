package models

import (
	"time"

	"gorm.io/gorm"
)

// ContentReportStatus 举报处理状态
type ContentReportStatus string

const (
	ContentReportStatusPending    ContentReportStatus = "pending"
	ContentReportStatusProcessing ContentReportStatus = "processing"
	ContentReportStatusResolved   ContentReportStatus = "resolved"
	ContentReportStatusRejected   ContentReportStatus = "rejected"
)

// ContentReportType 举报类型
type ContentReportType string

const (
	ContentReportTypePost    ContentReportType = "post"    // 动态
	ContentReportTypeComment ContentReportType = "comment" // 评论
	ContentReportTypeUser    ContentReportType = "user"    // 用户
	ContentReportTypeMessage ContentReportType = "message" // 私信
)

// ContentReport 内容举报模型
type ContentReport struct {
	ID           uint                `json:"id" gorm:"primaryKey"`
	ReporterID   uint                `json:"reporter_id" gorm:"index;not null"`
	ReporterName string              `json:"reporter_name" gorm:"size:64"`
	TargetID     uint                `json:"target_id" gorm:"index;not null"`
	TargetType   ContentReportType   `json:"target_type" gorm:"size:20;not null"`
	Reason       string              `json:"reason" gorm:"size:500;not null"`
	Detail       string              `json:"detail" gorm:"type:text"`
	Status       ContentReportStatus `json:"status" gorm:"size:20;default:'pending'"`
	HandlerID    uint                `json:"handler_id" gorm:"default:0"`
	HandlerName  string              `json:"handler_name" gorm:"size:64"`
	HandleResult string              `json:"handle_result" gorm:"type:text"`
	HandledAt    *time.Time          `json:"handled_at"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	DeletedAt    gorm.DeletedAt      `json:"-" gorm:"index"`
}

// TableName 表名
func (ContentReport) TableName() string {
	return "content_reports"
}

// ContentReportRepository 举报数据访问层
type ContentReportRepository struct {
	db *gorm.DB
}

func NewContentReportRepository(db *gorm.DB) *ContentReportRepository {
	return &ContentReportRepository{db: db}
}

// Create 创建举报
func (r *ContentReportRepository) Create(report *ContentReport) error {
	return r.db.Create(report).Error
}

// FindByID 根据ID查询
func (r *ContentReportRepository) FindByID(id uint) (*ContentReport, error) {
	var report ContentReport
	result := r.db.First(&report, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &report, nil
}

// FindAll 获取举报列表
func (r *ContentReportRepository) FindAll(page, pageSize int, status string) ([]ContentReport, int64, error) {
	var reports []ContentReport
	var total int64

	query := r.db.Model(&ContentReport{}).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

// UpdateStatus 更新状态
func (r *ContentReportRepository) UpdateStatus(id uint, status ContentReportStatus, handlerID uint, handlerName, result string) error {
	now := time.Now()
	return r.db.Model(&ContentReport{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": result,
		"handled_at":    &now,
	}).Error
}

// CountByStatus 按状态统计
func (r *ContentReportRepository) CountByStatus(status ContentReportStatus) (int64, error) {
	var count int64
	err := r.db.Model(&ContentReport{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
