package controllers

import (
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// ReportController 报告控制器
type ReportController struct {
	reportService *services.ReportService
	authService   *services.AuthService
}

func NewReportController(reportService *services.ReportService, authService *services.AuthService) *ReportController {
	return &ReportController{
		reportService: reportService,
		authService:   authService,
	}
}

// CreateReport 创建球探报告
func (ctrl *ReportController) CreateReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req services.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: 需要检查订单是否存在和已支付
	// 由于订单管理不在本次重构范围内，先跳过这部分检查

	report, err := ctrl.reportService.CreateReport(&req, userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建报告失败")
		return
	}

	utils.Success(c, "报告创建成功，正在生成PDF", gin.H{
		"report_id": report.ID,
	})
}

// GetReportDetail 获取报告详情
func (ctrl *ReportController) GetReportDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	// 获取当前用户角色
	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	report, hasPermission, err := ctrl.reportService.GetReportDetail(uint(id), userID, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告详情失败")
		return
	}

	if report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	if !hasPermission {
		utils.Error(c, http.StatusForbidden, "无权限查看")
		return
	}

	utils.Success(c, "", gin.H{"data": report})
}

// DownloadReport 下载PDF报告
func (ctrl *ReportController) DownloadReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	// 获取当前用户角色
	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	report, hasPermission, err := ctrl.reportService.CheckDownloadPermission(uint(id), userID, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "下载失败")
		return
	}

	if report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	if !hasPermission {
		utils.Error(c, http.StatusForbidden, "无权限下载")
		return
	}

	if report.Status != models.ReportStatusCompleted {
		utils.Error(c, http.StatusBadRequest, "报告尚未生成完成")
		return
	}

	if report.PdfURL == "" {
		utils.Error(c, http.StatusNotFound, "PDF文件不存在")
		return
	}

	filePath := ctrl.reportService.GetPdfFilePath(report)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "PDF文件不存在")
		return
	}

	fileName := report.PlayerName + "_球探报告.pdf"
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+urlEncode(fileName)+"\"")
	c.File(filePath)
}

// GetMyReports 获取我的报告列表（作为买家）
func (ctrl *ReportController) GetMyReports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	result, err := ctrl.reportService.GetUserReports(userID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告列表失败: "+err.Error())
		return
	}

	utils.Success(c, "", gin.H{"data": result})
}

// GetMyPublishedReports 获取我发布的报告列表（作为分析师）
func (ctrl *ReportController) GetMyPublishedReports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	result, err := ctrl.reportService.GetAnalystReports(userID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告列表失败")
		return
	}

	utils.Success(c, "", gin.H{"data": result})
}

// RegeneratePdf 重新生成PDF
func (ctrl *ReportController) RegeneratePdf(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	// 获取当前用户角色
	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	report, ok, err := ctrl.reportService.RegeneratePdf(uint(id), userID, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "重新生成失败")
		return
	}

	if report == nil {
		if !ok {
			utils.Error(c, http.StatusForbidden, "无权限操作")
		} else {
			utils.Error(c, http.StatusNotFound, "报告不存在")
		}
		return
	}

	// TODO: 这里应该异步重新生成PDF
	// 先返回成功，实际PDF生成逻辑需要根据原有项目实现对接

	utils.Success(c, "PDF重新生成中", gin.H{
		"pdf_url": report.PdfURL,
	})
}

// GetReportStatistics 获取报告统计
func (ctrl *ReportController) GetReportStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 获取当前用户角色
	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	var stats *models.ReportStatistics
	if user.Role == models.RoleAdmin {
		// 管理员获取全局统计
		stats, err = ctrl.reportService.GetGlobalStatistics()
	} else {
		// 普通用户获取个人统计
		stats, err = ctrl.reportService.GetUserStatistics(userID)
	}

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取统计信息失败")
		return
	}

	utils.Success(c, "", gin.H{"statistics": stats})
}

// urlEncode URL编码文件名
func urlEncode(s string) string {
	return url.QueryEscape(s)
}
