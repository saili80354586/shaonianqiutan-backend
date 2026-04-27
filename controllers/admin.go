package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// AdminController 管理后台控制器
type AdminController struct {
	adminService       *services.AdminService
	videoAnalysisRepo *models.VideoAnalysisRepository
}

func NewAdminController(adminService *services.AdminService, videoAnalysisRepo *models.VideoAnalysisRepository) *AdminController {
	return &AdminController{adminService: adminService, videoAnalysisRepo: videoAnalysisRepo}
}

// DownloadVideoAnalysisDoc 管理员下载视频分析 MD 文档
func (ctrl *AdminController) DownloadVideoAnalysisDoc(c *gin.Context) {
	analysisIDStr := c.Param("id")
	analysisID, err := strconv.ParseUint(analysisIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	docType := c.DefaultQuery("type", "rating") // rating | player-info
	if docType != "rating" && docType != "player-info" {
		utils.Error(c, http.StatusBadRequest, "无效的文档类型")
		return
	}

	analysis, err := ctrl.videoAnalysisRepo.FindByID(uint(analysisID))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "视频分析记录不存在")
		return
	}

	var filePath string
	var fileName string
	if docType == "rating" {
		filePath = analysis.RatingReportMD
		fileName = fmt.Sprintf("评分报告_VA%d.md", analysis.ID)
	} else {
		filePath = analysis.PlayerInfoMD
		fileName = fmt.Sprintf("球员基础信息_%s.md", analysis.PlayerName)
	}

	if filePath == "" {
		utils.Error(c, http.StatusNotFound, "文档尚未生成，请先确认分析")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.File(filePath)
}

// GetDashboardStats 获取数据看板统计数据
func (ctrl *AdminController) GetDashboardStats(c *gin.Context) {
	stats, err := ctrl.adminService.GetDashboardStats()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取统计数据失败")
		return
	}
	utils.Success(c, "", stats)
}

// GetGrowthData 获取增长数据
func (ctrl *AdminController) GetGrowthData(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	data, err := ctrl.adminService.GetGrowthData(days)
	if err != nil {
		// 返回模拟数据
		var mockData []gin.H
		for i := days - 1; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			mockData = append(mockData, gin.H{
				"date":    date.Format("01-02"),
				"users":   10 + i*2,
				"orders":  5 + i,
				"revenue": 1000 + i*100,
			})
		}
		utils.Success(c, "", mockData)
		return
	}

	utils.Success(c, "", data)
}

// GetUserList 获取用户列表
func (ctrl *AdminController) GetUserList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	users, total, err := ctrl.adminService.GetUserList(page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateUserStatus 更新用户状态
func (ctrl *AdminController) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.adminService.UpdateUserStatus(uint(id), req.Status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, "状态更新成功", nil)
}

