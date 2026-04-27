package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓库
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓库
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// CreateBatch 批量创建通知
func (r *NotificationRepository) CreateBatch(notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return r.db.CreateInBatches(notifications, 100).Error
}

// GetByID 根据ID获取
func (r *NotificationRepository) GetByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// ListByUser 列出用户通知
func (r *NotificationRepository) ListByUser(userID uint, page, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&notifications).Error

	return notifications, total, err
}

// ListUnread 列出未读通知
func (r *NotificationRepository) ListUnread(userID uint, page, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&notifications).Error

	return notifications, total, err
}

// CountUnread 统计未读数量
func (r *NotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记为已读
func (r *NotificationRepository) MarkAsRead(id uint) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkAllAsRead 标记全部为已读
func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// Delete 删除通知
func (r *NotificationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Notification{}, id).Error
}

// DeleteOld 删除旧通知(超过指定天数)
func (r *NotificationRepository) DeleteOld(days int) error {
	return r.db.Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).Delete(&models.Notification{}).Error
}
