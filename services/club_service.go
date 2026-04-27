package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ClubService 俱乐部服务
type ClubService struct {
	db *gorm.DB
}

// NewClubService 创建俱乐部服务
func NewClubService(db *gorm.DB) *ClubService {
	return &ClubService{db: db}
}

// GetClubByUserID 根据用户ID获取俱乐部
func (s *ClubService) GetClubByUserID(userID uint) (*models.Club, error) {
	var club models.Club
	err := s.db.Where("user_id = ?", userID).First(&club).Error
	if err != nil {
		return nil, err
	}
	return &club, nil
}

// UpdateClubProfile 更新俱乐部资料
func (s *ClubService) UpdateClubProfile(userID uint, name, logo, description, address, contactName, contactPhone string) (*models.Club, error) {
	var club models.Club
	err := s.db.Where("user_id = ?", userID).First(&club).Error
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if logo != "" {
		updates["logo"] = logo
	}
	if description != "" {
		updates["description"] = description
	}
	if address != "" {
		updates["address"] = address
	}
	if contactName != "" {
		updates["contact_name"] = contactName
	}
	if contactPhone != "" {
		updates["contact_phone"] = contactPhone
	}

	if len(updates) > 0 {
		err = s.db.Model(&club).Updates(updates).Error
		if err != nil {
			return nil, err
		}
	}

	s.db.First(&club, club.ID)
	return &club, nil
}

// GetPlayerCount 获取球员数量
func (s *ClubService) GetPlayerCount(clubID uint) (int64, error) {
	var count int64
	err := s.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND status = ?", clubID, "active").Count(&count).Error
	return count, err
}