// GetAllOrders 获取所有订单列表
func (ctrl *AdminController) GetAllOrders(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := c.DefaultQuery("status", "")

	orders, total, err := ctrl.adminService.GetAllOrders(page, pageSize, status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     orders,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetPendingReports 获取待审核报告列表
func (ctrl *AdminController) GetPendingReports(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	reports, total, err := ctrl.adminService.GetPendingReports(page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     reports,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ReviewReportRequest 审核报告请求
type ReviewReportRequest struct {
	Status models.ReportStatus `json:"status" binding:"required,oneof=processing completed failed"`
	Remark string              `json:"remark"`
}

// ReviewReport 审核报告
func (ctrl *AdminController) ReviewReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	var req ReviewReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.adminService.ReviewReport(uint(id), req.Status, req.Remark)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "审核失败: "+err.Error())
		return
	}

	utils.Success(c, "审核完成", nil)
}

// DownloadReportDoc 管理员下载报告 MD 文档或 AI Word 报告
func (ctrl *AdminController) DownloadReportDoc(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	docType := c.DefaultQuery("type", "rating") // rating | player-info | report
	if docType != "rating" && docType != "player-info" && docType != "report" {
		utils.Error(c, http.StatusBadRequest, "无效的文档类型")
		return
	}

	report, err := ctrl.adminService.GetReportByID(uint(reportID))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	var filePath string
	var fileName string
	if docType == "rating" {
		filePath = report.RatingReportMD
		fileName = fmt.Sprintf("评分报告_%d.md", report.OrderID)
	} else if docType == "player-info" {
		filePath = report.PlayerInfoMD
		fileName = fmt.Sprintf("球员基础信息_%s.md", report.PlayerName)
	} else {
		// AI Word 报告
		filePath = "./uploads/reports/" + strings.TrimPrefix(report.AIReportURL, "/uploads/reports/")
		fileName = filepath.Base(filePath)
		if report.AIReportURL == "" {
			utils.Error(c, http.StatusNotFound, "AI 报告尚未生成")
			return
		}
	}

	if filePath == "" {
		utils.Error(c, http.StatusNotFound, "文档不存在")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	// 根据文件类型设置 Content-Type
	if strings.HasSuffix(filePath, ".docx") {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	} else if strings.HasSuffix(filePath, ".md") {
		c.Header("Content-Type", "text/markdown; charset=utf-8")
	} else {
		c.Header("Content-Type", "application/octet-stream")
	}
	c.File(filePath)
}

// UploadAIReport 上传 AI Word 报告
func (ctrl *AdminController) UploadAIReport(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	// 获取报告
	report, err := ctrl.adminService.GetReportByID(uint(reportID))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	// 上传文件（复用通用上传，允许 .docx）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*1024*1024+1024*1024)
	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			utils.Error(c, http.StatusBadRequest, "文件大小超过限制（最大 50MB）")
			return
		}
		utils.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".doc" && ext != ".docx" {
		utils.Error(c, http.StatusBadRequest, "仅支持 Word 文档格式（.doc / .docx）")
		return
	}

	// 保存文件
	uploadDir := "./uploads/reports"
	_ = os.MkdirAll(uploadDir, 0755)
	timestamp := time.Now().UnixNano()
	newFilename := fmt.Sprintf("ai_report_%d_%d%s", report.ID, timestamp, ext)
	savePath := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	// 更新报告 URL
	fileURL := fmt.Sprintf("/uploads/reports/%s", newFilename)
	if err := ctrl.adminService.UpdateReportAIURL(uint(reportID), fileURL, ""); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新报告URL失败")
		return
	}

	_ = report // suppress unused warning
	utils.Success(c, "AI 报告上传成功", gin.H{"url": fileURL})
}

// UploadAIVideo 上传 AI 视频分析
func (ctrl *AdminController) UploadAIVideo(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	report, err := ctrl.adminService.GetReportByID(uint(reportID))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 500*1024*1024+1024*1024)
	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			utils.Error(c, http.StatusBadRequest, "文件大小超过限制（最大 500MB）")
			return
		}
		utils.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedVideo := []string{".mp4", ".mov", ".avi", ".mkv", ".webm"}
	validExt := false
	for _, e := range allowedVideo {
		if ext == e {
			validExt = true
			break
		}
	}
	if !validExt {
		utils.Error(c, http.StatusBadRequest, "仅支持视频格式（.mp4 / .mov / .avi / .mkv / .webm）")
		return
	}

	uploadDir := "./uploads/reports"
	_ = os.MkdirAll(uploadDir, 0755)
	timestamp := time.Now().UnixNano()
	newFilename := fmt.Sprintf("ai_video_%d_%d%s", report.ID, timestamp, ext)
	savePath := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	fileURL := fmt.Sprintf("/uploads/reports/%s", newFilename)
	if err := ctrl.adminService.UpdateReportAIURL(uint(reportID), "", fileURL); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新视频URL失败")
		return
	}

	_ = report
	utils.Success(c, "AI 视频分析上传成功", gin.H{"url": fileURL})
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 管理员登录
func (ctrl *AdminController) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, admin, err := ctrl.adminService.AdminLogin(req.Username, req.Password)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	utils.Success(c, "登录成功", gin.H{
		"token": token,
		"admin": admin,
	})
}

