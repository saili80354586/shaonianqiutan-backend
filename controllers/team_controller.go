package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// TeamController 统一球队管理控制器
// 同时支持俱乐部管理员和教练访问
type TeamController struct {
	teamRepo         *repositories.TeamRepository
	weeklyReportRepo *repositories.WeeklyReportRepository
	matchSummaryRepo *repositories.MatchSummaryRepository
	activityRepo     *repositories.ActivityRepository
	clubRepo         *repositories.ClubRepository
	physicalTestRepo *repositories.PhysicalTestRepository
	db               *gorm.DB
}

// NewTeamController 创建统一球队控制器
func NewTeamController(
	teamRepo *repositories.TeamRepository,
	weeklyReportRepo *repositories.WeeklyReportRepository,
	matchSummaryRepo *repositories.MatchSummaryRepository,
	activityRepo *repositories.ActivityRepository,
	clubRepo *repositories.ClubRepository,
	physicalTestRepo *repositories.PhysicalTestRepository,
	db *gorm.DB,
) *TeamController {
	return &TeamController{
		teamRepo:         teamRepo,
		weeklyReportRepo: weeklyReportRepo,
		matchSummaryRepo: matchSummaryRepo,
		activityRepo:     activityRepo,
		clubRepo:         clubRepo,
		physicalTestRepo: physicalTestRepo,
		db:               db,
	}
}

// getAccessContext 获取访问上下文
func (c *TeamController) getAccessContext(ctx *gin.Context) *middleware.TeamAccessContext {
	return middleware.GetTeamAccessContext(ctx)
}

// getUserID 获取当前用户ID
func (c *TeamController) getUserID(ctx *gin.Context) uint {
	return middleware.GetUserID(ctx)
}

// ==================== 球队基本信息 ====================

// GetTeamDetail 获取球队详情
func (c *TeamController) GetTeamDetail(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	team, err := c.teamRepo.FindByID(uint(teamID))
	if err != nil {
		utils.NotFoundError(ctx, "球队不存在")
		return
	}

	playerCount, _ := c.teamRepo.CountPlayers(team.ID)
	coachCount, _ := c.teamRepo.CountCoaches(team.ID)

	utils.SuccessResponse(ctx, gin.H{
		"id":          team.ID,
		"name":        team.Name,
		"ageGroup":    team.AgeGroup,
		"description": team.Description,
		"clubId":      team.ClubID,
		"playerCount": playerCount,
		"coachCount":  coachCount,
		"createdAt":   utils.FormatDateTime(team.CreatedAt),
	})
}

// UpdateTeam 更新球队信息
func (c *TeamController) UpdateTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		AgeGroup    string `json:"ageGroup"`
		Description string `json:"description"`
		Status      string `json:"status"`
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
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := c.teamRepo.UpdateTeam(uint(teamID), updates); err != nil {
			utils.ServerError(ctx, "更新失败")
			return
		}
	}

	// 如果恢复为 active，同时清除软删除标记
	if req.Status == "active" {
		c.teamRepo.Restore(uint(teamID))
	}

	utils.SuccessResponse(ctx, gin.H{"message": "更新成功"})
}

// ==================== 球员管理 ====================

