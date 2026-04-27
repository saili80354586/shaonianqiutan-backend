package repositories

import (
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// CoachRepository 教练仓储层
type CoachRepository struct {
	db *gorm.DB
}

// NewCoachRepository 创建教练仓储
func NewCoachRepository(db *gorm.DB) *CoachRepository {
	return &CoachRepository{db: db}
}

// GetCoachByUserID 根据用户ID获取教练
func (r *CoachRepository) GetCoachByUserID(userID uint) (*models.Coach, error) {
	var coach models.Coach
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&coach).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}

// GetCoachByID 根据ID获取教练
func (r *CoachRepository) GetCoachByID(id uint) (*models.Coach, error) {
	var coach models.Coach
	err := r.db.Where("id = ?", id).First(&coach).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}

// CreateCoach 创建教练
func (r *CoachRepository) CreateCoach(coach *models.Coach) error {
	return r.db.Create(coach).Error
}

// UpdateCoach 更新教练
func (r *CoachRepository) UpdateCoach(coach *models.Coach) error {
	return r.db.Save(coach).Error
}

// CreateOrUpdateCoach 创建或更新教练资料
func (r *CoachRepository) CreateOrUpdateCoach(userID uint, licenseType, licenseNumber, specialties, bio string, coachingYears int, currentClub string) (*models.Coach, error) {
	coach, err := r.GetCoachByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		coach = &models.Coach{
			UserID:        userID,
			LicenseType:   licenseType,
			LicenseNumber: licenseNumber,
			Specialties:   specialties,
			Bio:           bio,
			CoachingYears: coachingYears,
			CurrentClub:   currentClub,
			Verified:      false,
		}
		err = r.CreateCoach(coach)
	} else if err != nil {
		return nil, err
	} else {
		coach.LicenseType = licenseType
		coach.LicenseNumber = licenseNumber
		coach.Specialties = specialties
		coach.Bio = bio
		coach.CoachingYears = coachingYears
		coach.CurrentClub = currentClub
		err = r.UpdateCoach(coach)
	}
	return coach, err
}

