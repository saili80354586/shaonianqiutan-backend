package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// ScoutService 球探服务
type ScoutService struct {
	db *gorm.DB
}

// NewScoutService 创建球探服务
func NewScoutService(db *gorm.DB) *ScoutService {
	return &ScoutService{db: db}
}

// GetOrCreateScout 获取或创建球探资料
func (s *ScoutService) GetOrCreateScout(userID uint) (*models.Scout, error) {
	var scout models.Scout
	result := s.db.Where("user_id = ?", userID).First(&scout)
	if result.Error == nil {
		return &scout, nil
	}
	if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	// 创建新的球探记录
	scout = models.Scout{
		UserID:             userID,
		ScoutingExperience: "",
		Specialties:        "[]",
		PreferredAgeGroups:  "[]",
		ScoutingRegions:    "[]",
		CurrentOrganization: "",
		Bio:                "",
		Verified:           false,
		TotalDiscovered:    0,
		TotalReports:        0,
		TotalAdopted:       0,
	}
	if err := s.db.Create(&scout).Error; err != nil {
		return nil, err
	}
	return &scout, nil
}

// GetScoutProfile 获取球探资料
func (s *ScoutService) GetScoutProfile(userID uint) (*models.Scout, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}
	// 加载关联的User信息
	s.db.Preload("User").First(&scout, scout.ID)
	return scout, nil
}

// UpdateScoutProfile 更新球探资料
func (s *ScoutService) UpdateScoutProfile(userID uint, data map[string]interface{}) (*models.Scout, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	// 允许更新的字段
	allowedFields := []string{
		"scouting_experience", "specialties", "preferred_age_groups",
		"scouting_regions", "current_organization", "bio",
	}

	updateData := make(map[string]interface{})
	for _, field := range allowedFields {
		if val, ok := data[field]; ok {
			updateData[field] = val
		}
	}

	if err := s.db.Model(scout).Updates(updateData).Error; err != nil {
		return nil, err
	}

	return s.GetScoutProfile(userID)
}

// GetScoutDashboard 获取球探工作台数据
func (s *ScoutService) GetScoutDashboard(userID uint) (map[string]interface{}, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	// 统计
	var totalReports int64
	var publishedReports int64
	var adoptedReports int64
	s.db.Model(&models.ScoutReport{}).Where("scout_id = ?", scout.ID).Count(&totalReports)
	s.db.Model(&models.ScoutReport{}).Where("scout_id = ? AND status = ?", scout.ID, "published").Count(&publishedReports)
	s.db.Model(&models.ScoutReport{}).Where("scout_id = ? AND status = ?", scout.ID, "adopted").Count(&adoptedReports)

	// 关注的球员数
	var followedPlayers int64
	s.db.Model(&models.ScoutFollowPlayer{}).Where("scout_id = ?", scout.ID).Count(&followedPlayers)

	// 最近报告
	var recentReports []models.ScoutReport
	s.db.Preload("Player").Where("scout_id = ?", scout.ID).
		Order("created_at DESC").Limit(5).Find(&recentReports)

	// 可接任务
	var openTasks []models.ScoutTask
	s.db.Where("status = ?", "open").Order("deadline ASC").Limit(5).Find(&openTasks)

	return map[string]interface{}{
		"total_discovered":    scout.TotalDiscovered,
		"total_reports":      totalReports,
		"published_reports":   publishedReports,
		"adopted_reports":    adoptedReports,
		"followed_players":   followedPlayers,
		"recent_reports":     recentReports,
		"available_tasks":    openTasks,
	}, nil
}

