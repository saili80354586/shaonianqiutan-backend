package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// TeamController 统一球队管理控制器
// 同时支持俱乐部管理员和教练访问
type TeamController struct {
	teamRepo            *repositories.TeamRepository
	weeklyReportRepo    *repositories.WeeklyReportRepository
	matchSummaryRepo    *repositories.MatchSummaryRepository
	activityRepo        *repositories.ActivityRepository
	clubRepo            *repositories.ClubRepository
	physicalTestRepo    *repositories.PhysicalTestRepository
	db                  *gorm.DB
	notificationService *services.NotificationService
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

func (c *TeamController) SetNotificationService(notificationService *services.NotificationService) {
	c.notificationService = notificationService
}

// getAccessContext 获取访问上下文
func (c *TeamController) getAccessContext(ctx *gin.Context) *middleware.TeamAccessContext {
	return middleware.GetTeamAccessContext(ctx)
}

// getUserID 获取当前用户ID
func (c *TeamController) getUserID(ctx *gin.Context) uint {
	return middleware.GetUserID(ctx)
}

func (c *TeamController) ensureClubOwner(ctx *gin.Context, clubID uint) bool {
	userID := c.getUserID(ctx)
	if userID == 0 {
		utils.ForbiddenError(ctx, "无权限访问该俱乐部")
		return false
	}
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return false
	}

	var count int64
	if err := c.db.Model(&models.Club{}).Where("id = ? AND user_id = ?", clubID, userID).Count(&count).Error; err != nil {
		utils.ServerError(ctx, "权限校验失败")
		return false
	}
	if count == 0 {
		utils.ForbiddenError(ctx, "无权限访问该俱乐部")
		return false
	}
	return true
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
		c.notifyWeeklyReportCreated(playerID, report.ID, uint(teamID), weekStart, weekEnd)
	}

	utils.SuccessResponse(ctx, gin.H{
		"created": created,
		"failed":  failed,
		"message": "成功创建 " + strconv.Itoa(created) + " 份周报",
	})
}

func (c *TeamController) notifyWeeklyReportCreated(playerID uint, reportID uint, teamID uint, weekStart time.Time, weekEnd time.Time) {
	if c.notificationService == nil {
		return
	}

	weekLabel := weekStart.Format("2006-01-02") + " 至 " + weekEnd.Format("2006-01-02")
	if err := c.notificationService.NotifyWeeklyReportCreated(playerID, "教练", "", weekLabel, reportID); err != nil {
		log.Printf("发送周报通知失败 (teamID=%d reportID=%d playerID=%d): %v", teamID, reportID, playerID, err)
	}
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
			"id":                  r.ID,
			"playerId":            r.PlayerID,
			"playerName":          playerName,
			"teamId":              r.TeamID,
			"coachId":             r.CoachID,
			"weekStart":           r.WeekStart,
			"weekEnd":             r.WeekEnd,
			"knowledgeSummary":    r.KnowledgeSummary,
			"tacticalContent":     r.TacticalContent,
			"physicalCondition":   r.PhysicalCondition,
			"matchPerformance":    r.MatchPerformance,
			"selfAttitudeRating":  r.SelfAttitudeRating,
			"selfTechniqueRating": r.SelfTechniqueRating,
			"selfTeamworkRating":  r.SelfTeamworkRating,
			"improvements":        r.ImprovementsDetail,
			"reviewStatus":        r.ReviewStatus,
			"reviewComment":       r.ReviewComment,
			"coachAttitudeRating": r.CoachAttitudeRating,
			"createdAt":           utils.FormatTime(&r.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
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
		Location      string `json:"location"`    // home/away/neutral
		MatchFormat   string `json:"matchFormat"` // 5人制/8人制/11人制
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

// GetTeamCalendar 获取球队日历聚合数据
func (c *TeamController) GetTeamCalendar(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	startDate, endDate := parseCalendarRange(ctx.Query("startDate"), ctx.Query("endDate"))
	typeSet := parseCalendarTypes(ctx.Query("types"))
	statusFilter := strings.TrimSpace(ctx.Query("status"))
	accessCtx := c.getAccessContext(ctx)
	clubID := uint(0)
	if accessCtx != nil {
		clubID = accessCtx.ClubID
	}

	payload, err := c.buildTeamCalendarPayload(uint(teamID), clubID, startDate, endDate, typeSet, statusFilter, "staff")
	if err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, payload)
}

// GetTeamMonthlyReport 获取球队月度运营报告
func (c *TeamController) GetTeamMonthlyReport(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	monthLabel, startDate, endDate := parseMonthlyReportRange(ctx.Query("month"))

	payload, err := c.buildTeamMonthlyReportPayload(uint(teamID), monthLabel, startDate, endDate)
	if err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, payload)
}

// GetTeamMonthlyReportArchive 获取球队月度报告固化归档
func (c *TeamController) GetTeamMonthlyReportArchive(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	monthLabel, _, _ := parseMonthlyReportRange(ctx.Query("month"))
	archive, payload, err := c.getTeamMonthlyReportArchive(uint(teamID), monthLabel)
	if err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}
	if archive == nil {
		utils.SuccessResponse(ctx, gin.H{"archived": false, "month": monthLabel})
		return
	}

	selectedArchive := *archive
	selectedPayload := payload
	var selectedVersion *models.TeamMonthlyReportArchiveVersion
	if versionParam := strings.TrimSpace(ctx.Query("version")); versionParam != "" {
		versionNumber, err := strconv.Atoi(versionParam)
		if err != nil || versionNumber < 1 {
			utils.ValidationError(ctx, "无效的归档版本")
			return
		}

		version, versionPayload, err := c.getTeamMonthlyReportArchiveVersion(archive.ID, versionNumber)
		if err != nil {
			utils.ServerError(ctx, err.Error())
			return
		}
		if version == nil {
			utils.NotFoundError(ctx, "归档版本不存在")
			return
		}
		selectedArchive.Version = version.Version
		selectedArchive.Snapshot = version.Snapshot
		selectedArchive.ArchivedBy = version.ArchivedBy
		selectedArchive.ArchivedAt = version.ArchivedAt
		selectedPayload = versionPayload
		selectedVersion = version
	} else {
		version, _, err := c.getTeamMonthlyReportArchiveVersion(archive.ID, normalizedMonthlyArchiveVersion(archive.Version))
		if err != nil {
			utils.ServerError(ctx, err.Error())
			return
		}
		selectedVersion = version
	}

	utils.SuccessResponse(ctx, gin.H{
		"archived":       true,
		"id":             archive.ID,
		"teamId":         archive.TeamID,
		"month":          archive.Month,
		"version":        normalizedMonthlyArchiveVersion(selectedArchive.Version),
		"versions":       c.monthlyReportArchiveVersionSummaries(archive),
		"archivedBy":     selectedArchive.ArchivedBy,
		"archivedByUser": c.monthlyReportArchiveUserSummary(selectedArchive.ArchivedBy),
		"archivedAt":     selectedArchive.ArchivedAt.Format(time.RFC3339),
		"review":         c.monthlyReportArchiveReviewSummary(selectedVersion),
		"snapshot":       selectedPayload,
	})
}