// GetStatistics 获取核心数据统计
func (ctrl *AdminController) GetStatistics(c *gin.Context) {
	stats, err := ctrl.adminService.GetDashboardStats()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取统计数据失败")
		return
	}
	utils.Success(c, "", stats)
}

// DeleteUser 删除用户
func (ctrl *AdminController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	err = ctrl.adminService.DeleteUser(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, "删除成功", nil)
}

// AssignOrder 管理员派单给分析师
func (ctrl *AdminController) AssignOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	var req services.AssignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 通过 adminService 获取 orderService（需要确认 adminService 是否有该方法）
	// 由于 AdminController 只有 adminService，这里需要调用 adminService 的 AssignOrder
	order, err := ctrl.adminService.AssignOrder(uint(id), req.AnalystID)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "派单失败: "+err.Error())
		return
	}

	utils.Success(c, "派单成功", gin.H{"order": order})
}

// CancelOrder 取消订单
func (ctrl *AdminController) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	err = ctrl.adminService.CancelOrder(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "取消订单失败: "+err.Error())
		return
	}

	utils.Success(c, "订单已取消", nil)
}

// GetAnalystList 获取分析师列表
func (ctrl *AdminController) GetAnalystList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := c.DefaultQuery("status", "")

	analysts, total, err := ctrl.adminService.GetAnalystList(page, pageSize, status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取分析师列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     analysts,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetAvailableAnalysts 获取可派单的分析师列表（含工作负载）
func (ctrl *AdminController) GetAvailableAnalysts(c *gin.Context) {
	analysts, err := ctrl.adminService.GetAvailableAnalysts()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取分析师列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":  analysts,
		"total": len(analysts),
	})
}

// AuditAnalystRequest 审核分析师请求
type AuditAnalystRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Remark string `json:"remark"`
}

// AuditAnalyst 审核分析师
func (ctrl *AdminController) AuditAnalyst(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	var req AuditAnalystRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.adminService.AuditAnalyst(uint(id), req.Status, req.Remark)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "审核失败: "+err.Error())
		return
	}

	utils.Success(c, "审核完成", nil)
}

// UpdateAnalystStatusRequest 更新分析师状态请求
type UpdateAnalystStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateAnalystStatus 更新分析师状态
func (ctrl *AdminController) UpdateAnalystStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	var req UpdateAnalystStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.adminService.UpdateAnalystStatus(uint(id), req.Status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, "状态更新成功", nil)
}

// GetFunnelData 获取漏斗数据
func (ctrl *AdminController) GetFunnelData(c *gin.Context) {
	data, err := ctrl.adminService.GetFunnelData()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取漏斗数据失败")
		return
	}
	utils.Success(c, "", data)
}

// GetRetentionData 获取留存数据
func (ctrl *AdminController) GetRetentionData(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := ctrl.adminService.GetRetentionData(days)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取留存数据失败")
		return
	}
	utils.Success(c, "", data)
}

// GetTopData 获取排行榜数据
func (ctrl *AdminController) GetTopData(c *gin.Context) {
	data, err := ctrl.adminService.GetTopData()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取排行榜数据失败")
		return
	}
	utils.Success(c, "", data)
}

// GetOrderStats 获取订单统计
func (ctrl *AdminController) GetOrderStats(c *gin.Context) {
	stats, err := ctrl.adminService.GetOrderStats()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单统计失败")
		return
	}
	utils.Success(c, "", stats)
}

// GetRevenueTrend 获取收入趋势
func (ctrl *AdminController) GetRevenueTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := ctrl.adminService.GetRevenueTrend(days)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取收入趋势失败")
		return
	}
	utils.Success(c, "", data)
}

