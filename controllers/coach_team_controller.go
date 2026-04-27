package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// CoachTeamController 教练视角的球队控制器
type CoachTeamController struct {
	teamRepo         *repositories.TeamRepository
	weeklyReportRepo *repositories.WeeklyReportRepository
	db               *gorm.DB
}

// NewCoachTeamController 创建教练球队控制器
func NewCoachTeamController(teamRepo *repositories.TeamRepository, weeklyReportRepo *repositories.WeeklyReportRepository, db *gorm.DB) *CoachTeamController {
	return &CoachTeamController{
		teamRepo:         teamRepo,
		weeklyReportRepo: weeklyReportRepo,
		db:               db,
	}
}

// checkTeamAccess 校验教练是否有权限操作球队
func (c *CoachTeamController) checkTeamAccess(ctx *gin.Context, teamID uint) (uint, bool) {
	userID := ctx.GetUint("userId")
	hasAuth, err := c.teamRepo.IsCoachOfTeam(userID, teamID)
	if err != nil || !hasAuth {
		utils.Error(ctx, 403, "无权操作该球队")
		return 0, false
	}
	return userID, true
}

// GetMyTeams 获取教练关联的球队列表
func (c *CoachTeamController) GetMyTeams(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	// 获取用户关联的球队（通过 TeamCoach 表）
	teams, err := c.teamRepo.GetCoachTeams(userID)
	if err != nil {
		utils.ServerError(ctx, "获取球队列表失败")
		return
	}

	result := make([]gin.H, 0, len(teams))
	seen := make(map[uint]bool)
	for _, team := range teams {
		if seen[team.ID] {
			continue
		}
		seen[team.ID] = true
		playerCount, _ := c.teamRepo.CountPlayers(team.ID)
		coachCount, _ := c.teamRepo.CountCoaches(team.ID)
		result = append(result, gin.H{
			"id":           team.ID,
			"name":         team.Name,
			"ageGroup":     team.AgeGroup,
			"description":  team.Description,
			"playerCount":  playerCount,
			"coachCount":   coachCount,
		})
	}

	utils.SuccessResponse(ctx, result)
}

// GetTeamDetail 获取球队详情
func (c *CoachTeamController) GetTeamDetail(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	userID := ctx.GetUint("userId")

	// 验证教练是否关联了该球队
	hasAuth, _ := c.teamRepo.IsCoachOfTeam(userID, uint(teamID))
	if !hasAuth {
		utils.Error(ctx, 403, "无权查看该球队")
		return
	}

	team, err := c.teamRepo.FindByID(uint(teamID))
	if err != nil {
		utils.Error(ctx, 404, "球队不存在")
		return
	}

	playerCount, _ := c.teamRepo.CountPlayers(team.ID)
	coachCount, _ := c.teamRepo.CountCoaches(team.ID)

	utils.SuccessResponse(ctx, gin.H{
		"id":           team.ID,
		"name":         team.Name,
		"ageGroup":     team.AgeGroup,
		"description":  team.Description,
		"playerCount": playerCount,
		"coachCount":  coachCount,
	})
}

// GetTeamPlayers 获取球队球员列表
func (c *CoachTeamController) GetTeamPlayers(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	userID := ctx.GetUint("userId")

	// 验证教练是否关联了该球队
	hasAuth, _ := c.teamRepo.IsCoachOfTeam(userID, uint(teamID))
	if !hasAuth {
		utils.Error(ctx, 403, "无权查看该球队球员")
		return
	}

	players, _, err := c.teamRepo.GetPlayers(uint(teamID), "", "", "")
	if err != nil {
		utils.ServerError(ctx, "获取球员列表失败")
		return
	}

	result := make([]gin.H, 0, len(players))
	for _, p := range players {
		user := p.User
		if user == nil {
			continue
		}
		// isRegistered: 当userId > 0 且 user存在且状态为active时为true
		isRegistered := p.UserID > 0 && user != nil && user.Status == "active"

		// 查询报告数量和平均评分
		var reportCount int64
		c.db.Model(&models.Report{}).Where("user_id = ? AND status = ?", p.UserID, "completed").Count(&reportCount)

		var avgScore float64
		c.db.Model(&models.Report{}).Where("user_id = ? AND status = ?", p.UserID, "completed").Select("COALESCE(AVG(overall_rating), 0)").Scan(&avgScore)

		result = append(result, gin.H{
			"id":           p.ID,
			"userId":       p.UserID,
			"name":         user.Name,
			"nickname":     user.Nickname,
			"avatar":       user.Avatar,
			"age":          user.Age,
			"birthDate":    user.BirthDate,
			"gender":       user.Gender,
			"phone":        user.Phone,
			"position":     p.Position,
			"jerseyNumber": p.JerseyNumber,
			"status":       p.Status,
			"joinDate":     utils.FormatTime(&p.JoinedAt),
			"isRegistered": isRegistered,
			"reportCount":  reportCount,
			"avgScore":     avgScore,
		})
	}

	utils.SuccessResponse(ctx, result)
}