// SaveTeamMonthlyReportArchive 固化球队月度报告快照
func (c *TeamController) SaveTeamMonthlyReportArchive(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	monthLabel, startDate, endDate := parseMonthlyReportRange(ctx.Query("month"))
	payload, err := c.buildTeamMonthlyReportPayload(uint(teamID), monthLabel, startDate, endDate)
	if err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	snapshot, err := json.Marshal(payload)
	if err != nil {
		utils.ServerError(ctx, "月报快照序列化失败")
		return
	}

	userID := c.getUserID(ctx)
	if userID == 0 {
		utils.ForbiddenError(ctx, "无权限归档月报")
		return
	}

	now := time.Now()
	archive := models.TeamMonthlyReportArchive{}
	if err := c.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Where("team_id = ? AND month = ?", uint(teamID), monthLabel).First(&archive).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("读取月报归档失败")
		}
		if err == gorm.ErrRecordNotFound {
			archive = models.TeamMonthlyReportArchive{
				TeamID:     uint(teamID),
				Month:      monthLabel,
				Version:    1,
				Snapshot:   string(snapshot),
				ArchivedBy: userID,
				ArchivedAt: now,
			}
			if err := tx.Create(&archive).Error; err != nil {
				return fmt.Errorf("保存月报归档失败")
			}
		} else {
			currentVersion := normalizedMonthlyArchiveVersion(archive.Version)
			if err := c.ensureMonthlyArchiveVersion(tx, archive, currentVersion); err != nil {
				return err
			}
			archive.Version = currentVersion + 1
			archive.Snapshot = string(snapshot)
			archive.ArchivedBy = userID
			archive.ArchivedAt = now
			if err := tx.Save(&archive).Error; err != nil {
				return fmt.Errorf("更新月报归档失败")
			}
		}

		version := models.TeamMonthlyReportArchiveVersion{
			ArchiveID:  archive.ID,
			TeamID:     archive.TeamID,
			Month:      archive.Month,
			Version:    normalizedMonthlyArchiveVersion(archive.Version),
			Snapshot:   archive.Snapshot,
			ArchivedBy: archive.ArchivedBy,
			ArchivedAt: archive.ArchivedAt,
		}
		if err := tx.Create(&version).Error; err != nil {
			return fmt.Errorf("保存月报归档版本失败")
		}
		return nil
	}); err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"archived":       true,
		"id":             archive.ID,
		"teamId":         archive.TeamID,
		"month":          archive.Month,
		"version":        normalizedMonthlyArchiveVersion(archive.Version),
		"versions":       c.monthlyReportArchiveVersionSummaries(&archive),
		"archivedBy":     archive.ArchivedBy,
		"archivedByUser": c.monthlyReportArchiveUserSummary(archive.ArchivedBy),
		"archivedAt":     archive.ArchivedAt.Format(time.RFC3339),
		"snapshot":       payload,
	})
}

