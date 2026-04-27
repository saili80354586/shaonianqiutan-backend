package repositories

import (
	"gorm.io/gorm"
)

// PhysicalTestRepository 体测仓储
type PhysicalTestRepository struct {
	db *gorm.DB
}

// NewPhysicalTestRepository 创建体测仓储
func NewPhysicalTestRepository(db *gorm.DB) *PhysicalTestRepository {
	return &PhysicalTestRepository{db: db}
}
