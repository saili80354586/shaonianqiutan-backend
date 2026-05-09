package models

import (
	"time"

	"gorm.io/gorm"
)

// RoleApplicationStatus 追加身份申请状态
type RoleApplicationStatus string

const (
	RoleApplicationStatusPending   RoleApplicationStatus = "pending"
	RoleApplicationStatusApproved  RoleApplicationStatus = "approved"
	RoleApplicationStatusRejected  RoleApplicationStatus = "rejected"
	RoleApplicationStatusSuspended RoleApplicationStatus = "suspended"
)

// RoleApplication 用于记录已有账号申请追加业务身份。
type RoleApplication struct {
	ID           uint                  `json:"id" gorm:"primaryKey"`
	UserID       uint                  `json:"user_id" gorm:"index;not null"`
	User         User                  `json:"user" gorm:"foreignKey:UserID"`
	Role         UserRole              `json:"role" gorm:"size:20;index;not null"`
	Status       RoleApplicationStatus `json:"status" gorm:"size:20;index;default:'pending'"`
	Source       string                `json:"source" gorm:"size:50;default:'self_apply'"`
	ProfileJSON  string                `json:"profile_json" gorm:"type:text"`
	RejectReason string                `json:"reject_reason" gorm:"type:text"`
	ReviewedBy   uint                  `json:"reviewed_by"`
	ReviewedAt   *time.Time            `json:"reviewed_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	DeletedAt    gorm.DeletedAt        `json:"-" gorm:"index"`
}