// GetTeamCoaches 获取球队教练列表
func (c *CoachTeamController) GetTeamCoaches(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	userID := ctx.GetUint("userId")

	// 验证教练是否关联了该球队（或者俱乐部管理员）
	hasAuth, _ := c.teamRepo.IsCoachOfTeam(userID, uint(teamID))
	if !hasAuth {
		utils.Error(ctx, 403, "无权查看")
		return
	}

	coaches, err := c.teamRepo.GetCoaches(uint(teamID), "")
	if err != nil {
		utils.ServerError(ctx, "获取教练列表失败")
		return
	}

	result := make([]gin.H, 0, len(coaches))
	for _, coach := range coaches {
		user := coach.User
		if user == nil {
			continue
		}
		result = append(result, gin.H{
			"id":        coach.ID,
			"userId":    coach.UserID,
			"name":      user.Name,
			"avatar":    user.Avatar,
			"role":      coach.Role,
			"roleLabel": models.GetCoachRoleLabel(coach.Role),
			"status":    coach.Status,
			"joinedAt":  utils.FormatTime(&coach.JoinedAt),
		})
	}

	utils.SuccessResponse(ctx, result)
}

// AddPlayer 添加球员到球队
func (c *CoachTeamController) AddPlayer(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	// 校验权限
	if _, ok := c.checkTeamAccess(ctx, uint(teamID)); !ok {
		return
	}

	var req struct {
		Phone        string `json:"phone" binding:"required"`
		Name         string `json:"name"`
		Position     string `json:"position"`
		JerseyNumber string `json:"jerseyNumber"`
		Age         int    `json:"age"`
		BirthDate    string `json:"birthDate"`
		Gender       string `json:"gender"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 查找或创建用户
	var user *models.User
	user, err = c.teamRepo.FindUserByPhone(req.Phone)
	if err != nil {
		// 用户不存在，创建新用户
		user = &models.User{
			Phone:    req.Phone,
			Name:     req.Name,
			Nickname: req.Name,
			Role:     models.RoleUser,
			Status:   models.StatusActive,
		}
		if err := c.teamRepo.CreateUser(user); err != nil {
			utils.ServerError(ctx, "创建用户失败")
			return
		}
	}

	// 添加球员到球队
	if err := c.teamRepo.AddPlayer(uint(teamID), user.ID, req.JerseyNumber, req.Position); err != nil {
		utils.ServerError(ctx, "添加球员失败")
		return
	}

	// 同步创建 ClubPlayer（如果不存在）
	team, err := c.teamRepo.FindByID(uint(teamID))
	teamAgeGroup := ""
	clubID := uint(0)
	if err == nil && team != nil {
		teamAgeGroup = team.AgeGroup
		clubID = team.ClubID
	}
	var clubPlayerCount int64
	c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND user_id = ?", clubID, user.ID).Count(&clubPlayerCount)
	if clubPlayerCount == 0 && clubID > 0 {
		cp := &models.ClubPlayer{
			ClubID:   clubID,
			UserID:   user.ID,
			JoinDate: time.Now(),
			AgeGroup: teamAgeGroup,
			Status:   "active",
		}
		_ = c.db.Create(cp).Error
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":      user.ID,
		"name":    user.Name,
		"phone":   user.Phone,
		"message": "添加成功",
	})
}

// UpdatePlayer 更新球员信息
func (c *CoachTeamController) UpdatePlayer(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	playerIDStr := ctx.Param("playerId")
	playerID, err := strconv.ParseUint(playerIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	// 校验权限
	if _, ok := c.checkTeamAccess(ctx, uint(teamID)); !ok {
		return
	}

	var req struct {
		Name         string `json:"name"`
		Position     string `json:"position"`
		JerseyNumber string `json:"jerseyNumber"`
		Status       string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 更新 TeamPlayer 字段（球队层面属性）
	updates := make(map[string]interface{})
	if req.Position != "" {
		updates["position"] = req.Position
	}
	if req.JerseyNumber != "" {
		updates["jersey_number"] = req.JerseyNumber
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := c.teamRepo.UpdatePlayer(uint(teamID), uint(playerID), updates); err != nil {
			utils.ServerError(ctx, "更新失败")
			return
		}
	}

	// 更新用户姓名（users 表，不属于 TeamPlayer）
	if req.Name != "" {
		player, err := c.teamRepo.GetTeamPlayer(uint(playerID))
		if err == nil && player != nil {
			c.db.Model(&models.User{}).Where("id = ?", player.UserID).Update("name", req.Name)
		}
	}

	utils.SuccessResponse(ctx, gin.H{"message": "更新成功"})
}

// RemovePlayer 从球队移除球员
func (c *CoachTeamController) RemovePlayer(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	playerIDStr := ctx.Param("playerId")
	playerID, err := strconv.ParseUint(playerIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	// 校验权限
	if _, ok := c.checkTeamAccess(ctx, uint(teamID)); !ok {
		return
	}

	// 获取球员的 userID
	player, err := c.teamRepo.GetTeamPlayer(uint(playerID))
	if err != nil {
		utils.Error(ctx, 404, "球员不存在")
		return
	}

	if err := c.teamRepo.RemovePlayer(uint(teamID), player.UserID, "transferred"); err != nil {
		utils.ServerError(ctx, "移除球员失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{"message": "移除成功"})
}

// UpdateTeam 更新球队信息
func (c *CoachTeamController) UpdateTeam(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	// 校验权限
	if _, ok := c.checkTeamAccess(ctx, uint(teamID)); !ok {
		return
	}

	var req struct {
		Name        string `json:"name"`
		AgeGroup    string `json:"ageGroup"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AgeGroup != "" {
		updates["age_group"] = req.AgeGroup
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) > 0 {
		if err := c.teamRepo.UpdateTeam(uint(teamID), updates); err != nil {
			utils.ServerError(ctx, "更新失败")
			return
		}
	}

	utils.SuccessResponse(ctx, gin.H{"message": "更新成功"})
}

// CreateWeeklyReport 教练为球员创建周报（批量）
func (c *CoachTeamController) CreateWeeklyReport(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	// 校验权限
	coachID, ok := c.checkTeamAccess(ctx, uint(teamID))
	if !ok {
		return
	}

	var req models.CreateWeeklyReportInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误: "+err.Error())
		return
	}

	// 解析周起始日期
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		utils.ValidationError(ctx, "无效的周起始日期，格式应为 2006-01-02")
		return
	}

	// 计算周结束日期
	weekEnd := weekStart.AddDate(0, 0, 6)
	if req.WeekEnd != "" {
		weekEnd, _ = time.Parse("2006-01-02", req.WeekEnd)
	}

	// 验证球员是否属于该球队
	teamPlayers, _, err := c.teamRepo.GetPlayers(uint(teamID), "", "", "")
	if err != nil {
		utils.ServerError(ctx, "获取球员列表失败")
		return
	}

	playerMap := make(map[uint]bool)
	for _, p := range teamPlayers {
		playerMap[p.ID] = true
	}

	// 创建周报
	created := 0
	failed := 0
	for _, playerID := range req.PlayerIDs {
		// 检查球员是否属于该球队
		if !playerMap[playerID] {
			failed++
			continue
		}

		// 检查是否已存在周报
		weekStartStr := weekStart.Format("2006-01-02")
		existing, _ := c.weeklyReportRepo.GetByPlayerAndWeek(playerID, weekStartStr)
		if existing != nil {
			failed++
			continue
		}

		report := &models.WeeklyReport{
			TeamID:       uint(teamID),
			PlayerID:     playerID,
			CoachID:      coachID,
			WeekStart:    weekStart,
			WeekEnd:      weekEnd,
			ReviewStatus: "pending",
		}

		if err := c.weeklyReportRepo.Create(report); err != nil {
			failed++
			continue
		}
		created++
	}

	// 异步通知球员
	if created > 0 {
		go func() {
			notificationHelper := NewNotificationHelper(c.db)
			var user models.User
			coachName := "教练"
			if err := c.db.First(&user, coachID).Error; err == nil {
				coachName = user.Name
			}
			var team models.Team
			teamName := "球队"
			if err := c.db.First(&team, teamID).Error; err == nil {
				teamName = team.Name
			}
			weekLabel := weekStart.Format("01/02") + " ~ " + weekEnd.Format("01/02")
			for _, playerID := range req.PlayerIDs {
				if !playerMap[playerID] {
					continue
				}
				var report models.WeeklyReport
				if err := c.db.Where("team_id = ? AND player_id = ? AND week_start = ?", teamID, playerID, weekStart.Format("2006-01-02")).First(&report).Error; err == nil {
					notificationHelper.NotifyWeeklyReportCreated(playerID, coachName, teamName, weekLabel, report.ID)
				}
			}
		}()
	}

	utils.SuccessResponse(ctx, gin.H{
		"created": created,
		"failed":  failed,
		"message": "成功创建 " + strconv.Itoa(created) + " 份周报",
	})
}
