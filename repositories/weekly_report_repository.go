package repositories

import (
	"github.com/shaonianqiutan/backend/models"

	"gorm.io/gorm"
)

// WeeklyReportRepository 周报仓库
type WeeklyReportRepository struct {
	db *gorm.DB
}

// NewWeeklyReportRepository 创建周报仓库
func NewWeeklyReportRepository(db *gorm.DB) *WeeklyReportRepository {
	return &WeeklyReportRepository{db: db}
}

// Create 创建周报
func (r *WeeklyReportRepository) Create(report *models.WeeklyReport) error {
	return r.db.Create(report).Error
}

// Update 更新周报
func (r *WeeklyReportRepository) Update(report *models.WeeklyReport) error {
	return r.db.Save(report).Error
}

// GetByID 根据ID获取周报
func (r *WeeklyReportRepository) GetByID(id uint) (*models.WeeklyReport, error) {
	var report models.WeeklyReport
	err := r.db.Preload("Player").Preload("Coach").Preload("Team").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetByPlayerAndWeek 获取球员指定周的周报
func (r *WeeklyReportRepository) GetByPlayerAndWeek(playerID uint, weekStart string) (*models.WeeklyReport, error) {
	var report models.WeeklyReport
	err := r.db.Where("player_id = ? AND week_start = ?", playerID, weekStart).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// ListByPlayer 列出球员的所有周报
func (r *WeeklyReportRepository) ListByPlayer(playerID uint, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	var reports []models.WeeklyReport
	var total int64

	query := r.db.Model(&models.WeeklyReport{}).Where("player_id = ?", playerID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Coach").Preload("Team").
		Order("week_start DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error

	return reports, total, err
}

// ListByTeam 列出球队的所有周报(教练查看)
func (r *WeeklyReportRepository) ListByTeam(teamID uint, status string, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	var reports []models.WeeklyReport
	var total int64

	query := r.db.Model(&models.WeeklyReport{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("review_status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Player").Preload("Coach").
		Order("week_start DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error

	return reports, total, err
}

// ListPendingByCoach 列出教练待审核的周报
func (r *WeeklyReportRepository) ListPendingByCoach(coachID uint, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	var reports []models.WeeklyReport
	var total int64

	query := r.db.Model(&models.WeeklyReport{}).
		Where("coach_id = ? AND review_status = ?", coachID, "pending")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Player").Preload("Team").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error

	return reports, total, err
}

// CountPending 待审核数量
func (r *WeeklyReportRepository) CountPending(teamID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.WeeklyReport{}).
		Where("team_id = ? AND review_status = ?", teamID, "pending").
		Count(&count).Error
	return count, err
}

// Delete 删除周报
func (r *WeeklyReportRepository) Delete(id uint) error {
	return r.db.Delete(&models.WeeklyReport{}, id).Error
}

// GetTeamReports 获取球队的所有周报
func (r *WeeklyReportRepository) GetTeamReports(teamID uint, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	return r.ListByTeam(teamID, "", page, pageSize)
}