// GetFollowedPlayers 获取关注的球员列表
func (s *ScoutService) GetFollowedPlayers(userID uint, page, pageSize int) ([]models.ScoutFollowPlayer, int64, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	var follows []models.ScoutFollowPlayer

	s.db.Model(&models.ScoutFollowPlayer{}).Where("scout_id = ?", scout.ID).Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.Preload("User").Where("scout_id = ?", scout.ID).
		Order("followed_at DESC").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

// FollowPlayer 关注球员
func (s *ScoutService) FollowPlayer(userID, playerID uint) (*models.ScoutFollowPlayer, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否已关注
	var existing models.ScoutFollowPlayer
	result := s.db.Where("scout_id = ? AND user_id = ?", scout.ID, playerID).First(&existing)
	if result.Error == nil {
		return &existing, nil // 已存在
	}

	follow := models.ScoutFollowPlayer{
		ScoutID:    scout.ID,
		UserID:     playerID,
		IsStarred:  false,
		Notes:      "",
		FollowedAt: time.Now(),
	}
	if err := s.db.Create(&follow).Error; err != nil {
		return nil, err
	}

	// 更新球探的发掘球员总数
	s.db.Model(scout).Update("total_discovered", gorm.Expr("total_discovered + 1"))

	return &follow, nil
}

// UnfollowPlayer 取消关注球员
func (s *ScoutService) UnfollowPlayer(userID, playerID uint) error {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return err
	}

	result := s.db.Where("scout_id = ? AND user_id = ?", scout.ID, playerID).Delete(&models.ScoutFollowPlayer{})
	if result.Error == nil && result.RowsAffected > 0 {
		// 更新球探的发掘球员总数
		s.db.Model(scout).Update("total_discovered", gorm.Expr("total_discovered - 1"))
	}
	return result.Error
}