// ReviewTeamMonthlyReportArchive 标记月报归档版本审阅结果
func (c *TeamController) ReviewTeamMonthlyReportArchive(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		Month           string   `json:"month"`
		Version         int      `json:"version"`
		Status          string   `json:"status"`
		Note            string   `json:"note"`
		AdjustmentItems []string `json:"adjustmentItems"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "请求参数无效")
		return
	}

	status := strings.TrimSpace(req.Status)
	if status != "confirmed" && status != "needs_revision" && status != "revision_submitted" {
		utils.ValidationError(ctx, "无效的审阅状态")
		return
	}

	userID := c.getUserID(ctx)
	if userID == 0 {
		utils.ForbiddenError(ctx, "无权限审阅月报")
		return
	}

	monthLabel, _, _ := parseMonthlyReportRange(req.Month)
	note := strings.TrimSpace(req.Note)
	if len([]rune(note)) > 500 {
		utils.ValidationError(ctx, "审阅备注不能超过500字")
		return
	}

	archive := models.TeamMonthlyReportArchive{}
	version := models.TeamMonthlyReportArchiveVersion{}
	now := time.Now()
	if err := c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ? AND month = ?", uint(teamID), monthLabel).First(&archive).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("月报归档不存在")
			}
			return fmt.Errorf("读取月报归档失败")
		}

		versionNumber := req.Version
		if versionNumber < 1 {
			versionNumber = normalizedMonthlyArchiveVersion(archive.Version)
		}
		if err := c.ensureMonthlyArchiveVersion(tx, archive, versionNumber); err != nil {
			return err
		}
		if err := tx.Where("archive_id = ? AND version = ?", archive.ID, versionNumber).First(&version).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("归档版本不存在")
			}
			return fmt.Errorf("读取月报归档版本失败")
		}

		version.ReviewStatus = status
		version.ReviewNote = note
		version.ReviewedBy = &userID
		version.ReviewedAt = &now
		if err := tx.Save(&version).Error; err != nil {
			return fmt.Errorf("保存月报审阅结果失败")
		}
		event := models.TeamMonthlyReportArchiveReviewEvent{
			ArchiveID: archive.ID,
			VersionID: version.ID,
			TeamID:    archive.TeamID,
			Month:     archive.Month,
			Version:   normalizedMonthlyArchiveVersion(version.Version),
			Status:    status,
			Note:      note,
			ActorID:   userID,
			CreatedAt: now,
		}
		if err := tx.Create(&event).Error; err != nil {
			return fmt.Errorf("保存月报审阅历史失败")
		}
		if status == "needs_revision" {
			if err := c.ensureMonthlyReportAdjustmentItems(tx, archive, version, req.AdjustmentItems, userID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"archived": true,
		"id":       archive.ID,
		"teamId":   archive.TeamID,
		"month":    archive.Month,
		"version":  normalizedMonthlyArchiveVersion(version.Version),
		"versions": c.monthlyReportArchiveVersionSummaries(&archive),
		"review":   c.monthlyReportArchiveReviewSummary(&version),
	})
}

// UpdateTeamMonthlyReportArchiveAdjustment 更新月报整改项状态
func (c *TeamController) UpdateTeamMonthlyReportArchiveAdjustment(ctx *gin.Context) {
	if c.db == nil {
		utils.ServerError(ctx, "数据库未初始化")
		return
	}

	itemID, err := strconv.ParseUint(ctx.Param("itemId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的整改项ID")
		return
	}

	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "请求参数无效")
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "open" && status != "completed" {
		utils.ValidationError(ctx, "无效的整改项状态")
		return
	}

	userID := c.getUserID(ctx)
	if userID == 0 {
		utils.ForbiddenError(ctx, "无权限更新整改项")
		return
	}

	var item models.TeamMonthlyReportArchiveAdjustmentItem
	if err := c.db.Where("id = ? AND team_id = ?", uint(itemID), uint(teamID)).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundError(ctx, "整改项不存在")
			return
		}
		utils.ServerError(ctx, "读取整改项失败")
		return
	}

	now := time.Now()
	item.Status = status
	if status == "completed" {
		item.CompletedBy = &userID
		item.CompletedAt = &now
	} else {
		item.CompletedBy = nil
		item.CompletedAt = nil
	}
	if err := c.db.Save(&item).Error; err != nil {
		utils.ServerError(ctx, "保存整改项失败")
		return
	}

	utils.SuccessResponse(ctx, c.monthlyReportArchiveAdjustmentItemSummary(item))
}

func (c *TeamController) getTeamMonthlyReportArchive(teamID uint, monthLabel string) (*models.TeamMonthlyReportArchive, gin.H, error) {
	var archive models.TeamMonthlyReportArchive
	err := c.db.Where("team_id = ? AND month = ?", teamID, monthLabel).First(&archive).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("读取月报归档失败")
	}

	var payload gin.H
	if err := json.Unmarshal([]byte(archive.Snapshot), &payload); err != nil {
		return nil, nil, fmt.Errorf("月报归档数据解析失败")
	}
	return &archive, payload, nil
}

func (c *TeamController) getTeamMonthlyReportArchiveVersion(archiveID uint, versionNumber int) (*models.TeamMonthlyReportArchiveVersion, gin.H, error) {
	var version models.TeamMonthlyReportArchiveVersion
	err := c.db.Where("archive_id = ? AND version = ?", archiveID, versionNumber).First(&version).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("读取月报归档版本失败")
	}

	var payload gin.H
	if err := json.Unmarshal([]byte(version.Snapshot), &payload); err != nil {
		return nil, nil, fmt.Errorf("月报归档版本数据解析失败")
	}
	return &version, payload, nil
}

func normalizedMonthlyArchiveVersion(version int) int {
	if version < 1 {
		return 1
	}
	return version
}

func (c *TeamController) ensureMonthlyArchiveVersion(tx *gorm.DB, archive models.TeamMonthlyReportArchive, version int) error {
	var count int64
	if err := tx.Model(&models.TeamMonthlyReportArchiveVersion{}).
		Where("archive_id = ? AND version = ?", archive.ID, version).
		Count(&count).Error; err != nil {
		return fmt.Errorf("读取月报归档版本失败")
	}
	if count > 0 {
		return nil
	}

	legacyVersion := models.TeamMonthlyReportArchiveVersion{
		ArchiveID:  archive.ID,
		TeamID:     archive.TeamID,
		Month:      archive.Month,
		Version:    version,
		Snapshot:   archive.Snapshot,
		ArchivedBy: archive.ArchivedBy,
		ArchivedAt: archive.ArchivedAt,
	}
	if err := tx.Create(&legacyVersion).Error; err != nil {
		return fmt.Errorf("补齐历史月报归档版本失败")
	}
	return nil
}

func (c *TeamController) monthlyReportArchiveVersionSummaries(archive *models.TeamMonthlyReportArchive) []gin.H {
	if archive == nil {
		return []gin.H{}
	}

	var versions []models.TeamMonthlyReportArchiveVersion
	if err := c.db.Where("archive_id = ?", archive.ID).
		Order("version DESC").
		Limit(6).
		Find(&versions).Error; err != nil {
		return []gin.H{}
	}

	items := make([]gin.H, 0, len(versions))
	for _, version := range versions {
		items = append(items, gin.H{
			"id":             version.ID,
			"version":        normalizedMonthlyArchiveVersion(version.Version),
			"archivedBy":     version.ArchivedBy,
			"archivedByUser": c.monthlyReportArchiveUserSummary(version.ArchivedBy),
			"archivedAt":     version.ArchivedAt.Format(time.RFC3339),
			"review":         c.monthlyReportArchiveReviewSummary(&version),
		})
	}
	if len(items) == 0 {
		items = append(items, gin.H{
			"id":             archive.ID,
			"version":        normalizedMonthlyArchiveVersion(archive.Version),
			"archivedBy":     archive.ArchivedBy,
			"archivedByUser": c.monthlyReportArchiveUserSummary(archive.ArchivedBy),
			"archivedAt":     archive.ArchivedAt.Format(time.RFC3339),
			"review":         c.monthlyReportArchiveReviewSummary(nil),
		})
	}
	return items
}

func (c *TeamController) monthlyReportArchiveReviewSummary(version *models.TeamMonthlyReportArchiveVersion) gin.H {
	status := "pending"
	if version != nil && strings.TrimSpace(version.ReviewStatus) != "" {
		status = strings.TrimSpace(version.ReviewStatus)
	}

	item := gin.H{
		"status": status,
	}
	if version == nil {
		return item
	}
	if strings.TrimSpace(version.ReviewNote) != "" {
		item["note"] = version.ReviewNote
	}
	if version.ReviewedBy != nil && *version.ReviewedBy > 0 {
		item["reviewedBy"] = *version.ReviewedBy
		item["reviewedByUser"] = c.monthlyReportArchiveUserSummary(*version.ReviewedBy)
	}
	if version.ReviewedAt != nil {
		item["reviewedAt"] = version.ReviewedAt.Format(time.RFC3339)
	}
	item["history"] = c.monthlyReportArchiveReviewHistory(version)
	item["adjustments"] = c.monthlyReportArchiveAdjustmentItems(version)
	return item
}

func (c *TeamController) ensureMonthlyReportAdjustmentItems(tx *gorm.DB, archive models.TeamMonthlyReportArchive, version models.TeamMonthlyReportArchiveVersion, rawItems []string, userID uint) error {
	seen := map[string]bool{}
	for _, raw := range rawItems {
		content := strings.TrimSpace(raw)
		if content == "" || seen[content] {
			continue
		}
		seen[content] = true

		var count int64
		if err := tx.Model(&models.TeamMonthlyReportArchiveAdjustmentItem{}).
			Where("version_id = ? AND content = ?", version.ID, content).
			Count(&count).Error; err != nil {
			return fmt.Errorf("读取月报整改项失败")
		}
		if count > 0 {
			continue
		}

		item := models.TeamMonthlyReportArchiveAdjustmentItem{
			ArchiveID: archive.ID,
			VersionID: version.ID,
			TeamID:    archive.TeamID,
			Month:     archive.Month,
			Version:   normalizedMonthlyArchiveVersion(version.Version),
			Content:   content,
			Status:    "open",
			CreatedBy: userID,
		}
		if err := tx.Create(&item).Error; err != nil {
			return fmt.Errorf("保存月报整改项失败")
		}
	}
	return nil
}

func (c *TeamController) monthlyReportArchiveAdjustmentItems(version *models.TeamMonthlyReportArchiveVersion) []gin.H {
	if c.db == nil || version == nil {
		return []gin.H{}
	}

	var items []models.TeamMonthlyReportArchiveAdjustmentItem
	if err := c.db.Where("version_id = ?", version.ID).
		Order("created_at ASC, id ASC").
		Find(&items).Error; err != nil {
		return []gin.H{}
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, c.monthlyReportArchiveAdjustmentItemSummary(item))
	}
	return result
}

func (c *TeamController) monthlyReportArchiveAdjustmentItemSummary(item models.TeamMonthlyReportArchiveAdjustmentItem) gin.H {
	result := gin.H{
		"id":            item.ID,
		"content":       item.Content,
		"status":        item.Status,
		"createdBy":     item.CreatedBy,
		"createdByUser": c.monthlyReportArchiveUserSummary(item.CreatedBy),
		"createdAt":     item.CreatedAt.Format(time.RFC3339),
	}
	if item.CompletedBy != nil && *item.CompletedBy > 0 {
		result["completedBy"] = *item.CompletedBy
		result["completedByUser"] = c.monthlyReportArchiveUserSummary(*item.CompletedBy)
	}
	if item.CompletedAt != nil {
		result["completedAt"] = item.CompletedAt.Format(time.RFC3339)
	}
	return result
}

func (c *TeamController) monthlyReportArchiveReviewHistory(version *models.TeamMonthlyReportArchiveVersion) []gin.H {
	if c.db == nil || version == nil {
		return []gin.H{}
	}

	var events []models.TeamMonthlyReportArchiveReviewEvent
	if err := c.db.Where("version_id = ?", version.ID).
		Order("created_at ASC, id ASC").
		Limit(12).
		Find(&events).Error; err != nil {
		return []gin.H{}
	}

	items := make([]gin.H, 0, len(events))
	for _, event := range events {
		item := gin.H{
			"id":        event.ID,
			"status":    event.Status,
			"actorId":   event.ActorID,
			"actorUser": c.monthlyReportArchiveUserSummary(event.ActorID),
			"createdAt": event.CreatedAt.Format(time.RFC3339),
		}
		if strings.TrimSpace(event.Note) != "" {
			item["note"] = event.Note
		}
		items = append(items, item)
	}

	if len(items) == 0 && version.ReviewedAt != nil && version.ReviewedBy != nil && *version.ReviewedBy > 0 {
		item := gin.H{
			"status":    strings.TrimSpace(version.ReviewStatus),
			"actorId":   *version.ReviewedBy,
			"actorUser": c.monthlyReportArchiveUserSummary(*version.ReviewedBy),
			"createdAt": version.ReviewedAt.Format(time.RFC3339),
		}
		if strings.TrimSpace(version.ReviewNote) != "" {
			item["note"] = version.ReviewNote
		}
		items = append(items, item)
	}

	return items
}

func (c *TeamController) monthlyReportArchiveUserSummary(userID uint) gin.H {
	if c.db == nil || userID == 0 {
		return nil
	}

	var user models.User
	if err := c.db.Select("id", "nickname", "name", "role", "current_role").First(&user, userID).Error; err != nil {
		return nil
	}

	displayName := strings.TrimSpace(user.Nickname)
	if displayName == "" {
		displayName = strings.TrimSpace(user.Name)
	}
	if displayName == "" {
		displayName = fmt.Sprintf("用户%d", user.ID)
	}
	role := user.CurrentRole
	if role == "" {
		role = user.Role
	}

	return gin.H{
		"id":          user.ID,
		"displayName": displayName,
		"role":        role,
	}
}

func (c *TeamController) buildTeamMonthlyReportPayload(teamID uint, monthLabel string, startDate time.Time, endDate time.Time) (gin.H, error) {
	var team models.Team
	if err := c.db.First(&team, teamID).Error; err != nil {
		return nil, fmt.Errorf("获取球队信息失败")
	}

	var activePlayers int64
	if err := c.db.Model(&models.TeamPlayer{}).
		Where("team_id = ? AND status = ?", teamID, "active").
		Count(&activePlayers).Error; err != nil {
		return nil, fmt.Errorf("获取球队球员失败")
	}

	var plans []models.TrainingPlan
	if err := c.db.
		Where("team_id = ? AND start_time >= ? AND start_time < ?", teamID, startDate, endDate).
		Order("start_time ASC").
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("获取训练计划失败")
	}

	planIDs := make([]uint, 0, len(plans))
	completedTrainingCount := 0
	reviewCount := 0
	recentReviews := make([]gin.H, 0)
	completionStatusCount := map[string]int{
		"excellent": 0,
		"good":      0,
		"normal":    0,
		"poor":      0,
		"unknown":   0,
	}
	expectedAttendanceTotal := 0
	for _, plan := range plans {
		planIDs = append(planIDs, plan.ID)
		expectedAttendanceTotal += monthlyPlanExpectedPlayers(plan, int(activePlayers))
		if plan.Status == models.TrainingPlanStatusCompleted {
			completedTrainingCount++
		}
		statusKey := strings.TrimSpace(plan.CompletionStatus)
		if statusKey == "" {
			statusKey = "unknown"
		}
		if _, ok := completionStatusCount[statusKey]; !ok {
			completionStatusCount[statusKey] = 0
		}
		completionStatusCount[statusKey]++
		if plan.Summary != "" || plan.CompletionStatus != "" || plan.KeyPlayerNotes != "" || plan.NextFocus != "" {
			reviewCount++
			recentReviews = append(recentReviews, gin.H{
				"id":               plan.ID,
				"title":            plan.Title,
				"date":             plan.StartTime.Format("2006-01-02"),
				"completionStatus": plan.CompletionStatus,
				"summary":          plan.Summary,
				"keyPlayerNotes":   plan.KeyPlayerNotes,
				"nextFocus":        plan.NextFocus,
			})
		}
	}

	var attendances []models.TrainingAttendance
	if len(planIDs) > 0 {
		if err := c.db.Where("training_plan_id IN ?", planIDs).Find(&attendances).Error; err != nil {
			return nil, fmt.Errorf("获取训练出勤失败")
		}
	}
	attendanceSummary := gin.H{
		"totalExpected":  expectedAttendanceTotal,
		"recorded":       len(attendances),
		"present":        0,
		"leave":          0,
		"absent":         0,
		"late":           0,
		"unmarked":       0,
		"attendanceRate": 0.0,
	}
	for _, attendance := range attendances {
		key := string(attendance.Status)
		if _, ok := attendanceSummary[key]; ok {
			attendanceSummary[key] = attendanceSummary[key].(int) + 1
		}
	}
	unmarked := expectedAttendanceTotal - len(attendances)
	if unmarked < 0 {
		unmarked = 0
	}
	attendanceSummary["unmarked"] = unmarked
	attendanceSummary["attendanceRate"] = monthlyReportRate(attendanceSummary["present"].(int)+attendanceSummary["late"].(int), expectedAttendanceTotal)

	var schedules []models.MatchSchedule
	if err := c.db.
		Where("team_id = ? AND match_time >= ? AND match_time < ?", teamID, startDate, endDate).
		Order("match_time ASC").
		Find(&schedules).Error; err != nil {
		return nil, fmt.Errorf("获取比赛计划失败")
	}
	completedMatchCount := 0
	summaryGeneratedCount := 0
	pendingSummaryCount := 0
	wins := 0
	draws := 0
	losses := 0
	for _, schedule := range schedules {
		if schedule.Status == models.MatchScheduleStatusCompleted {
			completedMatchCount++
			if schedule.MatchSummaryID == nil {
				pendingSummaryCount++
			}
		}
		if schedule.MatchSummaryID != nil {
			summaryGeneratedCount++
		}
		if schedule.HomeScore != nil && schedule.AwayScore != nil {
			switch {
			case *schedule.HomeScore > *schedule.AwayScore:
				wins++
			case *schedule.HomeScore == *schedule.AwayScore:
				draws++
			default:
				losses++
			}
		}
	}

	physicalItems, err := c.getCalendarPhysicalTests(team.ClubID, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取体测安排失败")
	}
	physicalIDs := make([]uint, 0, len(physicalItems))
	completedPhysicalCount := 0
	for _, item := range physicalItems {
		if sourceID, ok := item["sourceId"].(uint); ok {
			physicalIDs = append(physicalIDs, sourceID)
		}
		status := fmt.Sprint(item["status"])
		if status == "completed" || status == "report_generated" {
			completedPhysicalCount++
		}
	}
	physicalRecordCount := int64(0)
	if len(physicalIDs) > 0 {
		if err := c.db.Model(&models.PhysicalTestRecord{}).
			Where("activity_id IN ?", physicalIDs).
			Count(&physicalRecordCount).Error; err != nil {
			return nil, fmt.Errorf("获取体测记录失败")
		}
	}

	var periods []models.WeeklyReportPeriod
	if err := c.db.
		Where("team_id = ? AND week_start < ? AND week_end >= ?", teamID, endDate, startDate).
		Order("week_start ASC").
		Find(&periods).Error; err != nil {
		return nil, fmt.Errorf("获取周报周期失败")
	}
	weeklyTotalPlayers := 0
	weeklySubmittedCount := 0
	weeklyPendingCount := 0
	weeklyReviewedCount := 0
	for _, period := range periods {
		weeklyTotalPlayers += period.TotalPlayers
		weeklySubmittedCount += period.SubmittedCount
		weeklyPendingCount += period.PendingCount
		weeklyReviewedCount += period.ReviewedCount
	}

	sort.Slice(recentReviews, func(i, j int) bool {
		return fmt.Sprint(recentReviews[i]["date"]) > fmt.Sprint(recentReviews[j]["date"])
	})
	if len(recentReviews) > 5 {
		recentReviews = recentReviews[:5]
	}

	recommendations := buildMonthlyTrainingRecommendations(monthlyRecommendationInput{
		TrainingCount:            len(plans),
		CompletedTrainingCount:   completedTrainingCount,
		ReviewCount:              reviewCount,
		ActivePlayerCount:        int(activePlayers),
		ExpectedAttendance:       expectedAttendanceTotal,
		PresentAttendance:        attendanceSummary["present"].(int),
		LateAttendance:           attendanceSummary["late"].(int),
		UnmarkedAttendance:       attendanceSummary["unmarked"].(int),
		MatchCount:               len(schedules),
		PendingMatchSummaryCount: pendingSummaryCount,
		Wins:                     wins,
		Draws:                    draws,
		Losses:                   losses,
		PhysicalTestCount:        len(physicalItems),
		CompletedPhysicalCount:   completedPhysicalCount,
		PhysicalRecordCount:      int(physicalRecordCount),
		WeeklyTotalPlayers:       weeklyTotalPlayers,
		WeeklySubmittedCount:     weeklySubmittedCount,
		WeeklyReviewedCount:      weeklyReviewedCount,
		CompletionStatusCount:    completionStatusCount,
	})
	aiInsights := buildMonthlyAITrainingInsights(monthlyRecommendationInput{
		TrainingCount:            len(plans),
		CompletedTrainingCount:   completedTrainingCount,
		ReviewCount:              reviewCount,
		ActivePlayerCount:        int(activePlayers),
		ExpectedAttendance:       expectedAttendanceTotal,
		PresentAttendance:        attendanceSummary["present"].(int),
		LateAttendance:           attendanceSummary["late"].(int),
		UnmarkedAttendance:       attendanceSummary["unmarked"].(int),
		MatchCount:               len(schedules),
		PendingMatchSummaryCount: pendingSummaryCount,
		Wins:                     wins,
		Draws:                    draws,
		Losses:                   losses,
		PhysicalTestCount:        len(physicalItems),
		CompletedPhysicalCount:   completedPhysicalCount,
		PhysicalRecordCount:      int(physicalRecordCount),
		WeeklyTotalPlayers:       weeklyTotalPlayers,
		WeeklySubmittedCount:     weeklySubmittedCount,
		WeeklyReviewedCount:      weeklyReviewedCount,
		CompletionStatusCount:    completionStatusCount,
	})

	return gin.H{
		"teamId":   team.ID,
		"teamName": team.Name,
		"ageGroup": team.AgeGroup,
		"month":    monthLabel,
		"range": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"overview": gin.H{
			"playerCount":               activePlayers,
			"trainingCount":             len(plans),
			"completedTrainingCount":    completedTrainingCount,
			"trainingCompletionRate":    monthlyReportRate(completedTrainingCount, len(plans)),
			"attendanceRate":            attendanceSummary["attendanceRate"],
			"matchCount":                len(schedules),
			"completedMatchCount":       completedMatchCount,
			"physicalTestCount":         len(physicalItems),
			"weeklyPeriodCount":         len(periods),
			"weeklySubmissionRate":      monthlyReportRate(weeklySubmittedCount, weeklyTotalPlayers),
			"weeklyReviewRate":          monthlyReportRate(weeklyReviewedCount, weeklyTotalPlayers),
			"pendingMatchSummaryCount":  pendingSummaryCount,
			"reviewedTrainingPlanCount": reviewCount,
		},
		"training": gin.H{
			"count":                 len(plans),
			"completedCount":        completedTrainingCount,
			"reviewCount":           reviewCount,
			"completionStatusCount": completionStatusCount,
			"recentReviews":         recentReviews,
		},
		"attendance": attendanceSummary,
		"matches": gin.H{
			"count":                 len(schedules),
			"completedCount":        completedMatchCount,
			"summaryGeneratedCount": summaryGeneratedCount,
			"pendingSummaryCount":   pendingSummaryCount,
			"wins":                  wins,
			"draws":                 draws,
			"losses":                losses,
		},
		"physical": gin.H{
			"testCount":      len(physicalItems),
			"completedCount": completedPhysicalCount,
			"recordCount":    physicalRecordCount,
		},
		"weekly": gin.H{
			"periodCount":    len(periods),
			"totalPlayers":   weeklyTotalPlayers,
			"submittedCount": weeklySubmittedCount,
			"pendingCount":   weeklyPendingCount,
			"reviewedCount":  weeklyReviewedCount,
			"submissionRate": monthlyReportRate(weeklySubmittedCount, weeklyTotalPlayers),
			"reviewRate":     monthlyReportRate(weeklyReviewedCount, weeklyTotalPlayers),
		},
		"recommendations": recommendations,
		"aiInsights":      aiInsights,
	}, nil
}

func (c *TeamController) buildTeamCalendarPayload(teamID uint, clubID uint, startDate time.Time, endDate time.Time, typeSet map[string]bool, statusFilter string, audience string) (gin.H, error) {
	items := make([]gin.H, 0)
	stats := gin.H{
		"trainingCount":           0,
		"matchCount":              0,
		"physicalTestCount":       0,
		"weeklyPeriodCount":       0,
		"weeklyPendingCount":      0,
		"matchReviewPendingCount": 0,
	}

	if typeSet["training"] {
		var plans []models.TrainingPlan
		err := c.db.
			Where("team_id = ? AND start_time >= ? AND start_time < ?", uint(teamID), startDate, endDate).
			Order("start_time ASC").
			Find(&plans).Error
		if err != nil {
			return nil, fmt.Errorf("获取训练计划失败")
		}
		for _, plan := range plans {
			if !calendarStatusMatches(statusFilter, string(plan.Status), "training") {
				continue
			}
			item := gin.H{
				"id":        fmt.Sprintf("training:%d", plan.ID),
				"sourceId":  plan.ID,
				"type":      "training",
				"title":     plan.Title,
				"teamId":    plan.TeamID,
				"startTime": plan.StartTime.Format(time.RFC3339),
				"location":  plan.Location,
				"status":    string(plan.Status),
				"theme":     plan.Theme,
				"links": gin.H{
					"weeklyReportId": plan.WeeklyReportID,
					"physicalTestId": plan.PhysicalTestID,
					"matchSummaryId": nil,
				},
				"actions": calendarActions(audience, "training", []string{"view", "edit"}),
			}
			if plan.EndTime != nil {
				item["endTime"] = plan.EndTime.Format(time.RFC3339)
			}
			items = append(items, item)
			stats["trainingCount"] = stats["trainingCount"].(int) + 1
		}
	}

	if typeSet["match"] {
		var schedules []models.MatchSchedule
		err := c.db.
			Where("team_id = ? AND match_time >= ? AND match_time < ?", uint(teamID), startDate, endDate).
			Order("match_time ASC").
			Find(&schedules).Error
		if err != nil {
			return nil, fmt.Errorf("获取比赛计划失败")
		}
		pendingReviews := 0
		for _, schedule := range schedules {
			if !calendarStatusMatches(statusFilter, string(schedule.Status), "match") {
				continue
			}
			actions := []string{"view", "edit"}
			if schedule.MatchSummaryID != nil {
				actions = append(actions, "view_summary")
			} else if schedule.Status == models.MatchScheduleStatusCompleted {
				actions = append(actions, "create_summary")
				pendingReviews++
			}
			items = append(items, gin.H{
				"id":        fmt.Sprintf("match:%d", schedule.ID),
				"sourceId":  schedule.ID,
				"type":      "match",
				"title":     schedule.Name,
				"teamId":    schedule.TeamID,
				"startTime": schedule.MatchTime.Format(time.RFC3339),
				"location":  schedule.Location,
				"status":    string(schedule.Status),
				"subtype":   string(schedule.MatchType),
				"opponent":  schedule.Opponent,
				"links": gin.H{
					"weeklyReportId": nil,
					"physicalTestId": nil,
					"matchSummaryId": schedule.MatchSummaryID,
				},
				"actions": calendarActions(audience, "match", actions),
			})
			stats["matchCount"] = stats["matchCount"].(int) + 1
		}
		stats["matchReviewPendingCount"] = pendingReviews
	}

	if typeSet["physical"] {
		physicalItems, err := c.getCalendarPhysicalTests(clubID, uint(teamID), startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("获取体测安排失败")
		}
		if audience == "player" {
			for _, item := range physicalItems {
				item["actions"] = calendarActions(audience, "physical", []string{"view", "record"})
			}
		}
		for _, item := range physicalItems {
			if !calendarStatusMatches(statusFilter, fmt.Sprint(item["status"]), "physical") {
				continue
			}
			items = append(items, item)
			stats["physicalTestCount"] = stats["physicalTestCount"].(int) + 1
		}
	}

	if typeSet["weekly"] {
		var periods []models.WeeklyReportPeriod
		err := c.db.
			Where("team_id = ? AND week_start < ? AND week_end >= ?", uint(teamID), endDate, startDate).
			Order("week_start ASC").
			Find(&periods).Error
		if err != nil {
			return nil, fmt.Errorf("获取周报周期失败")
		}
		pendingCount := 0
		for _, period := range periods {
			if !calendarStatusMatches(statusFilter, period.Status, "weekly") {
				continue
			}
			pendingCount += period.PendingCount
			startTime := period.WeekStart
			if period.Deadline != nil {
				startTime = *period.Deadline
			}
			items = append(items, gin.H{
				"id":        fmt.Sprintf("weekly:%d", period.ID),
				"sourceId":  period.ID,
				"type":      "weekly",
				"title":     period.ToResponse().WeekLabel,
				"teamId":    period.TeamID,
				"startTime": startTime.Format(time.RFC3339),
				"endTime":   period.WeekEnd.Format(time.RFC3339),
				"status":    period.Status,
				"summary": gin.H{
					"totalPlayers":   period.TotalPlayers,
					"submittedCount": period.SubmittedCount,
					"pendingCount":   period.PendingCount,
					"reviewedCount":  period.ReviewedCount,
				},
				"links": gin.H{
					"weeklyReportId": period.ID,
					"weeklyPeriodId": period.ID,
					"physicalTestId": nil,
					"matchSummaryId": nil,
				},
				"actions": calendarActions(audience, "weekly", []string{"view", "remind"}),
			})
			stats["weeklyPeriodCount"] = stats["weeklyPeriodCount"].(int) + 1
		}
		stats["weeklyPendingCount"] = pendingCount
	}

	sort.Slice(items, func(i, j int) bool {
		left, _ := time.Parse(time.RFC3339, fmt.Sprint(items[i]["startTime"]))
		right, _ := time.Parse(time.RFC3339, fmt.Sprint(items[j]["startTime"]))
		return left.Before(right)
	})

	return gin.H{
		"teamId": uint(teamID),
		"range": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"items": items,
		"stats": stats,
	}, nil
}

func calendarStatusMatches(filter string, itemStatus string, itemType string) bool {
	filter = strings.TrimSpace(filter)
	if filter == "" || filter == "all" {
		return true
	}
	status := strings.TrimSpace(itemStatus)
	switch filter {
	case "open":
		return status == "draft" ||
			status == "published" ||
			status == "upcoming" ||
			status == "pending" ||
			status == "active"
	case "ongoing":
		return status == "ongoing" || (itemType == "weekly" && status == "active")
	case "completed":
		return status == "completed" ||
			status == "report_generated" ||
			status == "closed" ||
			status == "archived"
	case "cancelled":
		return status == "cancelled"
	default:
		return status == filter
	}
}

func calendarActions(audience string, itemType string, staffActions []string) []string {
	if audience != "player" {
		return staffActions
	}
	switch itemType {
	case "weekly":
		return []string{"view", "submit"}
	case "training", "match", "physical":
		return []string{"view"}
	default:
		return []string{"view"}
	}
}

func (c *TeamController) getCalendarPhysicalTests(clubID uint, teamID uint, startDate time.Time, endDate time.Time) ([]gin.H, error) {
	var teamPlayers []models.TeamPlayer
	if err := c.db.Where("team_id = ? AND status = ?", teamID, "active").Find(&teamPlayers).Error; err != nil {
		return nil, err
	}
	playerSet := make(map[uint]struct{}, len(teamPlayers))
	for _, player := range teamPlayers {
		playerSet[player.UserID] = struct{}{}
	}

	var tests []models.PhysicalTestActivity
	query := c.db.Where("start_date >= ? AND start_date < ?", startDate, endDate)
	if clubID > 0 {
		query = query.Where("club_id = ?", clubID)
	}
	if err := query.Order("start_date ASC").Find(&tests).Error; err != nil {
		return nil, err
	}

	items := make([]gin.H, 0, len(tests))
	for _, test := range tests {
		matchesTeam := false
		for _, playerID := range test.GetPlayerIDs() {
			if _, ok := playerSet[playerID]; ok {
				matchesTeam = true
				break
			}
		}
		if !matchesTeam {
			continue
		}
		item := gin.H{
			"id":        fmt.Sprintf("physical:%d", test.ID),
			"sourceId":  test.ID,
			"type":      "physical",
			"title":     test.Name,
			"teamId":    teamID,
			"startTime": test.StartDate.Format(time.RFC3339),
			"location":  test.Location,
			"status":    string(test.Status),
			"subtype":   string(test.Template),
			"links": gin.H{
				"weeklyReportId": nil,
				"physicalTestId": test.ID,
				"matchSummaryId": nil,
			},
			"actions": []string{"view", "record"},
		}
		if test.EndDate != nil {
			item["endTime"] = test.EndDate.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, nil
}

func parseCalendarRange(start string, end string) (time.Time, time.Time) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, 0)
	if parsed, err := time.Parse("2006-01-02", start); err == nil {
		startDate = parsed
	}
	if parsed, err := time.Parse("2006-01-02", end); err == nil {
		endDate = parsed.AddDate(0, 0, 1)
	}
	return startDate, endDate
}

func parseMonthlyReportRange(raw string) (string, time.Time, time.Time) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if parsed, err := time.Parse("2006-01", strings.TrimSpace(raw)); err == nil {
		startDate = parsed
	}
	return startDate.Format("2006-01"), startDate, startDate.AddDate(0, 1, 0)
}

func monthlyReportRate(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return math.Round(float64(numerator)/float64(denominator)*1000) / 10
}

func monthlyPlanExpectedPlayers(plan models.TrainingPlan, activePlayerCount int) int {
	playerIDs := plan.GetPlayerIDs()
	if len(playerIDs) > 0 {
		return len(playerIDs)
	}
	if activePlayerCount > 0 {
		return activePlayerCount
	}
	return 0
}

type monthlyRecommendationInput struct {
	TrainingCount            int
	CompletedTrainingCount   int
	ReviewCount              int
	ActivePlayerCount        int
	ExpectedAttendance       int
	PresentAttendance        int
	LateAttendance           int
	UnmarkedAttendance       int
	MatchCount               int
	PendingMatchSummaryCount int
	Wins                     int
	Draws                    int
	Losses                   int
	PhysicalTestCount        int
	CompletedPhysicalCount   int
	PhysicalRecordCount      int
	WeeklyTotalPlayers       int
	WeeklySubmittedCount     int
	WeeklyReviewedCount      int
	CompletionStatusCount    map[string]int
}

func buildMonthlyTrainingRecommendations(input monthlyRecommendationInput) []gin.H {
	recommendations := make([]gin.H, 0)
	add := func(category string, priority string, title string, reason string, action string) {
		recommendations = append(recommendations, gin.H{
			"category": category,
			"priority": priority,
			"title":    title,
			"reason":   reason,
			"action":   action,
		})
	}

	trainingCompletionRate := monthlyReportRate(input.CompletedTrainingCount, input.TrainingCount)
	attendanceRate := monthlyReportRate(input.PresentAttendance+input.LateAttendance, input.ExpectedAttendance)
	weeklySubmissionRate := monthlyReportRate(input.WeeklySubmittedCount, input.WeeklyTotalPlayers)
	weeklyReviewRate := monthlyReportRate(input.WeeklyReviewedCount, input.WeeklyTotalPlayers)

	if input.TrainingCount == 0 {
		add("training", "high", "先建立下月训练节奏", "本月没有训练计划数据，球队缺少可追踪的训练节奏。", "建议先排定未来4周训练主题，每周至少2次，并关联周报周期。")
	} else if trainingCompletionRate < 80 {
		add("training", "high", "提高训练计划完成率", fmt.Sprintf("本月训练完成率为 %.1f%%，低于80%%。", trainingCompletionRate), "建议减少临时取消，按周复盘未完成原因，并把关键训练改为优先保障事项。")
	}

	if input.ReviewCount < input.CompletedTrainingCount {
		add("review", "medium", "补齐训练后复盘", fmt.Sprintf("已完成训练 %d 次，其中 %d 次已沉淀复盘。", input.CompletedTrainingCount, input.ReviewCount), "建议每次训练结束后记录完成情况、重点球员和下次重点，支撑后续月报与家长摘要。")
	}

	if attendanceRate > 0 && attendanceRate < 85 {
		add("attendance", "high", "稳定训练出勤", fmt.Sprintf("本月训练出勤率为 %.1f%%，低于85%%。", attendanceRate), "建议优先跟进连续缺席或请假球员，训练计划提前通知家长，并在周报中同步提醒。")
	}
	if input.UnmarkedAttendance > 0 {
		add("attendance", "medium", "补齐出勤记录", fmt.Sprintf("仍有 %d 人次训练出勤未标记。", input.UnmarkedAttendance), "建议教练在训练当天完成出勤登记，避免月报和家长沟通数据失真。")
	}

	if weeklySubmissionRate > 0 && weeklySubmissionRate < 90 {
		add("weekly", "medium", "提升周报提交率", fmt.Sprintf("本月周报提交率为 %.1f%%。", weeklySubmissionRate), "建议对未提交球员发起提醒，并把训练重点、比赛反馈放进周报引导。")
	}
	if weeklyReviewRate > 0 && weeklyReviewRate < 70 {
		add("weekly", "medium", "提高周报审核覆盖", fmt.Sprintf("本月周报审核率为 %.1f%%。", weeklyReviewRate), "建议主教练固定每周审核窗口，优先点评训练出勤异常和比赛表现波动的球员。")
	}

	if input.PendingMatchSummaryCount > 0 {
		add("match", "high", "补齐赛后总结", fmt.Sprintf("有 %d 场已结束比赛待生成总结。", input.PendingMatchSummaryCount), "建议先完成比赛总结，再把共性问题反推到下周训练主题。")
	}
	if input.MatchCount > 0 && input.Losses > input.Wins {
		add("match", "medium", "用比赛问题反推训练主题", fmt.Sprintf("本月比赛战绩为 %d胜%d平%d负。", input.Wins, input.Draws, input.Losses), "建议从失球方式、对抗强度和转换效率中选择1-2个共性问题，纳入下一阶段训练模板。")
	}

	if input.PhysicalTestCount == 0 {
		add("physical", "medium", "安排阶段体测基线", "本月没有体测安排，训练效果缺少量化参照。", "建议下月安排一次基础体测，并将体测建议回流到训练计划。")
	} else if input.ActivePlayerCount > 0 && input.PhysicalRecordCount < input.ActivePlayerCount {
		add("physical", "medium", "补齐体测记录覆盖", fmt.Sprintf("本月体测记录 %d 条，当前在队球员 %d 人。", input.PhysicalRecordCount, input.ActivePlayerCount), "建议补录缺失球员体测数据，再生成训练建议。")
	}

	poorOrNormal := input.CompletionStatusCount["poor"] + input.CompletionStatusCount["normal"]
	if poorOrNormal > 0 {
		add("training", "medium", "调整训练目标难度", fmt.Sprintf("本月有 %d 次训练完成情况为基本完成或需要改进。", poorOrNormal), "建议复盘训练强度、分组和时长，下一阶段先聚焦1个核心目标。")
	}

	if len(recommendations) == 0 {
		add("training", "low", "延续当前训练节奏", "本月训练、出勤、周报和比赛复盘指标整体稳定。", "建议延续当前训练节奏，并选择1个体测或比赛指标作为下月重点提升目标。")
	}
	return recommendations
}

func buildMonthlyAITrainingInsights(input monthlyRecommendationInput) []gin.H {
	insights := make([]gin.H, 0, 3)
	add := func(title string, confidence float64, basis string, action string) {
		insights = append(insights, gin.H{
			"title":      title,
			"confidence": confidence,
			"basis":      basis,
			"action":     action,
		})
	}

	trainingCompletionRate := monthlyReportRate(input.CompletedTrainingCount, input.TrainingCount)
	attendanceRate := monthlyReportRate(input.PresentAttendance+input.LateAttendance, input.ExpectedAttendance)
	weeklySubmissionRate := monthlyReportRate(input.WeeklySubmittedCount, input.WeeklyTotalPlayers)

	if input.TrainingCount > 0 && trainingCompletionRate < 80 && attendanceRate >= 85 {
		add("训练执行节奏需要重排", 0.86, fmt.Sprintf("训练出勤率 %.1f%% 尚可，但训练完成率 %.1f%% 偏低，问题更可能来自计划排期或训练目标过满。", attendanceRate, trainingCompletionRate), "建议下月把训练计划拆成必达目标和加练目标，先保障每周核心课按时完成。")
	}
	if attendanceRate > 0 && attendanceRate < 85 {
		add("出勤稳定性是下月首要变量", 0.82, fmt.Sprintf("训练出勤率 %.1f%%，会直接影响周报反馈和比赛复盘质量。", attendanceRate), "建议筛出连续缺席、请假或迟到球员，安排教练在周报前完成一次家长沟通。")
	}
	if input.PendingMatchSummaryCount > 0 && input.ReviewCount < input.CompletedTrainingCount {
		add("比赛问题尚未充分回流训练", 0.78, fmt.Sprintf("有 %d 场比赛待总结，同时训练复盘覆盖 %d/%d。", input.PendingMatchSummaryCount, input.ReviewCount, input.CompletedTrainingCount), "建议先补齐赛后总结，再把共性问题转为下周 1-2 个训练主题。")
	}
	if weeklySubmissionRate > 0 && weeklySubmissionRate < 90 {
		add("周报提交率影响家长侧感知", 0.74, fmt.Sprintf("周报提交率 %.1f%%，家长侧可能无法稳定看到训练反馈。", weeklySubmissionRate), "建议固定每周提醒节点，并把球队日历中的训练和比赛事项带入周报提示。")
	}
	if len(insights) == 0 {
		add("当前训练闭环稳定，可进入专项提升", 0.72, "训练、出勤、比赛和周报数据未出现单点高风险。", "建议选择一个体测指标或比赛问题作为下月专项目标，并在月报归档后持续复核。")
	}

	if len(insights) > 3 {
		return insights[:3]
	}
	return insights
}

func parseCalendarTypes(raw string) map[string]bool {
	result := map[string]bool{
		"training": true,
		"match":    true,
		"physical": true,
		"weekly":   true,
	}
	if raw == "" {
		return result
	}
	result = map[string]bool{}
	var types []string
	if err := json.Unmarshal([]byte(raw), &types); err != nil {
		types = strings.Split(raw, ",")
	}
	for _, value := range types {
		value = strings.TrimSpace(value)
		if value == "all" {
			return map[string]bool{"training": true, "match": true, "physical": true, "weekly": true}
		}
		result[value] = true
	}
	return result
}

// ==================== 缺失的方法存根 ====================

// GetTeams 获取俱乐部球队列表
func (c *TeamController) GetTeams(ctx *gin.Context) {
	clubID, err := strconv.ParseUint(ctx.Param("clubId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}
	if !c.ensureClubOwner(ctx, uint(clubID)) {
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
	if !c.ensureClubOwner(ctx, uint(clubID)) {
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
		Type         string `json:"type" binding:"required"`
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
		"code":    inviteCode,
		"id":      inv.ID,
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
