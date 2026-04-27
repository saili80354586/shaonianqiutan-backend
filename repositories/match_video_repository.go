package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// MatchVideoRepository 比赛视频仓库
type MatchVideoRepository struct {
	db *gorm.DB
}

// NewMatchVideoRepository 创建比赛视频仓库
func NewMatchVideoRepository(db *gorm.DB) *MatchVideoRepository {
	return &MatchVideoRepository{db: db}
}

// Create 创建视频链接
func (r *MatchVideoRepository) Create(video *models.MatchVideo) error {
	return r.db.Create(video).Error
}

// Update 更新视频链接
func (r *MatchVideoRepository) Update(video *models.MatchVideo) error {
	return r.db.Save(video).Error
}

// GetByID 根据ID获取
func (r *MatchVideoRepository) GetByID(id uint) (*models.MatchVideo, error) {
	var video models.MatchVideo
	err := r.db.First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// ListByMatch 获取比赛的所有视频链接
func (r *MatchVideoRepository) ListByMatch(matchID uint) ([]models.MatchVideo, error) {
	var videos []models.MatchVideo
	err := r.db.Where("match_id = ? AND status = ?", matchID, "active").
		Order("sort_order ASC, created_at ASC").
		Find(&videos).Error
	return videos, err
}

// Delete 删除视频链接（软删除，设置status=deleted）
func (r *MatchVideoRepository) Delete(id uint) error {
	return r.db.Model(&models.MatchVideo{}).
		Where("id = ?", id).
		Update("status", "deleted").Error
}

// HardDelete 硬删除视频链接
func (r *MatchVideoRepository) HardDelete(id uint) error {
	return r.db.Delete(&models.MatchVideo{}, id).Error
}

// DeleteByMatch 删除比赛的所有视频链接
func (r *MatchVideoRepository) DeleteByMatch(matchID uint) error {
	return r.db.Where("match_id = ?", matchID).Delete(&models.MatchVideo{}).Error
}

// CountByMatch 统计比赛的视频数
func (r *MatchVideoRepository) CountByMatch(matchID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.MatchVideo{}).
		Where("match_id = ? AND status = ?", matchID, "active").
		Count(&count).Error
	return count, err
}