// GetFollowedPlayers 获取教练关注的球员列表
func (r *CoachRepository) GetFollowedPlayers(coachID uint, page, pageSize int) ([]models.CoachFollowPlayer, int64, error) {
	var follows []models.CoachFollowPlayer
	var total int64

	query := r.db.Model(&models.CoachFollowPlayer{}).Where("coach_id = ?", coachID)
	query.Count(&total)

	err := query.
		Preload("User").
		Order("followed_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&follows).Error

	return follows, total, err
}

// GetFollowingStatus 检查关注状态
func (r *CoachRepository) GetFollowingStatus(coachID, playerID uint) (*models.CoachFollowPlayer, error) {
	var follow models.CoachFollowPlayer
	err := r.db.Where("coach_id = ? AND user_id = ?", coachID, playerID).First(&follow).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

// FollowPlayer 关注球员
func (r *CoachRepository) FollowPlayer(coachID, playerID uint) error {
	// 检查是否已关注
	var existing models.CoachFollowPlayer
	err := r.db.Where("coach_id = ? AND user_id = ?", coachID, playerID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		follow := &models.CoachFollowPlayer{
			CoachID:    coachID,
			UserID:    playerID,
			FollowedAt: time.Now(),
		}
		return r.db.Create(follow).Error
	}
	return err
}

// UnfollowPlayer 取消关注球员
func (r *CoachRepository) UnfollowPlayer(coachID, playerID uint) error {
	return r.db.Where("coach_id = ? AND user_id = ?", coachID, playerID).Delete(&models.CoachFollowPlayer{}).Error
}

// UpdateFollowNotes 更新关注备注
func (r *CoachRepository) UpdateFollowNotes(coachID, playerID uint, notes string, isStarred bool) error {
	return r.db.Model(&models.CoachFollowPlayer{}).
		Where("coach_id = ? AND user_id = ?", coachID, playerID).
		Updates(map[string]interface{}{
			"notes":      notes,
			"is_starred": isStarred,
		}).Error
}

// GetTrainingNotes 获取训练笔记列表
func (r *CoachRepository) GetTrainingNotes(coachID uint, page, pageSize int, playerID *uint, category string) ([]models.TrainingNote, int64, error) {
	var notes []models.TrainingNote
	var total int64

	query := r.db.Model(&models.TrainingNote{}).Where("coach_id = ?", coachID)
	if playerID != nil {
		query = query.Where("player_id = ?", *playerID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)

	err := query.
		Preload("Player").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&notes).Error

	return notes, total, err
}

// GetTrainingNoteByID 获取训练笔记详情
func (r *CoachRepository) GetTrainingNoteByID(coachID, noteID uint) (*models.TrainingNote, error) {
	var note models.TrainingNote
	err := r.db.Preload("Player").Where("coach_id = ? AND id = ?", coachID, noteID).First(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// CreateTrainingNote 创建训练笔记
func (r *CoachRepository) CreateTrainingNote(note *models.TrainingNote) error {
	return r.db.Create(note).Error
}

// UpdateTrainingNote 更新训练笔记
func (r *CoachRepository) UpdateTrainingNote(note *models.TrainingNote) error {
	return r.db.Save(note).Error
}

// DeleteTrainingNote 删除训练笔记
func (r *CoachRepository) DeleteTrainingNote(coachID, noteID uint) error {
	return r.db.Where("coach_id = ? AND id = ?", coachID, noteID).Delete(&models.TrainingNote{}).Error
}

// GetPlayerProgress 获取球员进度数据
func (r *CoachRepository) GetPlayerProgress(coachID, playerID uint) ([]map[string]interface{}, error) {
	// 获取该球员被该教练关注的记录
	var follows []models.CoachFollowPlayer
	r.db.Where("coach_id = ? AND user_id = ?", coachID, playerID).Preload("User").Find(&follows)

	// 获取该球员的训练笔记
	var notes []models.TrainingNote
	r.db.Where("coach_id = ? AND player_id = ?", coachID, playerID).Order("created_at DESC").Limit(10).Find(&notes)

	// 获取该球员的报告
	var reports []models.Report
	r.db.Where("user_id = ?", playerID).Order("created_at DESC").Limit(10).Find(&reports)

	result := make([]map[string]interface{}, 0)

	// 添加笔记记录
	for _, n := range notes {
		result = append(result, map[string]interface{}{
			"type":      "note",
			"id":        n.ID,
			"date":      n.CreatedAt,
			"title":     n.Title,
			"content":   n.Content,
			"category":  n.Category,
			"rating":    n.Rating,
		})
	}

	// 添加报告记录
	for _, rep := range reports {
		result = append(result, map[string]interface{}{
			"type":   "report",
			"id":     rep.ID,
			"date":   rep.CreatedAt,
			"title":  "分析报告",
			"status": rep.Status,
		})
	}

	return result, nil
}

// GetFootballExperiences 获取足球经历列表
func (r *CoachRepository) GetFootballExperiences(coachID uint) ([]models.FootballExperience, error) {
	var experiences []models.FootballExperience
	err := r.db.Where("coach_id = ?", coachID).Order("start_year ASC").Find(&experiences).Error
	return experiences, err
}

// GetFootballExperienceByID 获取足球经历详情
func (r *CoachRepository) GetFootballExperienceByID(coachID, expID uint) (*models.FootballExperience, error) {
	var exp models.FootballExperience
	err := r.db.Where("coach_id = ? AND id = ?", coachID, expID).First(&exp).Error
	if err != nil {
		return nil, err
	}
	return &exp, nil
}

// CreateFootballExperience 创建足球经历
func (r *CoachRepository) CreateFootballExperience(exp *models.FootballExperience) error {
	return r.db.Create(exp).Error
}

// UpdateFootballExperience 更新足球经历
func (r *CoachRepository) UpdateFootballExperience(exp *models.FootballExperience) error {
	return r.db.Save(exp).Error
}

// DeleteFootballExperience 删除足球经历
func (r *CoachRepository) DeleteFootballExperience(coachID, expID uint) error {
	return r.db.Where("coach_id = ? AND id = ?", coachID, expID).Delete(&models.FootballExperience{}).Error
}