// GetAnalystIncomeStats 获取分析师收益统计
func (ctrl *AdminController) GetAnalystIncomeStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}
	stats, err := ctrl.adminService.GetAnalystIncomeStats(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取收益统计失败")
		return
	}
	utils.Success(c, "", stats)
}

// GetSettlementList 获取待结算列表
func (ctrl *AdminController) GetSettlementList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	orders, total, err := ctrl.adminService.GetSettlementList(pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取结算列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": orders, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// ProcessSettlementRequest 处理结算请求
type ProcessSettlementRequest struct {
	OrderIDs []uint `json:"order_ids" binding:"required"`
}

// ProcessSettlement 处理结算
func (ctrl *AdminController) ProcessSettlement(c *gin.Context) {
	adminID := c.GetUint("userId")
	var req ProcessSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.adminService.ProcessSettlement(req.OrderIDs, adminID); err != nil {
		utils.Error(c, http.StatusInternalServerError, "结算处理失败: "+err.Error())
		return
	}
	utils.Success(c, "结算处理成功", nil)
}

// GetContentReports 获取举报列表
func (ctrl *AdminController) GetContentReports(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	status := c.DefaultQuery("status", "")
	reports, total, err := ctrl.adminService.GetContentReports(pagination.Page, pagination.PageSize, status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取举报列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": reports, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// GetContentReportDetail 获取举报详情
func (ctrl *AdminController) GetContentReportDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的举报ID")
		return
	}
	report, err := ctrl.adminService.GetContentReportDetail(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取举报详情失败")
		return
	}
	utils.Success(c, "", report)
}

// HandleContentReportRequest 处理举报请求
type HandleContentReportRequest struct {
	Status string `json:"status" binding:"required,oneof=resolved rejected"`
	Result string `json:"result"`
}

// HandleContentReport 处理举报
func (ctrl *AdminController) HandleContentReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的举报ID")
		return
	}
	adminID := c.GetUint("userId")
	adminName := c.GetString("userName")
	var req HandleContentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	status := models.ContentReportStatus(req.Status)
	if err := ctrl.adminService.HandleContentReport(uint(id), status, adminID, adminName, req.Result); err != nil {
		utils.Error(c, http.StatusInternalServerError, "处理失败: "+err.Error())
		return
	}
	utils.Success(c, "处理成功", nil)
}

// GetSensitiveWords 获取敏感词列表
func (ctrl *AdminController) GetSensitiveWords(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	category := c.DefaultQuery("category", "")
	enabledStr := c.DefaultQuery("enabled", "")
	var enabled *bool
	if enabledStr != "" {
		e := enabledStr == "true"
		enabled = &e
	}
	words, total, err := ctrl.adminService.GetSensitiveWords(pagination.Page, pagination.PageSize, category, enabled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取敏感词列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": words, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// CreateSensitiveWordRequest 创建敏感词请求
type CreateSensitiveWordRequest struct {
	Word     string `json:"word" binding:"required"`
	Category string `json:"category"`
	Level    int    `json:"level"`
}

// CreateSensitiveWord 创建敏感词
func (ctrl *AdminController) CreateSensitiveWord(c *gin.Context) {
	var req CreateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	word := &models.SensitiveWord{Word: req.Word, Category: req.Category, Level: req.Level, Enabled: true}
	if err := ctrl.adminService.CreateSensitiveWord(word); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	utils.Success(c, "创建成功", word)
}

// UpdateSensitiveWordRequest 更新敏感词请求
type UpdateSensitiveWordRequest struct {
	Word     string `json:"word"`
	Category string `json:"category"`
	Level    int    `json:"level"`
	Enabled  *bool  `json:"enabled"`
}

// UpdateSensitiveWord 更新敏感词
func (ctrl *AdminController) UpdateSensitiveWord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的敏感词ID")
		return
	}
	var req UpdateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Word != "" {
		updates["word"] = req.Word
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if err := ctrl.adminService.UpdateSensitiveWord(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	utils.Success(c, "更新成功", nil)
}

// DeleteSensitiveWord 删除敏感词
func (ctrl *AdminController) DeleteSensitiveWord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的敏感词ID")
		return
	}
	if err := ctrl.adminService.DeleteSensitiveWord(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, "删除成功", nil)
}