// GetScoutReports 获取球探报告列表
func (s *ScoutService) GetScoutReports(userID uint, status string, page, pageSize int) ([]models.ScoutReport, int64, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	var reports []models.ScoutReport
	query := s.db.Model(&models.ScoutReport{}).Where("scout_id = ?", scout.ID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.Preload("Player").Where("scout_id = ?", scout.ID).
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetScoutReport 获取单个球探报告
func (s *ScoutService) GetScoutReport(userID, reportID uint) (*models.ScoutReport, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	var report models.ScoutReport
	if err := s.db.Preload("Player").Preload("Scout").Where("id = ? AND scout_id = ?", reportID, scout.ID).First(&report).Error; err != nil {
		return nil, err
	}

	// 增加浏览次数
	s.db.Model(&report).Update("views", gorm.Expr("views + 1"))

	return &report, nil
}

// CreateScoutReport 创建球探报告
func (s *ScoutService) CreateScoutReport(userID uint, data map[string]interface{}) (*models.ScoutReport, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	report := models.ScoutReport{
		ScoutID:         scout.ID,
		OverallRating:   75,
		PotentialRating: "B",
		Status:          "draft",
		Strengths:       "[]",
		Weaknesses:      "[]",
		TechnicalSkills: "{}",
	}

	// 设置字段
	if val, ok := data["player_id"]; ok {
		if playerID, ok := val.(float64); ok {
			report.PlayerID = uint(playerID)
		}
	}
	if val, ok := data["overall_rating"]; ok {
		if rating, ok := val.(float64); ok {
			report.OverallRating = int(rating)
		}
	}
	if val, ok := data["potential_rating"]; ok {
		if rating, ok := val.(string); ok {
			report.PotentialRating = rating
		}
	}
	if val, ok := data["strengths"]; ok {
		if str, ok := val.(string); ok {
			report.Strengths = str
		} else if arr, ok := val.([]interface{}); ok {
			bytes, _ := json.Marshal(arr)
			report.Strengths = string(bytes)
		}
	}
	if val, ok := data["weaknesses"]; ok {
		if str, ok := val.(string); ok {
			report.Weaknesses = str
		} else if arr, ok := val.([]interface{}); ok {
			bytes, _ := json.Marshal(arr)
			report.Weaknesses = string(bytes)
		}
	}
	if val, ok := data["technical_skills"]; ok {
		if str, ok := val.(string); ok {
			report.TechnicalSkills = str
		} else if m, ok := val.(map[string]interface{}); ok {
			bytes, _ := json.Marshal(m)
			report.TechnicalSkills = string(bytes)
		}
	}
	if val, ok := data["summary"]; ok {
		if summary, ok := val.(string); ok {
			report.Summary = summary
		}
	}
	if val, ok := data["recommendation"]; ok {
		if rec, ok := val.(string); ok {
			report.Recommendation = rec
		}
	}
	if val, ok := data["target_club"]; ok {
		if club, ok := val.(string); ok {
			report.TargetClub = club
		}
	}

	if err := s.db.Create(&report).Error; err != nil {
		return nil, err
	}

	// 更新球探的报告总数
	s.db.Model(scout).Update("total_reports", gorm.Expr("total_reports + 1"))

	return &report, nil
}

// UpdateScoutReport 更新球探报告
func (s *ScoutService) UpdateScoutReport(userID, reportID uint, data map[string]interface{}) (*models.ScoutReport, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	var report models.ScoutReport
	if err := s.db.Where("id = ? AND scout_id = ?", reportID, scout.ID).First(&report).Error; err != nil {
		return nil, err
	}

	// 允许更新的字段
	allowedFields := []string{
		"overall_rating", "potential_rating", "strengths", "weaknesses",
		"technical_skills", "summary", "recommendation", "target_club",
	}

	updateData := make(map[string]interface{})
	for _, field := range allowedFields {
		if val, ok := data[field]; ok {
			switch v := val.(type) {
			case float64:
				updateData[field] = int(v)
			case string:
				updateData[field] = v
			case []interface{}:
				bytes, _ := json.Marshal(v)
				updateData[field] = string(bytes)
			case map[string]interface{}:
				bytes, _ := json.Marshal(v)
				updateData[field] = string(bytes)
			}
		}
	}

	if err := s.db.Model(&report).Updates(updateData).Error; err != nil {
		return nil, err
	}

	return s.GetScoutReport(userID, reportID)
}

// PublishScoutReport 发布球探报告
func (s *ScoutService) PublishScoutReport(userID, reportID uint) (*models.ScoutReport, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	var report models.ScoutReport
	if err := s.db.Where("id = ? AND scout_id = ?", reportID, scout.ID).First(&report).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.db.Model(&report).Updates(map[string]interface{}{
		"status":       "published",
		"published_at":  &now,
	}).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

// DeleteScoutReport 删除球探报告
func (s *ScoutService) DeleteScoutReport(userID, reportID uint) error {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return err
	}

	result := s.db.Where("id = ? AND scout_id = ?", reportID, scout.ID).Delete(&models.ScoutReport{})
	if result.Error == nil && result.RowsAffected > 0 {
		// 更新球探的报告总数
		s.db.Model(scout).Update("total_reports", gorm.Expr("total_reports - 1"))
	}
	return result.Error
}

// GetScoutTasks 获取可接球探任务
func (s *ScoutService) GetScoutTasks(status string, page, pageSize int) ([]models.ScoutTask, int64, error) {
	var total int64
	var tasks []models.ScoutTask
	query := s.db.Model(&models.ScoutTask{})

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "open")
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// AcceptScoutTask 接取球探任务
func (s *ScoutService) AcceptScoutTask(userID, taskID uint) (*models.ScoutTask, error) {
	scout, err := s.GetOrCreateScout(userID)
	if err != nil {
		return nil, err
	}

	var task models.ScoutTask
	if err := s.db.Where("id = ? AND status = ?", taskID, "open").First(&task).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&task).Updates(map[string]interface{}{
		"status":      "accepted",
		"accepted_by": scout.ID,
	}).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// SearchPlayers 搜索球员（供球探使用）
func (s *ScoutService) SearchPlayers(keyword string, position, region string, page, pageSize int) ([]models.Player, int64, error) {
	var players []models.Player
	var total int64

	query := s.db.Model(&models.Player{})

	if keyword != "" {
		query = query.Where("nickname LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}
	if region != "" {
		query = query.Where("province LIKE ?", "%"+region+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&players).Error; err != nil {
		return nil, 0, err
	}

	return players, total, nil
}

// GetScoutPublicProfile 获取球探公开主页数据
type ScoutPublicProfile struct {
	Scout         *ScoutPublicInfo  `json:"scout"`
	User          *models.User      `json:"user,omitempty"`
	Stats         ScoutPublicStats  `json:"stats"`
	SampleReports []ScoutPublicReport `json:"sample_reports"`
}

// ScoutPublicInfo API响应专用的球探信息结构
type ScoutPublicInfo struct {
	ID                  uint     `json:"id"`
	UserID              uint     `json:"user_id"`
	ScoutingExperience  string   `json:"scouting_experience"`
	Specialties         []string `json:"specialties"`
	PreferredAgeGroups  []string `json:"preferred_age_groups"`
	ScoutingRegions     []string `json:"scouting_regions"`
	CurrentOrganization string   `json:"current_organization"`
	Bio                 string   `json:"bio"`
	Verified            bool     `json:"verified"`
	TotalDiscovered     int      `json:"total_discovered"`
	TotalReports        int      `json:"total_reports"`
	TotalAdopted        int      `json:"total_adopted"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type ScoutPublicStats struct {
	TotalDiscovered  int64 `json:"total_discovered"`
	TotalReports    int64 `json:"total_reports"`
	PublishedReports int64 `json:"published_reports"`
	FollowedPlayers  int64 `json:"followed_players"`
}

type ScoutPublicReport struct {
	ID               uint   `json:"id"`
	PlayerName       string `json:"player_name"`
	OverallRating    int    `json:"overall_rating"`
	PotentialRating string `json:"potential_rating"`
	Title            string `json:"title"`
	CreatedAt        string `json:"created_at"`
}

func (s *ScoutService) GetScoutPublicProfile(scoutID uint) (*ScoutPublicProfile, error) {
	// 直接用 scout_id 查询，不能调用 GetOrCreateScout（那个按 user_id 查询）
	var scout models.Scout
	if err := s.db.Preload("User").First(&scout, scoutID).Error; err != nil {
		return nil, errors.New("球探不存在")
	}

	// 统计
	var totalReports int64
	var publishedReports int64
	s.db.Model(&models.ScoutReport{}).Where("scout_id = ?", scout.ID).Count(&totalReports)
	s.db.Model(&models.ScoutReport{}).Where("scout_id = ? AND status = ?", scout.ID, "published").Count(&publishedReports)

	var followedPlayers int64
	s.db.Model(&models.ScoutFollowPlayer{}).Where("scout_id = ?", scout.ID).Count(&followedPlayers)

	// 示例报告
	var reports []models.ScoutReport
	s.db.Preload("Player").
		Where("scout_id = ? AND status = ?", scout.ID, "published").
		Order("created_at DESC").
		Limit(3).
		Find(&reports)

	sampleReports := make([]ScoutPublicReport, 0, len(reports))
	for _, r := range reports {
		playerName := ""
		if r.Player != nil {
			playerName = r.Player.Nickname
			if playerName == "" {
				playerName = r.Player.Name
			}
		}
		sampleReports = append(sampleReports, ScoutPublicReport{
			ID:              r.ID,
			PlayerName:       playerName,
			OverallRating:   r.OverallRating,
			PotentialRating: r.PotentialRating,
			Title:           r.Summary,
			CreatedAt:       r.CreatedAt.Format("2006-01-02"),
		})
	}

	// 解析 JSON 字符串字段
	specialties := parseJSONArray(scout.Specialties)
	preferredAgeGroups := parseJSONArray(scout.PreferredAgeGroups)
	scoutingRegions := parseJSONArray(scout.ScoutingRegions)

	return &ScoutPublicProfile{
		Scout: &ScoutPublicInfo{
			ID:                  scout.ID,
			UserID:              scout.UserID,
			ScoutingExperience:  scout.ScoutingExperience,
			Specialties:         specialties,
			PreferredAgeGroups:  preferredAgeGroups,
			ScoutingRegions:     scoutingRegions,
			CurrentOrganization: scout.CurrentOrganization,
			Bio:                 scout.Bio,
			Verified:            scout.Verified,
			TotalDiscovered:     scout.TotalDiscovered,
			TotalReports:        scout.TotalReports,
			TotalAdopted:        scout.TotalAdopted,
			CreatedAt:           scout.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:           scout.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		User:  scout.User,
		Stats: ScoutPublicStats{
			TotalDiscovered:  int64(scout.TotalDiscovered),
			TotalReports:     totalReports,
			PublishedReports: publishedReports,
			FollowedPlayers: followedPlayers,
		},
		SampleReports: sampleReports,
	}, nil
}

// parseJSONArray 解析 JSON 字符串为字符串数组
func parseJSONArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return []string{}
	}
	return result
}