// GetTeamPlayers 获取球队球员列表
func (c *TeamController) GetTeamPlayers(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
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

// AddPlayer 添加球员到球队
func (c *TeamController) AddPlayer(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		Phone        string `json:"phone" binding:"required"`
		Name         string `json:"name"`
		Position     string `json:"position"`
		JerseyNumber string `json:"jerseyNumber"`
		Age          int    `json:"age"`
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
func (c *TeamController) UpdatePlayer(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
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
func (c *TeamController) RemovePlayer(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	player, err := c.teamRepo.GetTeamPlayer(uint(playerID))
	if err != nil {
		utils.NotFoundError(ctx, "球员不存在")
		return
	}

	if err := c.teamRepo.RemovePlayer(uint(teamID), player.UserID, "transferred"); err != nil {
		utils.ServerError(ctx, "移除球员失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{"message": "移除成功"})
}

// ==================== 教练管理 ====================

// GetTeamCoaches 获取球队教练列表
func (c *TeamController) GetTeamCoaches(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
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

// ==================== 周报管理 ====================

// CreateWeeklyReport 创建周报（主教练/俱乐部管理员）
func (c *TeamController) CreateWeeklyReport(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	coachID := c.getUserID(ctx)

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
		playerMap[p.UserID] = true
	}

	// 创建周报
	created := 0
	failed := 0
	for _, playerID := range req.PlayerIDs {
		if !playerMap[playerID] {
			failed++
			continue
		}

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

	utils.SuccessResponse(ctx, gin.H{
		"created": created,
		"failed":  failed,
		"message": "成功创建 " + strconv.Itoa(created) + " 份周报",
	})
}

// GetWeeklyReports 获取球队周报列表
func (c *TeamController) GetWeeklyReports(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := ctx.Query("status")

	// 获取球队周报列表
	reports, total, err := c.weeklyReportRepo.ListByTeam(uint(teamID), status, page, pageSize)
	if err != nil {
		utils.ServerError(ctx, "获取周报列表失败")
		return
	}

	// 转换为响应格式
	result := make([]gin.H, 0, len(reports))
	for _, r := range reports {
		player := r.Player
		playerName := ""
		if player != nil {
			playerName = player.Name
		}
		result = append(result, gin.H{
			"id":            r.ID,
			"playerId":      r.PlayerID,
			"playerName":    playerName,
			"teamId":        r.TeamID,
			"coachId":       r.CoachID,
			"weekStart":      r.WeekStart,
			"weekEnd":        r.WeekEnd,
			"knowledgeSummary":    r.KnowledgeSummary,
			"tacticalContent":     r.TacticalContent,
			"physicalCondition":   r.PhysicalCondition,
			"matchPerformance":     r.MatchPerformance,
			"selfAttitudeRating":  r.SelfAttitudeRating,
			"selfTechniqueRating": r.SelfTechniqueRating,
			"selfTeamworkRating":  r.SelfTeamworkRating,
			"improvements":        r.ImprovementsDetail,
			"reviewStatus":        r.ReviewStatus,
			"reviewComment":       r.ReviewComment,
			"coachAttitudeRating": r.CoachAttitudeRating,
			"createdAt":        utils.FormatTime(&r.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":  result,
		"total": total,
		"page":  page,
		"pageSize": pageSize,
	})
}

// ==================== 比赛总结管理 ====================

// CreateMatchSummary 创建比赛总结
func (c *TeamController) CreateMatchSummary(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	coachID := ctx.GetUint("userId")

	var req struct {
		MatchName     string `json:"matchName" binding:"required"`
		MatchDate     string `json:"matchDate" binding:"required"`
		Opponent      string `json:"opponent" binding:"required"`
		Location      string `json:"location"`      // home/away/neutral
		MatchFormat   string `json:"matchFormat"`   // 5人制/8人制/11人制
		OurScore      int    `json:"ourScore"`
		OpponentScore int    `json:"opponentScore"`
		PlayerIDs     []uint `json:"playerIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 获取球队当前活跃球员
	teamPlayers, _, err := c.teamRepo.GetPlayers(uint(teamID), "active", "", "")
	if err != nil {
		utils.ServerError(ctx, "获取球员列表失败")
		return
	}

	// 构建有效的球员ID映射（使用 UserID 作为球员标识，与 TeamPlayer 和周报逻辑保持一致）
	playerMap := make(map[uint]bool)
	validPlayerIDs := make([]uint, 0, len(teamPlayers))
	for _, p := range teamPlayers {
		if p.UserID > 0 {
			playerMap[p.UserID] = true
			validPlayerIDs = append(validPlayerIDs, p.UserID)
		}
	}

	// 确定最终关联的球员列表
	finalPlayerIDs := make([]uint, 0)
	if len(req.PlayerIDs) > 0 {
		// 校验传入的球员是否都属于该球队
		for _, pid := range req.PlayerIDs {
			if playerMap[pid] {
				finalPlayerIDs = append(finalPlayerIDs, pid)
			}
		}
	} else {
		// 默认关联所有活跃球员
		finalPlayerIDs = validPlayerIDs
	}

	if len(finalPlayerIDs) == 0 {
		utils.ValidationError(ctx, "该球队暂无在队球员，请先添加球员或选择参赛球员")
		return
	}

	// 自动计算比赛结果
	matchResult := "draw"
	if req.OurScore > req.OpponentScore {
		matchResult = "win"
	} else if req.OurScore < req.OpponentScore {
		matchResult = "lose"
	}

	playerIDsJSON, _ := json.Marshal(finalPlayerIDs)

	summary := &models.MatchSummary{
		TeamID:      uint(teamID),
		CoachID:     coachID,
		MatchName:   req.MatchName,
		MatchDate:   req.MatchDate,
		Opponent:    req.Opponent,
		Location:    req.Location,
		MatchFormat: req.MatchFormat,
		OurScore:    req.OurScore,
		OppScore:    req.OpponentScore,
		Result:      matchResult,
		Status:      "pending",
		PlayerIDs:   string(playerIDsJSON),
		PlayerCount: len(finalPlayerIDs),
	}

	if err := c.matchSummaryRepo.Create(summary); err != nil {
		utils.ServerError(ctx, "创建比赛总结失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":      summary.ID,
		"message": "创建成功",
	})
}

// GetMatchSummaries 获取比赛总结列表
func (c *TeamController) GetMatchSummaries(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := ctx.Query("status")

	summaries, total, err := c.matchSummaryRepo.ListByTeam(uint(teamID), status, page, pageSize)
	if err != nil {
		utils.ServerError(ctx, "获取比赛总结列表失败")
		return
	}

	// 转换为响应格式
	result := make([]models.MatchSummaryListResponse, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, s.ToListResponse())
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ==================== 缺失的方法存根 ====================

// GetTeams 获取俱乐部球队列表
func (c *TeamController) GetTeams(ctx *gin.Context) {
	clubID, err := strconv.ParseUint(ctx.Param("clubId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	includeDeleted := ctx.Query("includeDeleted") == "true"
	status := ctx.DefaultQuery("status", "active")

	// 查询数据库获取球队列表
	teams, err := c.teamRepo.FindByClubID(uint(clubID), status, includeDeleted)
	if err != nil {
		utils.ServerError(ctx, "获取球队列表失败")
		return
	}

	// 为每个球队添加 playerCount 和 coachCount
	type TeamWithCount struct {
		models.Team
		PlayerCount int `json:"playerCount"`
		CoachCount  int `json:"coachCount"`
	}

	teamsWithCount := make([]TeamWithCount, len(teams))
	for i, t := range teams {
		playerCount, _ := c.teamRepo.CountPlayers(t.ID)
		coachCount, _ := c.teamRepo.CountCoaches(t.ID)
		teamsWithCount[i] = TeamWithCount{
			Team:        t,
			PlayerCount: int(playerCount),
			CoachCount:  int(coachCount),
		}
	}

	utils.SuccessResponse(ctx, teamsWithCount)
}

// CreateTeam 创建球队
func (c *TeamController) CreateTeam(ctx *gin.Context) {
	clubID, err := strconv.ParseUint(ctx.Param("clubId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req struct {
		Name           string `json:"name" binding:"required"`
		AgeGroup       string `json:"ageGroup" binding:"required"`
		Description    string `json:"description"`
		BirthYearStart int    `json:"birthYearStart"`
		BirthYearEnd   int    `json:"birthYearEnd"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	team := &models.Team{
		Name:        req.Name,
		ClubID:      uint(clubID),
		AgeGroup:    req.AgeGroup,
		Description: req.Description,
		Status:      "active",
	}
	if req.BirthYearStart > 0 {
		team.BirthYearStart = &req.BirthYearStart
	}
	if req.BirthYearEnd > 0 {
		team.BirthYearEnd = &req.BirthYearEnd
	}

	if err := c.teamRepo.Create(team); err != nil {
		utils.ServerError(ctx, "创建球队失败")
		return
	}

	utils.SuccessResponse(ctx, team)
}

// GetTeam 获取球队详情（别名）
func (c *TeamController) GetTeam(ctx *gin.Context) {
	c.GetTeamDetail(ctx)
}

// DeleteTeam 删除球队（软删除）
func (c *TeamController) DeleteTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	userID := c.getUserID(ctx)
	var team models.Team
	c.db.First(&team, uint(teamID))

	if err := c.teamRepo.Delete(uint(teamID)); err != nil {
		utils.ServerError(ctx, "删除球队失败")
		return
	}

	// 记录操作日志
	var adminName string
	var user models.User
	if err := c.db.First(&user, userID).Error; err == nil {
		adminName = user.Name
	}
	c.db.Create(&models.AdminOperationLog{
		ClubID:    team.ClubID,
		AdminID:   userID,
		AdminName: adminName,
		Action:    "delete_team",
		Target:    "team",
		TargetID:  uint(teamID),
		Detail:    "删除球队: " + team.Name,
		IP:        ctx.ClientIP(),
	})

	utils.SuccessResponse(ctx, gin.H{"message": "删除成功"})
}

// RestoreTeam 恢复已归档球队
func (c *TeamController) RestoreTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	userID := c.getUserID(ctx)
	var team models.Team
	c.db.Unscoped().First(&team, uint(teamID))

	if err := c.teamRepo.Restore(uint(teamID)); err != nil {
		utils.ServerError(ctx, "恢复球队失败")
		return
	}

	// 记录操作日志
	var adminName string
	var user models.User
	if err := c.db.First(&user, userID).Error; err == nil {
		adminName = user.Name
	}
	c.db.Create(&models.AdminOperationLog{
		ClubID:    team.ClubID,
		AdminID:   userID,
		AdminName: adminName,
		Action:    "restore_team",
		Target:    "team",
		TargetID:  uint(teamID),
		Detail:    "恢复球队: " + team.Name,
		IP:        ctx.ClientIP(),
	})

	utils.SuccessResponse(ctx, gin.H{"message": "恢复成功"})
}

// AddCoach 添加教练
func (c *TeamController) AddCoach(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		CoachID uint   `json:"coachId" binding:"required"`
		Role    string `json:"role"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	role := models.CoachRoleHead
	if req.Role != "" {
		role = models.CoachRole(req.Role)
	}

	if err := c.teamRepo.AddCoach(uint(teamID), req.CoachID, role); err != nil {
		utils.ServerError(ctx, "添加教练失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{"message": "添加教练成功"})
}

// UpdateCoach 更新教练
func (c *TeamController) UpdateCoach(ctx *gin.Context) {
	utils.SuccessResponse(ctx, gin.H{"message": "更新教练成功"})
}

// RemoveCoach 移除教练
func (c *TeamController) RemoveCoach(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	coachID, err := strconv.ParseUint(ctx.Param("coachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的教练ID")
		return
	}

	if err := c.teamRepo.RemoveCoach(uint(teamID), uint(coachID)); err != nil {
		utils.ServerError(ctx, "移除教练失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{"message": "移除教练成功"})
}

// CreateInvitation 创建邀请
func (c *TeamController) CreateInvitation(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	creatorID := ctx.GetUint("userId")

	var req struct {
		Type        string `json:"type" binding:"required"`
		TargetUserID *uint  `json:"targetUserId"`
		TargetPhone  string `json:"targetPhone"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 获取球队信息以填充 ClubID
	team, err := c.teamRepo.FindByID(uint(teamID))
	if err != nil {
		utils.ServerError(ctx, "获取球队信息失败")
		return
	}

	// 生成邀请码
	inviteCode := fmt.Sprintf("INV%s%d%d", time.Now().Format("20060102150405"), teamID, creatorID%10000)

	inv := &models.TeamInvitation{
		TeamID:       uint(teamID),
		ClubID:       team.ClubID,
		Type:         models.InvitationType(req.Type),
		InviteCode:   inviteCode,
		TargetUserID: req.TargetUserID,
		TargetPhone:  req.TargetPhone,
		Status:       models.InvitationStatusPending,
		CreatedBy:    creatorID,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7天过期
	}

	if err := c.teamRepo.CreateInvitation(inv); err != nil {
		utils.ServerError(ctx, "创建邀请失败")
		return
	}

	// 构建完整球队名称（俱乐部 + 球队）
	clubName := ""
	if team.Club != nil {
		clubName = team.Club.Name
	}
	fullTeamName := team.Name
	if clubName != "" {
		fullTeamName = clubName + " " + team.Name
	}

	// 发送通知给被邀请人
	if req.TargetUserID != nil && *req.TargetUserID > 0 {
		notification := models.Notification{
			UserID:   *req.TargetUserID,
			Type:     models.NotificationTypeInvitation,
			Title:    "收到球队邀请",
			Content:  fmt.Sprintf("%s 邀请您加入球队", fullTeamName),
			Data:     fmt.Sprintf(`{"invite_code":"%s","team_id":%d,"team_name":"%s","club_name":"%s","status":"pending"}`, inviteCode, teamID, fullTeamName, clubName),
			Priority: 2,
		}
		c.db.Create(&notification)
	}

	utils.SuccessResponse(ctx, gin.H{
		"code":  inviteCode,
		"id":    inv.ID,
		"message": "邀请创建成功",
	})
}

// GetInvitations 获取球队邀请列表
func (c *TeamController) GetInvitations(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	status := ctx.Query("status")
	invitations, err := c.teamRepo.GetInvitations(uint(teamID), status)
	if err != nil {
		utils.ServerError(ctx, "获取邀请列表失败")
		return
	}

	// 转换为响应格式
	result := make([]gin.H, 0, len(invitations))
	for _, inv := range invitations {
		item := gin.H{
			"id":         inv.ID,
			"teamId":     inv.TeamID,
			"type":       inv.Type,
			"inviteCode": inv.InviteCode,
			"status":     inv.Status,
			"createdAt":  utils.FormatTime(&inv.CreatedAt),
			"expiresAt":  utils.FormatTime(&inv.ExpiresAt),
		}
		if inv.TargetUser != nil {
			item["targetUser"] = gin.H{
				"id":       inv.TargetUser.ID,
				"name":     inv.TargetUser.Name,
				"phone":    inv.TargetUser.Phone,
				"nickname": inv.TargetUser.Nickname,
			}
		}
		if inv.Creator != nil {
			item["creator"] = gin.H{
				"id":   inv.Creator.ID,
				"name": inv.Creator.Name,
			}
		}
		result = append(result, item)
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}
func (c *TeamController) GetInvitation(ctx *gin.Context) {
	code := ctx.Param("code")

	inv, err := c.teamRepo.FindInvitationByCode(code)
	if err != nil {
		utils.NotFoundError(ctx, "邀请不存在或已过期")
		return
	}

	// 检查是否过期
	if inv.IsExpired() {
		utils.Error(ctx, 400, "邀请已过期")
		return
	}

	// 获取球队和俱乐部信息
	team := inv.Team
	clubName := ""
	if team != nil && team.Club != nil {
		clubName = team.Club.Name
	}
	creatorName := ""
	if inv.Creator != nil {
		creatorName = inv.Creator.Name
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":          inv.ID,
		"code":        inv.InviteCode,
		"teamId":      inv.TeamID,
		"teamName":    team.Name,
		"clubName":    clubName,
		"type":        inv.Type,
		"status":      inv.Status,
		"creatorName": creatorName,
		"createdAt":   utils.FormatTime(&inv.CreatedAt),
		"expiresAt":   utils.FormatTime(&inv.ExpiresAt),
		"isExpired":   inv.IsExpired(),
	})
}

// AcceptInvitation 接受邀请
func (c *TeamController) AcceptInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	userID := c.getUserID(ctx)

	if userID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	inv, err := c.teamRepo.FindInvitationByCode(code)
	if err != nil {
		utils.NotFoundError(ctx, "邀请不存在")
		return
	}

	// 检查邀请状态
	if !inv.CanAccept() {
		if inv.IsExpired() {
			utils.Error(ctx, 400, "邀请已过期")
			return
		}
		utils.Error(ctx, 400, "邀请状态不允许接受")
		return
	}

	// 根据邀请类型处理
	if inv.Type == models.InvitationTypePlayer {
		// 球员：添加到球队
		// 首先检查用户是否已在其他球队
		existing, err := c.teamRepo.GetPlayerTeam(userID)
		if err == nil && existing != nil && existing.TeamID == inv.TeamID {
			utils.Error(ctx, 400, "您已在该球队中")
			return
		}

		// 如果用户已在其他球队，先移除
		if existing != nil {
			c.teamRepo.RemovePlayer(existing.TeamID, userID, "transferred")
		}

		// 添加到新球队
		if err := c.teamRepo.AddPlayer(inv.TeamID, userID, "", ""); err != nil {
			utils.ServerError(ctx, "加入球队失败")
			return
		}

		// 同步创建 ClubPlayer（如果不存在）
		var clubPlayerCount int64
		c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND user_id = ?", inv.ClubID, userID).Count(&clubPlayerCount)
		if clubPlayerCount == 0 {
			// 获取球队年龄组
			var team models.Team
			teamAgeGroup := ""
			if err := c.db.First(&team, inv.TeamID).Error; err == nil {
				teamAgeGroup = team.AgeGroup
			}
			cp := &models.ClubPlayer{
				ClubID:   inv.ClubID,
				UserID:   userID,
				JoinDate: time.Now(),
				AgeGroup: teamAgeGroup,
				Status:   "active",
			}
			_ = c.db.Create(cp).Error
		}
	} else if inv.Type == models.InvitationTypeCoach {
		// 教练：添加到球队
		isCoach, _ := c.teamRepo.IsCoachOfTeam(userID, inv.TeamID)
		if isCoach {
			utils.Error(ctx, 400, "您已是该球队教练")
			return
		}
		if err := c.teamRepo.AddCoach(inv.TeamID, userID, models.CoachRoleAssistant); err != nil {
			utils.ServerError(ctx, "加入球队失败")
			return
		}
	}

	// 更新邀请状态
	if err := c.teamRepo.UpdateInvitationStatus(inv.ID, models.InvitationStatusAccepted); err != nil {
		utils.ServerError(ctx, "更新邀请状态失败")
		return
	}

	// 获取球队信息用于返回
	team := inv.Team
	teamName := ""
	if team != nil {
		teamName = team.Name
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"teamId":   inv.TeamID,
		"teamName": teamName,
	}, "成功加入球队")
}

// RejectInvitation 拒绝邀请
func (c *TeamController) RejectInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	userID := c.getUserID(ctx)

	if userID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	inv, err := c.teamRepo.FindInvitationByCode(code)
	if err != nil {
		utils.NotFoundError(ctx, "邀请不存在")
		return
	}

	// 检查邀请是否属于该用户
	if inv.TargetUserID == nil || *inv.TargetUserID != userID {
		utils.ForbiddenError(ctx, "无权操作此邀请")
		return
	}

	// 检查邀请状态
	if inv.Status != models.InvitationStatusPending {
		utils.Error(ctx, 400, "邀请状态不允许拒绝")
		return
	}

	// 更新邀请状态
	if err := c.teamRepo.UpdateInvitationStatus(inv.ID, models.InvitationStatusRejected); err != nil {
		utils.ServerError(ctx, "更新邀请状态失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"code": code,
	}, "已拒绝邀请")
}

// SearchUsers 搜索用户（用于邀请）
func (c *TeamController) SearchUsers(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	if keyword == "" {
		utils.ValidationError(ctx, "请输入搜索关键字")
		return
	}

	userType := ctx.Query("type") // player 或 coach

	users, err := c.teamRepo.SearchUsers(keyword, userType)
	if err != nil {
		utils.ServerError(ctx, "搜索用户失败")
		return
	}

	// 转换为响应格式
	result := make([]gin.H, 0, len(users))
	for _, user := range users {
		result = append(result, gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"nickname": user.Nickname,
			"phone":    user.Phone,
			"avatar":   user.Avatar,
			"role":     user.Role,
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":    result,
		"keyword": keyword,
	})
}

// GetMyInvitations 获取当前用户的邀请列表
func (c *TeamController) GetMyInvitations(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	// 获取当前用户手机号（用于匹配 target_phone 的邀请）
	var user models.User
	if err := c.db.First(&user, userID).Error; err != nil {
		utils.ServerError(ctx, "获取用户信息失败")
		return
	}

	status := ctx.Query("status")
	invitations, err := c.teamRepo.GetUserInvitations(userID, user.Phone, status)
	if err != nil {
		utils.ServerError(ctx, "获取邀请列表失败")
		return
	}

	// 转换为响应格式
	result := make([]gin.H, 0, len(invitations))
	for _, inv := range invitations {
		item := gin.H{
			"id":         inv.ID,
			"teamId":     inv.TeamID,
			"clubId":     inv.ClubID,
			"type":       inv.Type,
			"status":     inv.Status,
			"inviteCode": inv.InviteCode,
			"createdAt":  inv.CreatedAt,
			"expiresAt":  inv.ExpiresAt,
		}
		if inv.Team != nil {
			item["teamName"] = inv.Team.Name
			if inv.Team.Club != nil {
				item["clubName"] = inv.Team.Club.Name
			}
		}
		if inv.Creator != nil {
			item["creatorName"] = inv.Creator.Name
		}
		result = append(result, item)
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}

// CreateApplication 提交入队/试训申请
func (c *TeamController) CreateApplication(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	playerID := ctx.GetUint("userId")
	if playerID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	var req struct {
		Type   string `json:"type" binding:"required"` // join / trial
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 获取球队信息
	team, err := c.teamRepo.FindByID(uint(teamID))
	if err != nil {
		utils.NotFoundError(ctx, "球队不存在")
		return
	}

	// 检查是否已在该球队
	var existingCount int64
	c.db.Model(&models.TeamPlayer{}).Where("team_id = ? AND user_id = ? AND status = ?", teamID, playerID, "active").Count(&existingCount)
	if existingCount > 0 {
		utils.Error(ctx, 400, "您已在该球队中")
		return
	}

	// 检查是否已有待处理申请
	existingApp, _ := c.teamRepo.FindPendingApplication(uint(teamID), playerID, req.Type)
	if existingApp != nil {
		utils.Error(ctx, 400, "您已提交过申请，请勿重复提交")
		return
	}

	app := &models.TeamApplication{
		TeamID:   uint(teamID),
		ClubID:   team.ClubID,
		PlayerID: playerID,
		Type:     req.Type,
		Status:   "pending",
		Reason:   req.Reason,
	}

	if err := c.teamRepo.CreateApplication(app); err != nil {
		utils.ServerError(ctx, "提交申请失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":     app.ID,
		"status": app.Status,
	}, "申请提交成功")
}

// GetTeamApplications 获取球队的申请列表（俱乐部/教练视角）
func (c *TeamController) GetTeamApplications(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	status := ctx.Query("status")
	apps, err := c.teamRepo.GetApplications(uint(teamID), status)
	if err != nil {
		utils.ServerError(ctx, "获取申请列表失败")
		return
	}

	result := make([]gin.H, 0, len(apps))
	for _, app := range apps {
		item := gin.H{
			"id":           app.ID,
			"teamId":       app.TeamID,
			"clubId":       app.ClubID,
			"playerId":     app.PlayerID,
			"type":         app.Type,
			"status":       app.Status,
			"reason":       app.Reason,
			"responseNote": app.ResponseNote,
			"createdAt":    app.CreatedAt,
		}
		if app.Player != nil {
			item["player"] = gin.H{
				"id":       app.Player.ID,
				"name":     app.Player.Name,
				"nickname": app.Player.Nickname,
				"avatar":   app.Player.Avatar,
				"position": app.Player.Position,
			}
		}
		if app.Reviewer != nil {
			item["reviewerName"] = app.Reviewer.Name
		}
		result = append(result, item)
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}

// ReviewApplication 审核入队/试训申请
func (c *TeamController) ReviewApplication(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的申请ID")
		return
	}

	reviewerID := ctx.GetUint("userId")
	if reviewerID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	var req struct {
		Status       string `json:"status" binding:"required"` // approved / rejected
		ResponseNote string `json:"responseNote"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 获取申请信息
	app, err := c.teamRepo.GetApplicationByID(uint(appID))
	if err != nil {
		utils.NotFoundError(ctx, "申请不存在")
		return
	}

	if app.TeamID != uint(teamID) {
		utils.ForbiddenError(ctx, "无权操作此申请")
		return
	}

	if app.Status != "pending" {
		utils.Error(ctx, 400, "申请已处理")
		return
	}

	// 更新申请状态
	if err := c.teamRepo.ReviewApplication(uint(appID), req.Status, req.ResponseNote, reviewerID); err != nil {
		utils.ServerError(ctx, "审核失败")
		return
	}

	// 如果通过，自动创建 TeamPlayer 关联
	if req.Status == "approved" {
		source := "applied"
		if app.Type == "trial" {
			source = "trial"
		}
		// 检查是否已存在
		var existingCount int64
		c.db.Model(&models.TeamPlayer{}).Where("team_id = ? AND user_id = ?", app.TeamID, app.PlayerID).Count(&existingCount)
		if existingCount == 0 {
			tp := &models.TeamPlayer{
				TeamID:   app.TeamID,
				UserID:   app.PlayerID,
				Status:   "active",
				Source:   source,
				JoinedAt: time.Now(),
			}
			_ = c.db.Create(tp).Error
		}

		// 同步创建 ClubPlayer（如果不存在）
		var clubPlayerCount int64
		c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND user_id = ?", app.ClubID, app.PlayerID).Count(&clubPlayerCount)
		if clubPlayerCount == 0 {
			// 获取球队年龄组
			var team models.Team
			teamAgeGroup := ""
			if err := c.db.First(&team, app.TeamID).Error; err == nil {
				teamAgeGroup = team.AgeGroup
			}
			cp := &models.ClubPlayer{
				ClubID:   app.ClubID,
				UserID:   app.PlayerID,
				JoinDate: time.Now(),
				AgeGroup: teamAgeGroup,
				Status:   "active",
			}
			_ = c.db.Create(cp).Error
		}
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":     app.ID,
		"status": req.Status,
	}, "审核完成")
}

// GetMyApplications 获取我提交的申请列表
func (c *TeamController) GetMyApplications(ctx *gin.Context) {
	playerID := ctx.GetUint("userId")
	if playerID == 0 {
		utils.Error(ctx, 401, "请先登录")
		return
	}

	status := ctx.Query("status")
	apps, err := c.teamRepo.GetMyApplications(playerID, status)
	if err != nil {
		utils.ServerError(ctx, "获取申请列表失败")
		return
	}

	result := make([]gin.H, 0, len(apps))
	for _, app := range apps {
		item := gin.H{
			"id":           app.ID,
			"teamId":       app.TeamID,
			"clubId":       app.ClubID,
			"type":         app.Type,
			"status":       app.Status,
			"reason":       app.Reason,
			"responseNote": app.ResponseNote,
			"createdAt":    app.CreatedAt,
		}
		if app.Team != nil {
			item["teamName"] = app.Team.Name
			if app.Team.Club != nil {
				item["clubName"] = app.Team.Club.Name
			}
		}
		result = append(result, item)
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}
