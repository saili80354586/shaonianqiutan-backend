package models

import (
	"time"

	"gorm.io/gorm"
)

// PlatformAnnouncement 平台公告模型
type PlatformAnnouncement struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"size:200;not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"size:20;default:'notice'"` // notice/activity/update
	IsPinned  bool           `json:"is_pinned" gorm:"default:false"`
	IsPublic  bool           `json:"is_public" gorm:"default:true"`
	StartAt   *time.Time     `json:"start_at"`
	EndAt     *time.Time     `json:"end_at"`
	CreatedBy uint           `json:"created_by" gorm:"not null"`
	AuthorName string        `json:"author_name" gorm:"size:64"`
	ViewCount int            `json:"view_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (PlatformAnnouncement) TableName() string {
	return "platform_announcements"
}

// PlatformAnnouncementRepository 平台公告数据访问层
type PlatformAnnouncementRepository struct {
	db *gorm.DB
}

func NewPlatformAnnouncementRepository(db *gorm.DB) *PlatformAnnouncementRepository {
	return &PlatformAnnouncementRepository{db: db}
}

// Create 创建公告
func (r *PlatformAnnouncementRepository) Create(announcement *PlatformAnnouncement) error {
	return r.db.Create(announcement).Error
}

// FindByID 根据ID查询
func (r *PlatformAnnouncementRepository) FindByID(id uint) (*PlatformAnnouncement, error) {
	var a PlatformAnnouncement
	result := r.db.First(&a, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &a, nil
}

// FindAll 获取公告列表
func (r *PlatformAnnouncementRepository) FindAll(page, pageSize int, annType string, pinned *bool) ([]PlatformAnnouncement, int64, error) {
	var list []PlatformAnnouncement
	var total int64

	query := r.db.Model(&PlatformAnnouncement{}).Order("is_pinned DESC, created_at DESC")
	if annType != "" {
		query = query.Where("type = ?", annType)
	}
	if pinned != nil {
		query = query.Where("is_pinned = ?", *pinned)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新公告
func (r *PlatformAnnouncementRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&PlatformAnnouncement{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除公告
func (r *PlatformAnnouncementRepository) Delete(id uint) error {
	return r.db.Delete(&PlatformAnnouncement{}, id).Error
}

// IncrementViewCount 增加浏览量
func (r *PlatformAnnouncementRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&PlatformAnnouncement{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
