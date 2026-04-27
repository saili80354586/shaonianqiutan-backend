package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// PlayerReviewRepository 球员自评仓库
type PlayerReviewRepository struct {
	db *gorm.DB
}

// NewPlayerReviewRepository 创建球员自评仓库
func NewPlayerReviewRepository(db *gorm.DB) *PlayerReviewRepository {
	return &PlayerReviewRepository{db: db}
}

// Create 创建球员自评
func (r *PlayerReviewRepository) Create(review *models.PlayerReview) error {
	return r.db.Create(review).Error
}

// Update 更新球员自评
func (r *PlayerReviewRepository) Update(review *models.PlayerReview) error {
	return r.db.Save(review).Error
}

// GetByID 根据ID获取
func (r *PlayerReviewRepository) GetByID(id uint) (*models.PlayerReview, error) {
	var review models.PlayerReview
	err := r.db.Preload("Player").
		First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// GetByMatchAndPlayer 根据比赛ID和球员ID获取（唯一记录）
func (r *PlayerReviewRepository) GetByMatchAndPlayer(matchID, playerID uint) (*models.PlayerReview, error) {
	var review models.PlayerReview
	err := r.db.Preload("Player").
		Where("match_id = ? AND player_id = ?", matchID, playerID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// ListByMatch 获取比赛的所有球员自评
func (r *PlayerReviewRepository) ListByMatch(matchID uint) ([]models.PlayerReview, error) {
	var reviews []models.PlayerReview
	err := r.db.Preload("Player").
		Where("match_id = ?", matchID).
		Order("submitted_at ASC").
		Find(&reviews).Error
	return reviews, err
}

// ListByPlayer 获取球员的所有自评（分页）
func (r *PlayerReviewRepository) ListByPlayer(playerID uint, page, pageSize int) ([]models.PlayerReview, int64, error) {
	var reviews []models.PlayerReview
	var total int64

	query := r.db.Model(&models.PlayerReview{}).Where("player_id = ?", playerID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Player").Preload("Match").
		Order("submitted_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&reviews).Error

	return reviews, total, err
}

// CountByMatch 统计比赛的已提交自评数
func (r *PlayerReviewRepository) CountByMatch(matchID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.PlayerReview{}).
		Where("match_id = ?", matchID).
		Count(&count).Error
	return count, err
}

// HasPlayerSubmitted 检查球员是否已提交自评
func (r *PlayerReviewRepository) HasPlayerSubmitted(matchID, playerID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PlayerReview{}).
		Where("match_id = ? AND player_id = ?", matchID, playerID).
		Count(&count).Error
	return count > 0, err
}

// DeleteByMatch 删除比赛的所有自评
func (r *PlayerReviewRepository) DeleteByMatch(matchID uint) error {
	return r.db.Where("match_id = ?", matchID).Delete(&models.PlayerReview{}).Error
}

// Delete 删除自评
func (r *PlayerReviewRepository) Delete(id uint) error {
	return r.db.Delete(&models.PlayerReview{}, id).Error
}
