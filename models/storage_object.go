package models

import (
	"time"

	"gorm.io/gorm"
)

type StorageObjectStatus string

const (
	StorageObjectStatusPendingUpload StorageObjectStatus = "pending_upload"
	StorageObjectStatusActive        StorageObjectStatus = "active"
	StorageObjectStatusPendingDelete StorageObjectStatus = "pending_delete"
	StorageObjectStatusDeleted       StorageObjectStatus = "deleted"
	StorageObjectStatusFailed        StorageObjectStatus = "failed"
)

const (
	StorageOwnerOrder         = "order"
	StorageOwnerVideoAnalysis = "video_analysis"

	StorageBusinessOrderSourceVideo = "order_source_video"
	StorageBusinessAnalysisClip     = "analysis_clip"
)

type StorageObject struct {
	ID           uint                `json:"id" gorm:"primaryKey"`
	Bucket       string              `json:"bucket" gorm:"size:128;not null;index:idx_storage_bucket_key,unique"`
	Region       string              `json:"region" gorm:"size:64;not null"`
	ObjectKey    string              `json:"object_key" gorm:"size:700;not null;index:idx_storage_bucket_key,unique"`
	OriginalName string              `json:"original_name" gorm:"size:255"`
	ContentType  string              `json:"content_type" gorm:"size:120"`
	Size         int64               `json:"size"`
	ETag         string              `json:"etag" gorm:"size:128"`
	OwnerType    string              `json:"owner_type" gorm:"size:50;index:idx_storage_owner"`
	OwnerID      uint                `json:"owner_id" gorm:"index:idx_storage_owner"`
	BusinessType string              `json:"business_type" gorm:"size:80;index"`
	Status       StorageObjectStatus `json:"status" gorm:"size:30;index"`
	DeleteAfter  *time.Time          `json:"delete_after" gorm:"index"`
	DeletedAtTS  *time.Time          `json:"deleted_at"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	DeletedAt    gorm.DeletedAt      `json:"-" gorm:"index"`
}

type StorageObjectRepository struct {
	db *gorm.DB
}

func NewStorageObjectRepository(db *gorm.DB) *StorageObjectRepository {
	return &StorageObjectRepository{db: db}
}

func (r *StorageObjectRepository) Create(object *StorageObject) error {
	return r.db.Create(object).Error
}

func (r *StorageObjectRepository) FindByID(id uint) (*StorageObject, error) {
	var object StorageObject
	if err := r.db.First(&object, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &object, nil
}

func (r *StorageObjectRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&StorageObject{}).Where("id = ?", id).Updates(updates).Error
}