// GetPlatformAnnouncements 获取平台公告列表
func (ctrl *AdminController) GetPlatformAnnouncements(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	annType := c.DefaultQuery("type", "")
	pinnedStr := c.DefaultQuery("pinned", "")
	var pinned *bool
	if pinnedStr != "" {
		p := pinnedStr == "true"
		pinned = &p
	}
	announcements, total, err := ctrl.adminService.GetPlatformAnnouncements(pagination.Page, pagination.PageSize, annType, pinned)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取公告列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": announcements, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// CreatePlatformAnnouncementRequest 创建平台公告请求
type CreatePlatformAnnouncementRequest struct {
	Title    string     `json:"title" binding:"required"`
	Content  string     `json:"content" binding:"required"`
	Type     string     `json:"type"`
	IsPinned bool       `json:"is_pinned"`
	IsPublic bool       `json:"is_public"`
	StartAt  *time.Time `json:"start_at"`
	EndAt    *time.Time `json:"end_at"`
}

// CreatePlatformAnnouncement 创建平台公告
func (ctrl *AdminController) CreatePlatformAnnouncement(c *gin.Context) {
	adminID := c.GetUint("userId")
	adminName := c.GetString("userName")
	var req CreatePlatformAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ann := &models.PlatformAnnouncement{
		Title: req.Title, Content: req.Content, Type: req.Type,
		IsPinned: req.IsPinned, IsPublic: req.IsPublic,
		StartAt: req.StartAt, EndAt: req.EndAt,
		CreatedBy: adminID, AuthorName: adminName,
	}
	if err := ctrl.adminService.CreatePlatformAnnouncement(ann); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	utils.Success(c, "创建成功", ann)
}

// UpdatePlatformAnnouncementRequest 更新平台公告请求
type UpdatePlatformAnnouncementRequest struct {
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Type     string     `json:"type"`
	IsPinned *bool      `json:"is_pinned"`
	IsPublic *bool      `json:"is_public"`
	StartAt  *time.Time `json:"start_at"`
	EndAt    *time.Time `json:"end_at"`
}

// UpdatePlatformAnnouncement 更新平台公告
func (ctrl *AdminController) UpdatePlatformAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的公告ID")
		return
	}
	var req UpdatePlatformAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.StartAt != nil {
		updates["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		updates["end_at"] = req.EndAt
	}
	if err := ctrl.adminService.UpdatePlatformAnnouncement(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	utils.Success(c, "更新成功", nil)
}

// DeletePlatformAnnouncement 删除平台公告
func (ctrl *AdminController) DeletePlatformAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的公告ID")
		return
	}
	if err := ctrl.adminService.DeletePlatformAnnouncement(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, "删除成功", nil)
}

