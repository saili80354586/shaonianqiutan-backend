package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// AdminOperationLogRepository 管理员操作日志仓库
type AdminOperationLogRepository struct {
	db *gorm.DB
}

// NewAdminOperationLogRepository 创建仓库
func NewAdminOperationLogRepository(db *gorm.DB) *AdminOperationLogRepository {
	return &AdminOperationLogRepository{db: db}
}

// Create 创建日志
func (r *AdminOperationLogRepository) Create(log *models.AdminOperationLog) error {
	return r.db.Create(log).Error
}

// GetLogsByClubID 获取俱乐部操作日志
func (r *AdminOperationLogRepository) GetLogsByClubID(clubID uint, page, pageSize int) ([]*models.AdminOperationLog, int64, error) {
	var logs []*models.AdminOperationLog
	var total int64

	query := r.db.Model(&models.AdminOperationLog{}).Where("club_id = ?", clubID)
	query.Count(&total)

	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
