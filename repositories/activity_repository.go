package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// ActivityRepository 动态/活动仓储
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建动态仓储
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// GetActivitiesByTeamID 获取球队动态列表
func (r *ActivityRepository) GetActivitiesByTeamID(teamID uint, limit int) ([]models.TeamDynamic, error) {
	var activities []models.TeamDynamic
	err := r.db.Where("team_id = ?", teamID).Order("created_at DESC").Limit(limit).Find(&activities).Error
	return activities, err
}