// GetBanners 获取轮播图列表
func (ctrl *AdminController) GetBanners(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	position := c.DefaultQuery("position", "")
	enabledStr := c.DefaultQuery("enabled", "")
	var enabled *bool
	if enabledStr != "" {
		e := enabledStr == "true"
		enabled = &e
	}
	banners, total, err := ctrl.adminService.GetBanners(pagination.Page, pagination.PageSize, position, enabled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取轮播图列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": banners, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// CreateBannerRequest 创建轮播图请求
type CreateBannerRequest struct {
	Title     string     `json:"title"`
	ImageURL  string     `json:"image_url" binding:"required"`
	LinkURL   string     `json:"link_url"`
	Position  string     `json:"position"`
	SortOrder int        `json:"sort_order"`
	Enabled   bool       `json:"enabled"`
	StartAt   *time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
}

// CreateBanner 创建轮播图
func (ctrl *AdminController) CreateBanner(c *gin.Context) {
	adminID := c.GetUint("userId")
	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	banner := &models.Banner{
		Title: req.Title, ImageURL: req.ImageURL, LinkURL: req.LinkURL,
		Position: req.Position, SortOrder: req.SortOrder, Enabled: req.Enabled,
		StartAt: req.StartAt, EndAt: req.EndAt, CreatedBy: adminID,
	}
	if err := ctrl.adminService.CreateBanner(banner); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	utils.Success(c, "创建成功", banner)
}

// UpdateBannerRequest 更新轮播图请求
type UpdateBannerRequest struct {
	Title     string     `json:"title"`
	ImageURL  string     `json:"image_url"`
	LinkURL   string     `json:"link_url"`
	Position  string     `json:"position"`
	SortOrder int        `json:"sort_order"`
	Enabled   *bool      `json:"enabled"`
	StartAt   *time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
}

// UpdateBanner 更新轮播图
func (ctrl *AdminController) UpdateBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的轮播图ID")
		return
	}
	var req UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.LinkURL != "" {
		updates["link_url"] = req.LinkURL
	}
	if req.Position != "" {
		updates["position"] = req.Position
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.StartAt != nil {
		updates["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		updates["end_at"] = req.EndAt
	}
	if err := ctrl.adminService.UpdateBanner(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	utils.Success(c, "更新成功", nil)
}

// DeleteBanner 删除轮播图
func (ctrl *AdminController) DeleteBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的轮播图ID")
		return
	}
	if err := ctrl.adminService.DeleteBanner(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, "删除成功", nil)
}

// GetFAQs 获取FAQ列表
func (ctrl *AdminController) GetFAQs(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	category := c.DefaultQuery("category", "")
	enabledStr := c.DefaultQuery("enabled", "")
	var enabled *bool
	if enabledStr != "" {
		e := enabledStr == "true"
		enabled = &e
	}
	faqs, total, err := ctrl.adminService.GetFAQs(pagination.Page, pagination.PageSize, category, enabled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取FAQ列表失败")
		return
	}
	utils.Success(c, "", gin.H{"list": faqs, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// CreateFAQRequest 创建FAQ请求
type CreateFAQRequest struct {
	Question  string `json:"question" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
	Category  string `json:"category"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
}

// CreateFAQ 创建FAQ
func (ctrl *AdminController) CreateFAQ(c *gin.Context) {
	var req CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	faq := &models.FAQ{
		Question: req.Question, Answer: req.Answer,
		Category: req.Category, SortOrder: req.SortOrder, Enabled: req.Enabled,
	}
	if err := ctrl.adminService.CreateFAQ(faq); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	utils.Success(c, "创建成功", faq)
}

// UpdateFAQRequest 更新FAQ请求
type UpdateFAQRequest struct {
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Category  string `json:"category"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}

// UpdateFAQ 更新FAQ
func (ctrl *AdminController) UpdateFAQ(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的FAQ ID")
		return
	}
	var req UpdateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Question != "" {
		updates["question"] = req.Question
	}
	if req.Answer != "" {
		updates["answer"] = req.Answer
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if err := ctrl.adminService.UpdateFAQ(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	utils.Success(c, "更新成功", nil)
}

// DeleteFAQ 删除FAQ
func (ctrl *AdminController) DeleteFAQ(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的FAQ ID")
		return
	}
	if err := ctrl.adminService.DeleteFAQ(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, "删除成功", nil)
}

// GetLoginLogs 获取登录日志
func (ctrl *AdminController) GetLoginLogs(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	userIDStr := c.DefaultQuery("user_id", "0")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	status := c.DefaultQuery("status", "")
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	logs, total, err := ctrl.adminService.GetLoginLogs(pagination.Page, pagination.PageSize, uint(userID), status, startDate, endDate)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取登录日志失败")
		return
	}
	utils.Success(c, "", gin.H{"list": logs, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// GetLoginLogStats 获取登录日志统计
func (ctrl *AdminController) GetLoginLogStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	stats, err := ctrl.adminService.GetLoginLogStats(days)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取登录日志统计失败")
		return
	}
	utils.Success(c, "", stats)
}
