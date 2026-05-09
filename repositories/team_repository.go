package repositories

import (
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// TeamRepository 球队数据访问层
type TeamRepository struct {
	db *gorm.DB
}

// NewTeamRepository 创建球队仓储
func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create 创建球队
func (r *TeamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

// Update 更新球队
func (r *TeamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

// Delete 删除球队
func (r *TeamRepository) Delete(id uint) error {
	return r.db.Delete(&models.Team{}, id).Error
}

// FindByID 根据ID获取球队
func (r *TeamRepository) FindByID(id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Club").First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// FindByClubID 获取俱乐部的所有球队
func (r *TeamRepository) FindByClubID(clubID uint, status string, includeDeleted bool) ([]models.Team, error) {
	var teams []models.Team
	query := r.db
	if includeDeleted {
		query = query.Unscoped()
	}
	query = query.Where("club_id = ?", clubID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("age_group ASC, created_at DESC").Find(&teams).Error
	return teams, err
}

// FindByAgeGroup 按年龄组筛选
func (r *TeamRepository) FindByAgeGroup(clubID uint, ageGroup string) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Where("club_id = ? AND age_group = ? AND status = ?", clubID, ageGroup, models.TeamStatusActive).
		Order("created_at DESC").
		Find(&teams).Error
	return teams, err
}

// AddPlayer 添加球员到球队
func (r *TeamRepository) AddPlayer(teamID, userID uint, jerseyNumber, position string) error {
	tp := &models.TeamPlayer{
		TeamID:       teamID,
		UserID:       userID,
		JerseyNumber: jerseyNumber,
		Position:     position,
		Status:       "active",
		JoinedAt:     time.Now(),
	}
	return r.db.Create(tp).Error
}

// RemovePlayer 从球队移除球员
func (r *TeamRepository) RemovePlayer(teamID, userID uint, status string) error {
	updates := map[string]interface{}{
		"status":  status, // transferred 表示转会
		"left_at": time.Now(),
	}
	return r.db.Model(&models.TeamPlayer{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Updates(updates).Error
}

// UpdatePlayer 更新球员信息
func (r *TeamRepository) UpdatePlayer(teamID, playerID uint, updates map[string]interface{}) error {
	return r.db.Model(&models.TeamPlayer{}).
		Where("team_id = ? AND id = ?", teamID, playerID).
		Updates(updates).Error
}

// GetPlayers 获取球队球员列表
func (r *TeamRepository) GetPlayers(teamID uint, status, position, keyword string) ([]models.TeamPlayer, int64, error) {
	var players []models.TeamPlayer
	var total int64

	query := r.db.Model(&models.TeamPlayer{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}
	if keyword != "" {
		query = query.Joins("JOIN users ON team_players.user_id = users.id").
			Where("users.name LIKE ? OR users.nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	err := query.Preload("User").Order("joined_at DESC").Find(&players).Error
	return players, total, err
}

// AddCoach 添加教练到球队
func (r *TeamRepository) AddCoach(teamID, userID uint, role models.CoachRole) error {
	tc := &models.TeamCoach{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		Status:   "active",
		JoinedAt: time.Now(),
	}
	return r.db.Create(tc).Error
}

// RemoveCoach 从球队移除教练
func (r *TeamRepository) RemoveCoach(teamID, userID uint) error {
	return r.db.Model(&models.TeamCoach{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Updates(map[string]interface{}{
			"status":  "inactive",
			"left_at": time.Now(),
		}).Error
}

// UpdateCoach 更新教练信息
func (r *TeamRepository) UpdateCoach(teamID, coachID uint, updates map[string]interface{}) error {
	return r.db.Model(&models.TeamCoach{}).
		Where("team_id = ? AND id = ?", teamID, coachID).
		Updates(updates).Error
}

// GetCoaches 获取球队教练列表
func (r *TeamRepository) GetCoaches(teamID uint, status string) ([]models.TeamCoach, error) {
	var coaches []models.TeamCoach
	query := r.db.Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("User").Order("role ASC, joined_at ASC").Find(&coaches).Error
	return coaches, err
}

// GetCoachTeams 获取教练关联的球队列表
func (r *TeamRepository) GetCoachTeams(userID uint) ([]models.Team, error) {
	var teamCoaches []models.TeamCoach
	err := r.db.Preload("Team").Preload("Team.Club").
		Where("user_id = ? AND status = ?", userID, "active").
		Find(&teamCoaches).Error
	if err != nil {
		return nil, err
	}

	teams := make([]models.Team, 0, len(teamCoaches))
	seen := make(map[uint]bool)
	for _, tc := range teamCoaches {
		if tc.Team != nil && !seen[tc.Team.ID] {
			teams = append(teams, *tc.Team)
			seen[tc.Team.ID] = true
		}
	}
	return teams, nil
}

// IsCoachOfTeam 检查用户是否是球队的教练
func (r *TeamRepository) IsCoachOfTeam(userID, teamID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamCoach{}).
		Where("user_id = ? AND team_id = ? AND status = ?", userID, teamID, "active").
		Count(&count).Error
	return count > 0, err
}

// CreateInvitation 创建邀请记录
func (r *TeamRepository) CreateInvitation(inv *models.TeamInvitation) error {
	return r.db.Create(inv).Error
}

// FindInvitationByCode 根据邀请码查找
func (r *TeamRepository) FindInvitationByCode(code string) (*models.TeamInvitation, error) {
	var inv models.TeamInvitation
	err := r.db.Preload("Team").Preload("Team.Club").Preload("TargetUser").Preload("Creator").
		Where("invite_code = ?", code).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// UpdateInvitationStatus 更新邀请状态
func (r *TeamRepository) UpdateInvitationStatus(id uint, status models.InvitationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.InvitationStatusAccepted {
		updates["accepted_at"] = time.Now()
	} else if status == models.InvitationStatusRejected {
		updates["rejected_at"] = time.Now()
	}
	return r.db.Model(&models.TeamInvitation{}).Where("id = ?", id).Updates(updates).Error
}

// GetInvitations 获取球队的邀请列表
func (r *TeamRepository) GetInvitations(teamID uint, status string) ([]models.TeamInvitation, error) {
	var invitations []models.TeamInvitation
	query := r.db.Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("TargetUser").Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

// SearchUsers 搜索用户（用于邀请）
func (r *TeamRepository) SearchUsers(keyword string, userType string) ([]models.User, error) {
	var users []models.User
	query := r.db.Model(&models.User{}).
		Where("phone LIKE ? OR nickname LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	if userType != "" {
		// 球员在数据库中 role 为 "user"，但前端传入 "player"，需要兼容处理
		if userType == "player" {
			query = query.Where("role IN (?)", []string{"user", "player"})
		} else if userType == string(models.RoleCoach) {
			query = query.Where(r.db.Where("role = ?", userType).Or(
				"id IN (SELECT user_id FROM user_roles WHERE role = ? AND status IN ? AND deleted_at IS NULL)",
				userType,
				[]string{"active", "approved"},
			))
		} else {
			query = query.Where("role = ?", userType)
		}
	}
	err := query.Limit(20).Find(&users).Error
	return users, err
}

// CountPlayers 统计球员数量
func (r *TeamRepository) CountPlayers(teamID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.TeamPlayer{}).
		Where("team_id = ? AND status = ?", teamID, "active").
		Count(&count).Error
	return count, err
}

// CountCoaches 统计教练数量
func (r *TeamRepository) CountCoaches(teamID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.TeamCoach{}).
		Where("team_id = ? AND status = ?", teamID, "active").
		Count(&count).Error
	return count, err
}

// GetPlayerTeam 获取用户所在的球队
func (r *TeamRepository) GetPlayerTeam(userID uint) (*models.TeamPlayer, error) {
	var tp models.TeamPlayer
	err := r.db.Preload("Team").Preload("Team.Club").
		Where("user_id = ? AND status = ?", userID, "active").
		First(&tp).Error
	if err != nil {
		return nil, err
	}
	return &tp, nil
}

// GetUserInvitations 获取用户收到的邀请（通过 userID 或 phone 匹配）
func (r *TeamRepository) GetUserInvitations(userID uint, phone string, status string) ([]models.TeamInvitation, error) {
	var invitations []models.TeamInvitation
	query := r.db.Where("target_user_id = ? OR target_phone = ?", userID, phone)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("Team").Preload("Team.Club").Preload("Creator").Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

// FindUserByPhone 根据手机号查找用户
func (r *TeamRepository) FindUserByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func (r *TeamRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetTeamPlayer 根据ID获取球队球员
func (r *TeamRepository) GetTeamPlayer(playerID uint) (*models.TeamPlayer, error) {
	var player models.TeamPlayer
	err := r.db.Preload("User").First(&player, playerID).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// ========== 入队申请 ==========

// CreateApplication 创建入队/试训申请
func (r *TeamRepository) CreateApplication(app *models.TeamApplication) error {
	return r.db.Create(app).Error
}

// GetApplications 获取球队的申请列表
func (r *TeamRepository) GetApplications(teamID uint, status string) ([]models.TeamApplication, error) {
	var apps []models.TeamApplication
	query := r.db.Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("Player").Preload("Team").Preload("Reviewer").Order("created_at DESC").Find(&apps).Error
	return apps, err
}

// GetMyApplications 获取我提交的申请列表
func (r *TeamRepository) GetMyApplications(playerID uint, status string) ([]models.TeamApplication, error) {
	var apps []models.TeamApplication
	query := r.db.Where("player_id = ?", playerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Preload("Team").Preload("Team.Club").Order("created_at DESC").Find(&apps).Error
	return apps, err
}

// GetApplicationByID 根据ID获取申请
func (r *TeamRepository) GetApplicationByID(id uint) (*models.TeamApplication, error) {
	var app models.TeamApplication
	err := r.db.Preload("Player").Preload("Team").Preload("Team.Club").First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// ReviewApplication 审核申请
func (r *TeamRepository) ReviewApplication(id uint, status string, responseNote string, reviewedBy uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"response_note": responseNote,
		"reviewed_by":   reviewedBy,
		"reviewed_at":   now,
	}
	return r.db.Model(&models.TeamApplication{}).Where("id = ?", id).Updates(updates).Error
}

// FindPendingApplication 查找待处理申请（去重检查）
func (r *TeamRepository) FindPendingApplication(teamID, playerID uint, appType string) (*models.TeamApplication, error) {
	var app models.TeamApplication
	err := r.db.Where("team_id = ? AND player_id = ? AND type = ? AND status = ?", teamID, playerID, appType, "pending").
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// ========== 通用 ==========

// UpdateTeam 更新球队信息
func (r *TeamRepository) UpdateTeam(teamID uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Team{}).Where("id = ?", teamID).Updates(updates).Error
}

// Restore 恢复软删除的球队
func (r *TeamRepository) Restore(teamID uint) error {
	return r.db.Unscoped().Model(&models.Team{}).Where("id = ?", teamID).Update("deleted_at", nil).Error
}
