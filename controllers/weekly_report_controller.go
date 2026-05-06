package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// WeeklyReportController 周报控制器
type WeeklyReportController struct {
	service *services.WeeklyReportService
	db      *gorm.DB
}

// NewWeeklyReportController 创建周报控制器
func NewWeeklyReportController(service *services.WeeklyReportService, db *gorm.DB) *WeeklyReportController {
	return &WeeklyReportController{service: service, db: db}
}

// Create 创建周报(球员提交)
// POST /api/club/weekly-reports
func (c *WeeklyReportController) Create(ctx *gin.Context) {
	var input models.WeeklyReportSubmit
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	report, err := c.service.Submit(userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report.ToResponse(),
	})
}

// Update 更新周报
// PUT /api/club/weekly-reports/:id
func (c *WeeklyReportController) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的周报ID"})
		return
	}

	var input models.WeeklyReportSubmit
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	report, err := c.service.Update(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report.ToResponse(),
	})
}

// Review 审核周报
// POST /api/club/weekly-reports/:id/review
func (c *WeeklyReportController) Review(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的周报ID"})
		return
	}

	var input models.WeeklyReportReview
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	report, err := c.service.Review(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 发送通知给球员
	notificationHelper := NewNotificationHelper(c.db)
	resp := report.ToResponse()
	coachName := resp.CoachName
	if coachName == "" {
		coachName = "教练"
	}

	// 异步发送通知
	go func() {
		if input.Status == "rejected" {
			// 通知球员周报被退回
			notificationHelper.NotifyWeeklyReportRejected(report.PlayerID, coachName, input.ReviewComment, report.ID)
		} else if input.Status == "approved" {
			// 通知球员周报审核完成 - 使用综合评价中的态度评分作为代表
			notificationHelper.NotifyWeeklyReportApproved(report.PlayerID, coachName, input.CoachAttitudeRating, report.ID)
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// Get 获取周报详情
// GET /api/club/weekly-reports/:id
func (c *WeeklyReportController) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的周报ID"})
		return
	}

	userID := ctx.GetUint("userId")
	report, err := c.service.GetByIDForUser(uint(id), userID)
	if err != nil {
		if errors.Is(err, services.ErrWeeklyReportAccessDenied) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "无权访问该周报"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "周报不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report.ToResponse(),
	})
}

// ListByPlayer 列出球员周报
// GET /api/club/players/:playerId/weekly-reports
func (c *WeeklyReportController) ListByPlayer(ctx *gin.Context) {
	playerIDStr := ctx.Param("playerId")
	playerID, err := strconv.ParseUint(playerIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球员ID"})
		return
	}

	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	userID := ctx.GetUint("userId")
	reports, total, err := c.service.ListByPlayerForUser(uint(playerID), userID, page, pageSize)
	if err != nil {
		if errors.Is(err, services.ErrWeeklyReportAccessDenied) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "无权访问该球员周报"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]models.WeeklyReportResponse, len(reports))
	for i, r := range reports {
		list[i] = r.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ListByTeam 列出球队周报
// GET /api/club/teams/:teamId/weekly-reports
func (c *WeeklyReportController) ListByTeam(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	status := ctx.Query("status")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	reports, total, err := c.service.ListByTeam(uint(teamID), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]models.WeeklyReportResponse, len(reports))
	for i, r := range reports {
		list[i] = r.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ListPending 列出待审核周报
// GET /api/club/weekly-reports/pending
func (c *WeeklyReportController) ListPending(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	reports, total, err := c.service.ListPendingByCoach(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]models.WeeklyReportResponse, len(reports))
	for i, r := range reports {
		list[i] = r.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// GetPendingCount 获取待审核数量
// GET /api/club/teams/:teamId/weekly-reports/pending-count
func (c *WeeklyReportController) GetPendingCount(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	count, err := c.service.GetPendingCount(uint(teamID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count": count,
		},
	})
}

// Delete 删除周报
// DELETE /api/club/weekly-reports/:id
func (c *WeeklyReportController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的周报ID"})
		return
	}

	userID := ctx.GetUint("userId")
	if err := c.service.Delete(uint(id), userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// ==================== 周报周期管理 API ====================

// GetPeriods 获取球队的周报周期列表
// GET /api/club/teams/:teamId/weekly-periods
func (c *WeeklyReportController) GetPeriods(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	status := ctx.Query("status")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	periods, total, err := c.service.GetPeriodsByTeam(uint(teamID), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]models.WeeklyReportPeriodResponse, len(periods))
	for i, p := range periods {
		list[i] = p.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// GetPeriodStats 获取周期统计信息
// GET /api/club/teams/:teamId/weekly-periods/:periodId/stats
func (c *WeeklyReportController) GetPeriodStats(ctx *gin.Context) {
	periodIDStr := ctx.Param("periodId")
	periodID, err := strconv.ParseUint(periodIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的周期ID"})
		return
	}

	stats, err := c.service.GetPeriodStats(uint(periodID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetPeriodPlayers 获取周期内的球员提交情况
// GET /api/club/teams/:teamId/weekly-periods/:periodId/players
func (c *WeeklyReportController) GetPeriodPlayers(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	weekStart := ctx.Query("weekStart")
	if weekStart == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少weekStart参数"})
		return
	}

	weekStartTime, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的weekStart格式"})
		return
	}

	status := ctx.Query("status")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	reports, total, err := c.service.GetPeriodPlayers(uint(teamID), weekStartTime, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]models.WeeklyReportResponse, len(reports))
	for i, r := range reports {
		list[i] = r.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ==================== 一键提醒 ====================

// Remind 一键提醒未提交周报的球员
// POST /api/club/teams/:teamId/weekly-reports/remind
func (c *WeeklyReportController) Remind(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	var input struct {
		WeekStart string `json:"weekStart" binding:"required"` // 周期起始日期
		PlayerIDs []uint `json:"playerIds"`                    // 指定提醒的球员ID列表，为空则提醒所有未提交
		Message   string `json:"message"`                      // 自定义提醒消息
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coachID := ctx.GetUint("userId")

	// 调用服务层发送提醒
	result, err := c.service.RemindPlayers(uint(teamID), coachID, input.WeekStart, input.PlayerIDs, input.Message)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ==================== 导出周报 ====================

// Export 导出周报为CSV格式
// GET /api/club/teams/:teamId/weekly-reports/export
func (c *WeeklyReportController) Export(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的球队ID"})
		return
	}

	// 获取查询参数
	weekStart := ctx.Query("weekStart")
	status := ctx.Query("status")

	// 验证权限（教练或俱乐部管理员）
	userID := ctx.GetUint("userId")
	if !c.service.CanManageTeam(uint(teamID), userID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权导出此球队周报"})
		return
	}

	// 导出数据
	csvData, filename, err := c.service.ExportWeeklyReports(uint(teamID), weekStart, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 写入BOM (UTF-8 Byte Order Mark) 以便Excel正确识别中文
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	ctx.String(http.StatusOK, csvData)
}