// GetDashboardOverview 获取工作台概览数据
func (s *ClubService) GetDashboardOverview(clubID uint) (map[string]interface{}, error) {
	overview := make(map[string]interface{})

	// 球员数量
	var totalPlayers int64
	s.db.Model(&models.ClubPlayer{}).Where("club_id = ?", clubID).Count(&totalPlayers)
	overview["totalPlayers"] = totalPlayers

	var activePlayers int64
	s.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND status = ?", clubID, "active").Count(&activePlayers)
	overview["activePlayers"] = activePlayers

	// 订单统计
	var totalOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ?", clubID).Count(&totalOrders)
	overview["totalOrders"] = totalOrders

	var pendingOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ? AND status = ?", clubID, "pending").Count(&pendingOrders)
	overview["pendingOrders"] = pendingOrders

	var completedOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ? AND status = ?", clubID, "completed").Count(&completedOrders)
	overview["completedOrders"] = completedOrders

	// 体测统计
	var totalPhysicalTests int64
	s.db.Model(&models.PhysicalTestActivity{}).Where("club_id = ?", clubID).Count(&totalPlayers)
	overview["totalPhysicalTests"] = totalPhysicalTests

	var thisMonthTests int64
	s.db.Model(&models.PhysicalTestActivity{}).Where("club_id = ? AND created_at >= ?", clubID, getStartOfMonth()).Count(&thisMonthTests)
	overview["physicalTestsThisMonth"] = thisMonthTests

	// ========== 运营洞察数据 ==========
	insights := make(map[string]interface{})

	// 1. 本周周报提交率
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
	if now.Weekday() == time.Sunday {
		weekStart = now.AddDate(0, 0, -6)
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	var weeklyTotalPlayers, weeklySubmitted int
	s.db.Model(&models.WeeklyReportPeriod{}).
		Joins("JOIN teams ON teams.id = weekly_report_periods.team_id").
		Where("teams.club_id = ? AND weekly_report_periods.week_start >= ?", clubID, weekStart).
		Select("COALESCE(SUM(weekly_report_periods.total_players), 0) AS total, COALESCE(SUM(weekly_report_periods.submitted_count), 0) AS submitted").
		Row().Scan(&weeklyTotalPlayers, &weeklySubmitted)
	weeklySubmitRate := 0
	if weeklyTotalPlayers > 0 {
		weeklySubmitRate = weeklySubmitted * 100 / weeklyTotalPlayers
	}
	insights["weeklyReportSubmitRate"] = weeklySubmitRate
	insights["weeklyReportTotal"] = weeklyTotalPlayers
	insights["weeklyReportSubmitted"] = weeklySubmitted

	// 2. 待点评比赛总结数
	var pendingMatchSummaries int64
	s.db.Model(&models.MatchSummary{}).
		Joins("JOIN teams ON teams.id = match_summaries.team_id").
		Where("teams.club_id = ? AND match_summaries.status = ?", clubID, "player_submitted").
		Count(&pendingMatchSummaries)
	insights["pendingMatchSummaries"] = pendingMatchSummaries

	// 3. 待完成体测记录数（进行中体测且未录入记录的球员）
	var pendingPhysicalTestRecords int64
	s.db.Raw(`
		SELECT COALESCE(SUM(pt.player_count - COALESCE(complete_counts.completed, 0)), 0)
		FROM physical_test_activities pt
		JOIN (
			SELECT physical_test_id, COUNT(DISTINCT player_id) as completed
			FROM physical_test_records
			WHERE deleted_at IS NULL
			GROUP BY physical_test_id
		) complete_counts ON complete_counts.physical_test_id = pt.id
		WHERE pt.club_id = ? AND pt.status IN ?
	`, clubID, []string{"pending", "ongoing"}).Scan(&pendingPhysicalTestRecords)
	insights["pendingPhysicalTestRecords"] = pendingPhysicalTestRecords

	// 4. 待支付订单数（复用 pendingOrders）
	insights["pendingOrders"] = pendingOrders

	overview["insights"] = insights

	return overview, nil
}

// GetRecentOrders 获取最近订单
func (s *ClubService) GetRecentOrders(clubID uint, limit int) ([]map[string]interface{}, error) {
	var orders []models.ClubOrder
	err := s.db.Where("club_id = ?", clubID).
		Order("created_at DESC").
		Limit(limit).
		Preload("Player").
		Preload("Analyst").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(orders))
	for _, o := range orders {
		result = append(result, map[string]interface{}{
			"id":          o.ID,
			"playerName":  o.Player.Name,
			"analystName": getAnalystName(o.Analyst),
			"status":      o.Status,
			"createdAt":   o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

// GetUpcomingTests 获取即将到来的体测
func (s *ClubService) GetUpcomingTests(clubID uint, limit int) ([]map[string]interface{}, error) {
	var tests []models.PhysicalTestActivity
	err := s.db.Where("club_id = ? AND status IN ?", clubID, []string{"pending", "ongoing"}).
		Order("start_date ASC").
		Limit(limit).
		Find(&tests).Error

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(tests))
	for _, t := range tests {
		playerCount := len(t.GetPlayerIDs())
		result = append(result, map[string]interface{}{
			"id":          t.ID,
			"name":        t.Name,
			"testDate":    t.StartDate.Format("2006-01-02"),
			"playerCount": playerCount,
			"status":      t.Status,
		})
	}
	return result, nil
}

// GetPlayers 获取球员列表
func (s *ClubService) GetPlayers(clubID uint, page, pageSize int, keyword, ageGroup, position, tag, status, sortBy, sortOrder string) ([]models.ClubPlayer, int64, error) {
	var players []models.ClubPlayer
	var total int64

	query := s.db.Model(&models.ClubPlayer{}).Where("club_id = ?", clubID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if ageGroup != "" {
		query = query.Where("age_group = ?", ageGroup)
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}

	// 关键词搜索（通过User关联）
	if keyword != "" {
		query = query.Joins("JOIN users ON users.id = club_players.user_id").
			Where("users.name LIKE ? OR users.nickname LIKE ? OR users.phone LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 标签筛选
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	// 统计总数
	query.Count(&total)

	// 排序字段映射（兼容前端驼峰命名）
	sortFieldMap := map[string]string{
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"joinDate":  "join_date",
		"ageGroup":  "age_group",
		"position":  "position",
		"status":    "status",
		"id":        "id",
		"name":      "name",
	}
	if mapped, ok := sortFieldMap[sortBy]; ok {
		sortBy = mapped
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	orderClause := clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: sortOrder == "desc"}

	// 分页（复制 query 避免 Count 影响后续查询）
	offset := (page - 1) * pageSize
	listQuery := query.Session(&gorm.Session{}).Order(orderClause).Offset(offset).Limit(pageSize).Preload("User")
	err := listQuery.Find(&players).Error

	return players, total, err
}

// GetPlayerByID 根据ID获取球员关联
func (s *ClubService) GetPlayerByID(playerID uint) (*models.ClubPlayer, error) {
	var player models.ClubPlayer
	err := s.db.Preload("User").First(&player, playerID).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// UpdatePlayerTags 更新球员标签
func (s *ClubService) UpdatePlayerTags(clubID, playerID uint, tags []string) error {
	tagsJSON, _ := json.Marshal(tags)
	return s.db.Model(&models.ClubPlayer{}).Where("id = ? AND club_id = ?", playerID, clubID).Update("tags", string(tagsJSON)).Error
}

// RemovePlayer 移除球员
func (s *ClubService) RemovePlayer(clubID, playerID uint) error {
	return s.db.Model(&models.ClubPlayer{}).Where("id = ? AND club_id = ?", playerID, clubID).Update("status", "left").Error
}

// GetPlayerAgeGroupStats 获取球员年龄分布统计
func (s *ClubService) GetPlayerAgeGroupStats(clubID uint) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 使用原始查询按年龄分组
	rows, err := s.db.Raw(`
		SELECT
			CASE
				WHEN u.age BETWEEN 6 AND 7 THEN 'U8'
				WHEN u.age BETWEEN 8 AND 9 THEN 'U10'
				WHEN u.age BETWEEN 10 AND 11 THEN 'U12'
				WHEN u.age BETWEEN 12 AND 13 THEN 'U14'
				WHEN u.age BETWEEN 14 AND 15 THEN 'U16'
				WHEN u.age BETWEEN 16 AND 17 THEN 'U18'
				ELSE '其他'
			END as age_group,
			COUNT(*) as count
		FROM club_players cp
		JOIN users u ON cp.user_id = u.id
		WHERE cp.club_id = ? AND cp.status = 'active'
		GROUP BY age_group
		ORDER BY age_group
	`, clubID).Rows()

	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var ageGroup string
		var count int
		rows.Scan(&ageGroup, &count)
		results = append(results, map[string]interface{}{
			"ageGroup": ageGroup,
			"count":    count,
		})
	}

	// 如果没有数据，返回默认分布
	if len(results) == 0 {
		results = []map[string]interface{}{
			{"ageGroup": "U8", "count": 0},
			{"ageGroup": "U10", "count": 0},
			{"ageGroup": "U12", "count": 0},
			{"ageGroup": "U14", "count": 0},
			{"ageGroup": "U16", "count": 0},
		}
	}

	return results, nil
}

// GetPlayerPositionStats 获取球员位置分布统计
func (s *ClubService) GetPlayerPositionStats(clubID uint) ([]map[string]interface{}, error) {
	var stats []map[string]interface{}

	rows, err := s.db.Raw(`
		SELECT COALESCE(cp.position, 'unknown') as position, COUNT(*) as count
		FROM club_players cp
		WHERE cp.club_id = ? AND cp.status = 'active'
		GROUP BY cp.position
	`, clubID).Rows()

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	positionNames := map[string]string{
		"forward":    "前锋",
		"midfielder": "中场",
		"defender":   "后卫",
		"goalkeeper": "守门员",
		"unknown":    "未知",
	}

	for rows.Next() {
		var position string
		var count int
		rows.Scan(&position, &count)
		name := positionNames[position]
		if name == "" {
			name = position
		}
		stats = append(stats, map[string]interface{}{
			"position": position,
			"name":     name,
			"count":    count,
		})
	}

	if len(stats) == 0 {
		stats = []map[string]interface{}{
			{"position": "forward", "name": "前锋", "count": 0},
			{"position": "midfielder", "name": "中场", "count": 0},
			{"position": "defender", "name": "后卫", "count": 0},
			{"position": "goalkeeper", "name": "守门员", "count": 0},
		}
	}

	return stats, nil
}

// GetAbilityRadar 获取梯队能力雷达图数据
func (s *ClubService) GetAbilityRadar(clubID uint) (map[string]interface{}, error) {
	// TODO: 基于体测数据计算真实能力值
	// 目前返回模拟数据
	return map[string]interface{}{
		"labels":      []string{"速度", "力量", "耐力", "灵敏", "柔韧", "技术"},
		"teamAvg":     []int{75, 68, 72, 70, 65, 78},
		"platformAvg": []int{70, 65, 68, 67, 63, 72},
	}, nil
}

// GetTopPerformers 获取TOP球员排行
func (s *ClubService) GetTopPerformers(clubID uint) ([]map[string]interface{}, error) {
	var performers []map[string]interface{}

	// TODO: 基于体测数据找出各维度TOP球员
	// 目前返回空列表
	rows, err := s.db.Raw(`
		SELECT cp.id, u.name, u.nickname, cp.age_group, COUNT(ptr.id) as test_count
		FROM club_players cp
		JOIN users u ON cp.user_id = u.id
		LEFT JOIN physical_test_records ptr ON ptr.player_id = cp.user_id AND ptr.club_id = cp.club_id
		WHERE cp.club_id = ? AND cp.status = 'active'
		GROUP BY cp.id, u.name, u.nickname, cp.age_group
		ORDER BY test_count DESC
		LIMIT 5
	`, clubID).Rows()

	if err != nil {
		return performers, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var name, nickname, ageGroup string
		var testCount int64
		rows.Scan(&id, &name, &nickname, &ageGroup, &testCount)

		displayName := name
		if displayName == "" {
			displayName = nickname
		}

		performers = append(performers, map[string]interface{}{
			"playerId":  id,
			"name":      displayName,
			"ageGroup":  ageGroup,
			"testCount": testCount,
			"metric":    "体测次数",
			"value":     testCount,
		})
	}

	return performers, nil
}

// 辅助函数：获取月初时间
func getStartOfMonth() string {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
}

// 辅助函数：获取分析师名称
func getAnalystName(analyst *models.Analyst) string {
	if analyst == nil {
		return "待分配"
	}
	if analyst.Name != "" {
		return analyst.Name
	}
	return analyst.User.Nickname
}

// ClubHomeData 俱乐部主页数据
type ClubHomeData struct {
	Hero         map[string]interface{}   `json:"hero"`
	About        map[string]interface{}   `json:"about"`
	Achievements []map[string]interface{} `json:"achievements"`
	Teams        []map[string]interface{} `json:"teams"`
	Coaches      []map[string]interface{} `json:"coaches"`
	Contact      map[string]interface{}   `json:"contact"`
}

// GetClubHome 获取俱乐部主页数据
func (s *ClubService) GetClubHome(clubID uint) (*ClubHomeData, error) {
	// 获取俱乐部信息
	var club models.Club
	if err := s.db.First(&club, clubID).Error; err != nil {
		return nil, err
	}

	// 构建主页数据
	home := &ClubHomeData{
		Hero: map[string]interface{}{
			"title":    club.Name,
			"subtitle": club.Description,
			"enabled":  true,
		},
		About: map[string]interface{}{
			"title":   "关于我们",
			"content": club.Description,
			"images":  []interface{}{},
			"enabled": true,
		},
		Achievements: []map[string]interface{}{},
		Teams:        []map[string]interface{}{},
		Coaches:      []map[string]interface{}{},
		Contact: map[string]interface{}{
			"address": club.Address,
			"phone":   club.ContactPhone,
			"enabled": true,
		},
	}

	return home, nil
}

// UpdateClubHome 更新俱乐部主页数据
func (s *ClubService) UpdateClubHome(clubID uint, data map[string]interface{}) error {
	// 目前支持更新俱乐部描述和基本信息
	updates := map[string]interface{}{}

	if hero, ok := data["hero"].(map[string]interface{}); ok {
		if title, ok := hero["title"].(string); ok && title != "" {
			updates["name"] = title
		}
		if subtitle, ok := hero["subtitle"].(string); ok {
			updates["description"] = subtitle
		}
	}

	if about, ok := data["about"].(map[string]interface{}); ok {
		if content, ok := about["content"].(string); ok {
			updates["description"] = content
		}
	}

	if contact, ok := data["contact"].(map[string]interface{}); ok {
		if address, ok := contact["address"].(string); ok {
			updates["address"] = address
		}
		if phone, ok := contact["phone"].(string); ok {
			updates["contact_phone"] = phone
		}
	}

	if len(updates) > 0 {
		return s.db.Model(&models.Club{}).Where("id = ?", clubID).Updates(updates).Error
	}

	return nil
}

// GetClubHomeTeams 获取俱乐部主页展示的球队列表
func (s *ClubService) GetClubHomeTeams(clubID uint) ([]map[string]interface{}, error) {
	var teams []map[string]interface{}

	// 获取球队列表
	rows, err := s.db.Table("teams").
		Select("teams.id, teams.name, teams.age_group, teams.description, teams.status, COUNT(team_players.id) as player_count").
		Joins("LEFT JOIN team_players ON teams.id = team_players.team_id AND team_players.status = 'active'").
		Where("teams.club_id = ? AND teams.status = 'active'", clubID).
		Group("teams.id").
		Order("teams.age_group ASC, teams.created_at DESC").
		Rows()

	if err != nil {
		return teams, err
	}

	for rows.Next() {
		var id uint
		var name, ageGroup, description, status string
		var playerCount int
		rows.Scan(&id, &name, &ageGroup, &description, &status, &playerCount)
		teams = append(teams, map[string]interface{}{
			"id":          id,
			"name":        name,
			"ageGroup":    ageGroup,
			"description": description,
			"playerCount": playerCount,
		})
	}
	rows.Close()

	return teams, nil
}

// ========== 俱乐部教练管理 ==========

// GetClubCoaches 获取俱乐部的教练列表
func (s *ClubService) GetClubCoaches(clubID uint, status string, keyword string, page, pageSize int) ([]models.ClubCoach, int64, error) {
	var coaches []models.ClubCoach
	var total int64

	query := s.db.Model(&models.ClubCoach{}).Where("club_id = ?", clubID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if keyword != "" {
		query = query.Joins("JOIN users ON users.id = club_coaches.user_id").
			Where("users.name LIKE ? OR users.nickname LIKE ? OR users.phone LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&coaches).Error
	return coaches, total, err
}

// GetClubCoachByID 根据ID获取俱乐部教练
func (s *ClubService) GetClubCoachByID(clubCoachID uint) (*models.ClubCoach, error) {
	var coach models.ClubCoach
	err := s.db.Preload("User").Preload("Club").First(&coach, clubCoachID).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}

// AddClubCoach 添加教练到俱乐部
func (s *ClubService) AddClubCoach(clubID, userID uint, primaryRole models.CoachRole, notes string) (*models.ClubCoach, error) {
	coach := &models.ClubCoach{
		ClubID:      clubID,
		UserID:      userID,
		PrimaryRole: primaryRole,
		Status:      models.ClubCoachStatusActive,
		Notes:       notes,
		JoinedAt:    time.Now(),
	}
	err := s.db.Create(coach).Error
	if err != nil {
		return nil, err
	}
	return coach, nil
}

// UpdateClubCoach 更新俱乐部教练信息
func (s *ClubService) UpdateClubCoach(clubCoachID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.ClubCoach{}).Where("id = ?", clubCoachID).Updates(updates).Error
}

// RemoveClubCoach 从俱乐部移除教练（软删除/标记离职）
func (s *ClubService) RemoveClubCoach(clubCoachID uint) error {
	return s.db.Model(&models.ClubCoach{}).Where("id = ?", clubCoachID).Updates(map[string]interface{}{
		"status":  models.ClubCoachStatusInactive,
		"left_at": time.Now(),
	}).Error
}

// GetClubCoachTeams 获取教练在俱乐部下的球队分配
func (s *ClubService) GetClubCoachTeams(clubCoachID uint) ([]models.TeamCoach, error) {
	var clubCoach models.ClubCoach
	if err := s.db.First(&clubCoach, clubCoachID).Error; err != nil {
		return nil, err
	}

	var teamCoaches []models.TeamCoach
	err := s.db.Preload("Team").Where("user_id = ? AND status = ?", clubCoach.UserID, "active").Find(&teamCoaches).Error
	return teamCoaches, err
}

// AssignCoachToTeam 将俱乐部教练分配到球队
func (s *ClubService) AssignCoachToTeam(userID, teamID uint, role models.CoachRole) (*models.TeamCoach, error) {
	// 检查球队是否存在
	var team models.Team
	if err := s.db.First(&team, teamID).Error; err != nil {
		return nil, err
	}

	// 检查是否已存在相同角色的分配
	var existing models.TeamCoach
	err := s.db.Where("user_id = ? AND team_id = ? AND role = ? AND status = ?", userID, teamID, role, "active").First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("该教练已在该球队担任此角色")
	}

	tc := &models.TeamCoach{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		Status:   "active",
		JoinedAt: time.Now(),
	}
	err = s.db.Create(tc).Error
	if err != nil {
		return nil, err
	}
	return tc, nil
}

// RemoveCoachFromTeam 从球队移除教练
func (s *ClubService) RemoveCoachFromTeam(teamCoachID uint) error {
	return s.db.Model(&models.TeamCoach{}).Where("id = ?", teamCoachID).Updates(map[string]interface{}{
		"status":  "inactive",
		"left_at": time.Now(),
	}).Error
}

// IsCoachOfClub 检查用户是否属于该俱乐部
func (s *ClubService) IsCoachOfClub(userID, clubID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.ClubCoach{}).Where("user_id = ? AND club_id = ? AND status = ?", userID, clubID, models.ClubCoachStatusActive).Count(&count).Error
	return count > 0, err
}

// GetClubCoachByUserID 根据用户ID获取俱乐部教练关系
func (s *ClubService) GetClubCoachByUserID(userID, clubID uint) (*models.ClubCoach, error) {
	var coach models.ClubCoach
	err := s.db.Where("user_id = ? AND club_id = ?", userID, clubID).First(&coach).Error
	if err != nil {
		return nil, err
	}
	return &coach, nil
}

// GetClubHomeCoaches 获取俱乐部主页展示的教练列表
func (s *ClubService) GetClubHomeCoaches(clubID uint) ([]map[string]interface{}, error) {
	var coaches []map[string]interface{}

	// 获取球队列表
	var teamIDs []uint
	s.db.Table("teams").Where("club_id = ? AND status = ?", clubID, "active").Pluck("id", &teamIDs)

	if len(teamIDs) == 0 {
		return coaches, nil
	}

	// 获取主教练作为俱乐部主页展示教练，旧版 is_admin 字段已废弃，统一用 role=head_coach 表示球队负责人
	type clubHomeCoachRow struct {
		ID       uint
		Name     string
		Nickname string
		Avatar   string
		Role     models.CoachRole
	}
	var coachUsers []clubHomeCoachRow
	s.db.Table("team_coaches").
		Select("users.id, users.name, users.nickname, users.avatar, team_coaches.role").
		Joins("JOIN users ON team_coaches.user_id = users.id").
		Where("team_coaches.team_id IN ? AND team_coaches.status = ? AND team_coaches.role = ?", teamIDs, "active", models.CoachRoleHead).
		Group("users.id, users.name, users.nickname, users.avatar, team_coaches.role").
		Limit(10).
		Scan(&coachUsers)

	for _, u := range coachUsers {
		coaches = append(coaches, map[string]interface{}{
			"id":         u.ID,
			"name":       u.Name,
			"nickname":   u.Nickname,
			"avatar":     u.Avatar,
			"role":       u.Role,
			"roleLabel":  models.GetCoachRoleLabel(u.Role),
			"isAdmin":    u.Role == models.CoachRoleHead,
			"license":    "",
			"experience": "",
			"status":     "active",
		})
	}

	return coaches, nil
}
