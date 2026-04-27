package repositories

import (
	"strconv"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// MatchSummaryRepository 比赛总结仓库
type MatchSummaryRepository struct {
	db *gorm.DB
}

// NewMatchSummaryRepository 创建比赛总结仓库
func NewMatchSummaryRepository(db *gorm.DB) *MatchSummaryRepository {
	return &MatchSummaryRepository{db: db}
}

// Create 创建比赛总结
func (r *MatchSummaryRepository) Create(summary *models.MatchSummary) error {
	return r.db.Create(summary).Error
}

// Update 更新比赛总结
func (r *MatchSummaryRepository) Update(summary *models.MatchSummary) error {
	return r.db.Save(summary).Error
}

// GetByID 根据ID获取
func (r *MatchSummaryRepository) GetByID(id uint) (*models.MatchSummary, error) {
	var summary models.MatchSummary
	err := r.db.Preload("Team").Preload("Coach").
		First(&summary, id).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListByTeam 列出球队的所有比赛总结(支持状态筛选)
func (r *MatchSummaryRepository) ListByTeam(teamID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	var summaries []models.MatchSummary
	var total int64

	query := r.db.Model(&models.MatchSummary{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Coach").
		Order("match_date DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&summaries).Error

	return summaries, total, err
}

// ListByPlayer 列出球员参与的比赛总结
func (r *MatchSummaryRepository) ListByPlayer(playerID uint, page, pageSize int) ([]models.MatchSummary, int64, error) {
	var summaries []models.MatchSummary
	var total int64

	// 子查询：获取球员所属的球队ID
	teamSubQuery := r.db.Model(&models.TeamPlayer{}).
		Select("team_players.team_id").
		Joins("JOIN users ON users.id = team_players.user_id").
		Where("team_players.user_id = ? AND team_players.status = ?", playerID, "active")

	// 精确匹配：比赛总结的 PlayerIDs 包含该球员，或老数据（PlayerIDs 为空）按球队关联
	query := r.db.Model(&models.MatchSummary{}).
		Where(
			r.db.Where("team_id IN (?)", teamSubQuery).
				Where(
					r.db.Where("player_ids = ?", "").
						Or("player_ids IS NULL").
						Or("player_ids LIKE ?", "%"+strconv.FormatUint(uint64(playerID), 10)+"%"),
				),
		)

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Coach").Preload("Team").
		Order("match_date DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&summaries).Error

	return summaries, total, err
}

// ListByCoach 列出教练发起的比赛总结(支持状态筛选)
func (r *MatchSummaryRepository) ListByCoach(coachID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	var summaries []models.MatchSummary
	var total int64

	query := r.db.Model(&models.MatchSummary{}).Where("coach_id = ?", coachID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Team").
		Order("match_date DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&summaries).Error

	return summaries, total, err
}

// ListByClub 列出俱乐部所有球队的比赛总结
func (r *MatchSummaryRepository) ListByClub(clubID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	var summaries []models.MatchSummary
	var total int64

	// 子查询：俱乐部下的球队ID
	teamSubQuery := r.db.Model(&models.Team{}).
		Select("id").
		Where("club_id = ?", clubID)

	query := r.db.Model(&models.MatchSummary{}).
		Where("team_id IN (?)", teamSubQuery)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Team").Preload("Coach").
		Order("match_date DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&summaries).Error

	return summaries, total, err
}

// CountPending 统计待处理数量
func (r *MatchSummaryRepository) CountPending(teamID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.MatchSummary{}).
		Where("team_id = ? AND status IN ?", teamID, []string{"pending", "player_submitted"}).
		Count(&count).Error
	return count, err
}

// CountByStatus 统计各状态数量
func (r *MatchSummaryRepository) CountByStatus(clubID uint) (map[string]int64, error) {
	teamSubQuery := r.db.Model(&models.Team{}).
		Select("id").
		Where("club_id = ?", clubID)

	type statusCount struct {
		Status string
		Count  int64
	}
	var results []statusCount
	err := r.db.Model(&models.MatchSummary{}).
		Select("status, count(*) as count").
		Where("team_id IN (?)", teamSubQuery).
		Group("status").
		Find(&results).Error

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, err
}

// CountByResult 统计各结果数量
func (r *MatchSummaryRepository) CountByResult(clubID uint) (map[string]int64, error) {
	teamSubQuery := r.db.Model(&models.Team{}).
		Select("id").
		Where("club_id = ?", clubID)

	type resultCount struct {
		Result string
		Count  int64
	}
	var results []resultCount
	err := r.db.Model(&models.MatchSummary{}).
		Select("result, count(*) as count").
		Where("team_id IN (?)", teamSubQuery).
		Group("result").
		Find(&results).Error

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Result] = r.Count
	}
	return counts, err
}

// Delete 删除比赛总结
func (r *MatchSummaryRepository) Delete(id uint) error {
	return r.db.Delete(&models.MatchSummary{}, id).Error
}
