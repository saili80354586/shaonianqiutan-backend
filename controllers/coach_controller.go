package controllers

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// CoachController 教练控制器
type CoachController struct {
	coachService *services.CoachService
}

// NewCoachController 创建教练控制器
func NewCoachController(coachService *services.CoachService) *CoachController {
	return &CoachController{coachService: coachService}
}

// GetCoachProfile 获取教练资料
func (c *CoachController) GetCoachProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取教练资料失败")
		return
	}

	// 获取用户的 position（存储在 users 表，用于球探地图筛选）
	var user models.User
	if err := config.GetDB().Select("id, name, avatar, nickname, position, province, city").Where("id = ?", userID).First(&user).Error; err == nil {
		// 返回时将 user.position 附加到响应中
		data := gin.H{
			"coach":    coach,
			"position": user.Position,
			"user": gin.H{
				"id":       user.ID,
				"name":     user.Name,
				"nickname": user.Nickname,
				"avatar":   user.Avatar,
				"position": user.Position,
				"province": user.Province,
				"city":     user.City,
			},
		}
		utils.SuccessResponse(ctx, data)
		return
	}

	utils.SuccessResponse(ctx, gin.H{"coach": coach})
}

// GetCoachPublicProfile 获取教练公开主页
func (c *CoachController) GetCoachPublicProfile(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(ctx, 400, "无效的教练ID")
		return
	}

	profile, err := c.coachService.GetCoachPublicProfile(uint(id))
	if err != nil {
		utils.ServerError(ctx, "获取教练主页失败")
		return
	}
	if profile == nil {
		utils.Error(ctx, 404, "教练不存在")
		return
	}

	utils.Success(ctx, "", profile)
}

// GetCoachPublicProfileByUser 通过 user_id 获取教练公开主页
func (c *CoachController) GetCoachPublicProfileByUser(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		utils.Error(ctx, 400, "缺少 user_id 参数")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.Error(ctx, 400, "无效的 user_id")
		return
	}

	// 先通过 user_id 获取 coach_id
	coach, err := c.coachService.GetCoachByUserID(uint(userID))
	if err != nil || coach == nil {
		utils.Error(ctx, 404, "该用户不是教练或教练不存在")
		return
	}

	profile, err := c.coachService.GetCoachPublicProfile(coach.ID)
	if err != nil {
		utils.ServerError(ctx, "获取教练主页失败")
		return
	}

	utils.Success(ctx, "", profile)
}

// UpdateCoachProfile 更新教练资料
func (c *CoachController) UpdateCoachProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		LicenseType   string   `json:"licenseType"`
		LicenseNumber string   `json:"licenseNumber"`
		Specialties   []string `json:"specialties"`
		Style         []string `json:"style"`
		AgeGroups     []string `json:"ageGroups"`
		Bio           string   `json:"bio"`
		City          string   `json:"city"`
		CoachingYears int      `json:"coachingYears"`
		CurrentClub   string   `json:"currentClub"`
		Position      string   `json:"position"` // 执教位置（同步到 users.position，球探地图筛选）
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	specialtiesJSON, _ := json.Marshal(req.Specialties)
	styleJSON, _ := json.Marshal(req.Style)
	ageGroupsJSON, _ := json.Marshal(req.AgeGroups)
	coach, err := c.coachService.UpdateCoachProfile(
		userID,
		req.LicenseType,
		req.LicenseNumber,
		string(specialtiesJSON),
		string(styleJSON),
		string(ageGroupsJSON),
		req.Bio,
		req.City,
		req.CurrentClub,
		req.CoachingYears,
		req.Position,
	)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":      coach.ID,
		"updated": true,
	}, "资料更新成功")
}

// GetDashboard 获取工作台数据
func (c *CoachController) GetDashboard(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.Error(ctx, 500, "获取数据失败")
		return
	}

	stats := c.coachService.GetDashboardStats(coach.ID, userID)
	activities, _ := c.coachService.GetRecentActivities(coach.ID, 10)

	utils.SuccessResponse(ctx, gin.H{
		"stats":            stats,
		"recentActivities": activities,
	})
}

// GetFollowedPlayers 获取关注的球员列表
func (c *CoachController) GetFollowedPlayers(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	keyword := ctx.Query("keyword")

	players, total, err := c.coachService.GetFollowedPlayers(coach.ID, page, pageSize, keyword)
	if err != nil {
		utils.ServerError(ctx, "获取列表失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     players,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// FollowPlayer 关注球员
func (c *CoachController) FollowPlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		PlayerID uint `json:"playerId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.FollowPlayer(coach.ID, req.PlayerID)
	if err != nil {
		utils.ServerError(ctx, "关注失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"followed": true,
	}, "关注成功")
}

// UnfollowPlayer 取消关注球员
func (c *CoachController) UnfollowPlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.UnfollowPlayer(coach.ID, uint(playerID))
	if err != nil {
		utils.ServerError(ctx, "取消关注失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "已取消关注")
}

// UpdateFollowNotes 更新关注备注
func (c *CoachController) UpdateFollowNotes(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	var req struct {
		Notes     string `json:"notes"`
		IsStarred bool   `json:"isStarred"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.UpdateFollowNotes(coach.ID, uint(playerID), req.Notes, req.IsStarred)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// GetTrainingNotes 获取训练笔记列表
func (c *CoachController) GetTrainingNotes(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	playerIDStr := ctx.Query("playerId")
	category := ctx.Query("category")

	var playerID *uint
	if playerIDStr != "" {
		if pid, err := strconv.ParseUint(playerIDStr, 10, 32); err == nil {
			pidVal := uint(pid)
			playerID = &pidVal
		}
	}

	notes, total, err := c.coachService.GetTrainingNotes(coach.ID, page, pageSize, playerID, category)
	if err != nil {
		utils.ServerError(ctx, "获取列表失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     notes,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateTrainingNote 创建训练笔记
func (c *CoachController) CreateTrainingNote(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		PlayerID uint     `json:"playerId" binding:"required"`
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Rating   int      `json:"rating"`
		IsPublic bool     `json:"isPublic"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	note, err := c.coachService.CreateTrainingNote(
		coach.ID, req.PlayerID, req.Title, req.Content, req.Category, req.Tags, req.Rating, req.IsPublic,
	)
	if err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id": note.ID,
	}, "创建成功")
}

// UpdateTrainingNote 更新训练笔记
func (c *CoachController) UpdateTrainingNote(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	noteID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的笔记ID")
		return
	}

	var req struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Rating   int      `json:"rating"`
		IsPublic bool     `json:"isPublic"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.UpdateTrainingNote(
		coach.ID, uint(noteID), req.Title, req.Content, req.Category, req.Tags, req.Rating, req.IsPublic,
	)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// DeleteTrainingNote 删除训练笔记
func (c *CoachController) DeleteTrainingNote(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	noteID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的笔记ID")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.DeleteTrainingNote(coach.ID, uint(noteID))
	if err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// GetPlayerProgress 获取球员进度
func (c *CoachController) GetPlayerProgress(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	progress, err := c.coachService.GetPlayerProgress(coach.ID, uint(playerID))
	if err != nil {
		if errors.Is(err, services.ErrCoachPlayerAccessDenied) {
			utils.ForbiddenError(ctx, "无权访问该球员")
			return
		}
		utils.ServerError(ctx, "获取进度失败")
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"progress": progress,
	})
}
