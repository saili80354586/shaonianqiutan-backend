package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// ScoutController 球探控制器
type ScoutController struct {
	scoutService *services.ScoutService
}

// NewScoutController 创建球探控制器
func NewScoutController(scoutService *services.ScoutService) *ScoutController {
	return &ScoutController{scoutService: scoutService}
}

// GetScoutProfile 获取球探资料
func (c *ScoutController) GetScoutProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	scout, err := c.scoutService.GetScoutProfile(userID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取球探资料失败")
		return
	}

	utils.Success(ctx, "查询成功", scout)
}

// UpdateScoutProfile 更新球探资料
func (c *ScoutController) UpdateScoutProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var data map[string]interface{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	scout, err := c.scoutService.UpdateScoutProfile(userID, data)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "更新失败")
		return
	}

	utils.Success(ctx, "更新成功", scout)
}

// GetScoutDashboard 获取球探工作台
func (c *ScoutController) GetScoutDashboard(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	data, err := c.scoutService.GetScoutDashboard(userID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取数据失败")
		return
	}

	utils.Success(ctx, "查询成功", data)
}

// GetFollowedPlayers 获取关注的球员列表
func (c *ScoutController) GetFollowedPlayers(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	follows, total, err := c.scoutService.GetFollowedPlayers(userID, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  follows,
		"total": total,
	})
}

// FollowPlayer 关注球员
func (c *ScoutController) FollowPlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var data map[string]interface{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	playerIDVal, ok := data["player_id"]
	if !ok {
		utils.Error(ctx, http.StatusBadRequest, "缺少player_id")
		return
	}

	playerID, ok := playerIDVal.(float64)
	if !ok {
		utils.Error(ctx, http.StatusBadRequest, "player_id格式错误")
		return
	}

	follow, err := c.scoutService.FollowPlayer(userID, uint(playerID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "关注失败")
		return
	}

	utils.Success(ctx, "关注成功", follow)
}

// UnfollowPlayer 取消关注球员
func (c *ScoutController) UnfollowPlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	if err := c.scoutService.UnfollowPlayer(userID, uint(playerID)); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "取消关注失败")
		return
	}

	utils.Success(ctx, "取消关注成功", nil)
}

// GetScoutReports 获取球探报告列表
func (c *ScoutController) GetScoutReports(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	status := ctx.Query("status")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	reports, total, err := c.scoutService.GetScoutReports(userID, status, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  reports,
		"total": total,
	})
}

// GetScoutReport 获取单个球探报告
func (c *ScoutController) GetScoutReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	reportID, err := strconv.ParseUint(ctx.Param("reportId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	report, err := c.scoutService.GetScoutReport(userID, uint(reportID))
	if err != nil {
		utils.Error(ctx, http.StatusNotFound, "报告不存在")
		return
	}

	utils.Success(ctx, "查询成功", report)
}

// CreateScoutReport 创建球探报告
func (c *ScoutController) CreateScoutReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var data map[string]interface{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	report, err := c.scoutService.CreateScoutReport(userID, data)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "创建失败")
		return
	}

	utils.Success(ctx, "创建成功", report)
}

// UpdateScoutReport 更新球探报告
func (c *ScoutController) UpdateScoutReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	reportID, err := strconv.ParseUint(ctx.Param("reportId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	var data map[string]interface{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	report, err := c.scoutService.UpdateScoutReport(userID, uint(reportID), data)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "更新失败")
		return
	}

	utils.Success(ctx, "更新成功", report)
}

// PublishScoutReport 发布球探报告
func (c *ScoutController) PublishScoutReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	reportID, err := strconv.ParseUint(ctx.Param("reportId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	report, err := c.scoutService.PublishScoutReport(userID, uint(reportID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "发布失败")
		return
	}

	utils.Success(ctx, "发布成功", report)
}

// DeleteScoutReport 删除球探报告
func (c *ScoutController) DeleteScoutReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	reportID, err := strconv.ParseUint(ctx.Param("reportId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	if err := c.scoutService.DeleteScoutReport(userID, uint(reportID)); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	utils.Success(ctx, "删除成功", nil)
}

// GetScoutTasks 获取球探任务列表
func (c *ScoutController) GetScoutTasks(ctx *gin.Context) {
	status := ctx.Query("status")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	tasks, total, err := c.scoutService.GetScoutTasks(status, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  tasks,
		"total": total,
	})
}

// AcceptScoutTask 接取球探任务
func (c *ScoutController) AcceptScoutTask(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	taskID, err := strconv.ParseUint(ctx.Param("taskId"), 10, 64)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	task, err := c.scoutService.AcceptScoutTask(userID, uint(taskID))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "接取失败")
		return
	}

	utils.Success(ctx, "接取成功", task)
}

// SearchPlayers 搜索球员
func (c *ScoutController) SearchPlayers(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	if userID == 0 {
		utils.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	keyword := ctx.Query("keyword")
	position := ctx.Query("position")
	region := ctx.Query("region")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	players, total, err := c.scoutService.SearchPlayers(keyword, position, region, page, pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "搜索失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  players,
		"total": total,
	})
}

// GetScoutPublicProfileByID 通过 scout_id 获取球探公开主页
func (c *ScoutController) GetScoutPublicProfileByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "无效的球探ID")
		return
	}

	profile, err := c.scoutService.GetScoutPublicProfile(uint(id))
	if err != nil {
		utils.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(ctx, "", profile)
}

// GetScoutPublicProfile 通过 user_id 获取球探公开主页
func (c *ScoutController) GetScoutPublicProfile(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		utils.Error(ctx, http.StatusBadRequest, "缺少 user_id 参数")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "无效的 user_id")
		return
	}

	// 先通过 user_id 获取 scout
	scout, err := c.scoutService.GetOrCreateScout(uint(userID))
	if err != nil || scout == nil {
		utils.Error(ctx, http.StatusNotFound, "该用户不是球探或球探不存在")
		return
	}

	profile, err := c.scoutService.GetScoutPublicProfile(scout.ID)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取球探主页失败")
		return
	}

	utils.Success(ctx, "", profile)
}
