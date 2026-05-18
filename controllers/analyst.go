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
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// AnalystController 分析师控制器
type AnalystController struct {
	analystService *services.AnalystService
	db             *gorm.DB
}

// NewAnalystController 创建分析师控制器
func NewAnalystController(analystService *services.AnalystService, db *gorm.DB) *AnalystController {
	return &AnalystController{analystService: analystService, db: db}
}

// GetAnalystList 获取分析师列表
func (ctrl *AnalystController) GetAnalystList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	analysts, total, err := ctrl.analystService.GetAnalystList(page, pageSize)
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

// GetAnalystByID 获取分析师详情
func (ctrl *AnalystController) GetAnalystByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	analyst, err := ctrl.analystService.GetAnalystByID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, "", gin.H{"analyst": analyst})
}

// GetAnalystPublicProfile 获取分析师公开主页数据
func (ctrl *AnalystController) GetAnalystPublicProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	profile, err := ctrl.analystService.GetAnalystPublicProfile(uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, "", profile)
}

// GetAnalystOfficialWorks 获取分析师公开主页官方采用作品
func (ctrl *AnalystController) GetAnalystOfficialWorks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 6)
	works, err := ctrl.analystService.GetAnalystOfficialWorks(uint(id), pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取官方采用作品失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     works,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

// GetAnalystPublicProfileByUser 通过 user_id 获取分析师公开主页数据
func (ctrl *AnalystController) GetAnalystPublicProfileByUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		utils.Error(c, http.StatusBadRequest, "缺少 user_id 参数")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的 user_id")
		return
	}

	// 先通过 user_id 获取 analyst
	analyst, err := ctrl.analystService.GetAnalystByUserID(uint(userID))
	if err != nil || analyst == nil {
		utils.Error(c, http.StatusNotFound, "该用户不是分析师或分析师不存在")
		return
	}

	profile, err := ctrl.analystService.GetAnalystPublicProfile(analyst.ID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, "", profile)
}

// GetMyAnalystProfile 获取当前用户的分析师资料
func (ctrl *AnalystController) GetMyAnalystProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	analyst, err := ctrl.analystService.GetAnalystByUserID(userID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, "", gin.H{"analyst": analyst})
}

// UpdateMyProfile 更新分析师资料
func (ctrl *AnalystController) UpdateMyProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 获取分析师ID
	analyst, err := ctrl.analystService.GetAnalystByUserID(userID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "您不是分析师")
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.analystService.UpdateAnalystProfile(analyst.ID, &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "资料更新成功", nil)
}

// GetMyOrders 获取当前分析师的订单列表
func (ctrl *AnalystController) GetMyOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 获取分析师ID
	analyst, err := ctrl.analystService.GetAnalystByUserID(userID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "您不是分析师")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	orders, total, err := ctrl.analystService.GetAnalystOrders(analyst.ID, page, pageSize)
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

// GetMyRevenue 获取当前分析师的收益统计
func (ctrl *AnalystController) GetMyRevenue(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 获取分析师ID
	analyst, err := ctrl.analystService.GetAnalystByUserID(userID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "您不是分析师")
		return
	}

	revenue, err := ctrl.analystService.GetAnalystRevenue(analyst.ID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取收益统计失败")
		return
	}

	utils.Success(c, "", gin.H{"revenue": revenue})
}

// GetDashboardStats 获取工作台统计
func (ctrl *AnalystController) GetDashboardStats(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	stats, err := ctrl.analystService.GetDashboardStats(analystID.(uint))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", stats)
}

// GetPendingOrders 获取待处理订单
func (ctrl *AnalystController) GetPendingOrders(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	orders, total, err := ctrl.analystService.GetPendingOrders(analystID.(uint), pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", gin.H{
		"list":     orders,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

// GetActiveOrders 获取进行中订单
func (ctrl *AnalystController) GetActiveOrders(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	orders, total, err := ctrl.analystService.GetActiveOrders(analystID.(uint), pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", gin.H{
		"list":     orders,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

// GetHistoryOrders 获取历史订单
func (ctrl *AnalystController) GetHistoryOrders(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := c.DefaultQuery("status", "")
	orderType := c.DefaultQuery("order_type", "")
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	keyword := c.DefaultQuery("keyword", "")

	orders, total, err := ctrl.analystService.GetHistoryOrders(analystID.(uint), status, orderType, startDate, endDate, keyword, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", gin.H{
		"list":     orders,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AcceptOrder 接单
func (ctrl *AnalystController) AcceptOrder(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	if err := ctrl.analystService.AcceptOrder(analystID.(uint), uint(orderID)); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, "接单成功", nil)
}

// RejectOrder 拒绝订单
func (ctrl *AnalystController) RejectOrder(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请填写拒绝原因")
		return
	}

	if err := ctrl.analystService.RejectOrder(analystID.(uint), uint(orderID), req.Reason); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, "已拒绝订单", nil)
}

// GetIncomeDetails 获取收益明细
func (ctrl *AnalystController) GetIncomeDetails(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")

	result, err := ctrl.analystService.GetIncomeDetails(analystID.(uint), startDate, endDate, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", result)
}

// GetIncomeTrend 获取收益趋势
func (ctrl *AnalystController) GetIncomeTrend(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	rangeParam := c.DefaultQuery("range", "month")
	now := time.Now()
	startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	switch rangeParam {
	case "week":
		startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	case "month":
		startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
	case "quarter":
		startDate = now.AddDate(0, -3, 0).Format("2006-01-02")
	case "year":
		startDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
	}

	trend, err := ctrl.analystService.GetIncomeTrend(analystID.(uint), startDate, endDate)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "", trend)
}

// SubmitReport 提交报告
func (ctrl *AnalystController) SubmitReport(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	var req services.SubmitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.analystService.SubmitReport(analystID.(uint), uint(orderID), &req); err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "报告已提交审核", nil)
}

// DownloadReportDoc 下载报告 MD 文档
func (ctrl *AnalystController) DownloadReportDoc(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	docType := c.DefaultQuery("type", "rating") // rating | player-info
	if docType != "rating" && docType != "player-info" {
		utils.Error(c, http.StatusBadRequest, "无效的文档类型")
		return
	}

	// 获取订单信息，验证权限
	order, err := ctrl.analystService.GetOrderByID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusForbidden, err.Error())
		return
	}
	if order.AnalystID == nil || *order.AnalystID != analystID.(uint) {
		utils.Error(c, http.StatusForbidden, "无权操作该订单")
		return
	}

	// 检查是否有视频分析记录（video/pro 类型的订单会有 video_analyses 记录）
	var analysis models.VideoAnalysis
	if err := ctrl.db.Where("order_id = ?", order.ID).First(&analysis).Error; err == nil && analysis.ID > 0 {
		// 视频分析订单：从 video_analyses 表获取 MD 文档
		var filePath string
		var fileName string
		if docType == "rating" {
			filePath = analysis.RatingReportMD
			fileName = fmt.Sprintf("评分报告_%s.md", order.OrderNo)
		} else {
			filePath = analysis.PlayerInfoMD
			fileName = fmt.Sprintf("球员基础信息_%s.md", order.PlayerName)
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
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.File(filePath)
		return
	}

	// 文字报告订单：从 reports 表获取 MD 文档
	report, err := ctrl.findReportForOrder(order)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告失败")
		return
	}
	if report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	var filePath string
	var fileName string
	if docType == "rating" {
		filePath = report.RatingReportMD
		fileName = fmt.Sprintf("评分报告_%s.md", order.OrderNo)
	} else {
		filePath = report.PlayerInfoMD
		fileName = fmt.Sprintf("球员基础信息_%s.md", order.PlayerName)
	}

	if filePath == "" {
		utils.Error(c, http.StatusNotFound, "文档不存在")
		return
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.File(filePath)
}

// DownloadAIReport 下载 AI 报告（分析师端）
func (ctrl *AnalystController) DownloadAIReport(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	reportType := c.DefaultQuery("type", "report") // report | video
	if reportType != "report" && reportType != "video" {
		utils.Error(c, http.StatusBadRequest, "无效的类型")
		return
	}

	order, err := ctrl.analystService.GetOrderByID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusForbidden, err.Error())
		return
	}
	if order.AnalystID == nil || *order.AnalystID != analystID.(uint) {
		utils.Error(c, http.StatusForbidden, "无权操作该订单")
		return
	}

	report, err := ctrl.findReportForOrder(order)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告失败")
		return
	}
	if report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	var filePath string
	if reportType == "video" {
		filePath = "./uploads/reports/" + strings.TrimPrefix(report.AIVideoURL, "/uploads/reports/")
		if report.AIVideoURL == "" {
			utils.Error(c, http.StatusNotFound, "AI 视频分析尚未上传")
			return
		}
	} else {
		if report.AIReportURL == "" && strings.TrimSpace(report.Content) != "" {
			var analysis models.VideoAnalysis
			if err := ctrl.db.Where("order_id = ?", order.ID).First(&analysis).Error; err == nil && analysis.ID > 0 {
				switch analysis.AIReportStatus {
				case "generating", "regenerating":
					utils.Error(c, http.StatusConflict, "视频分析报告正在生成，请稍后再下载")
					return
				case "failed":
					utils.Error(c, http.StatusConflict, "视频分析报告生成失败，请重新提交或联系管理员")
					return
				default:
					utils.Error(c, http.StatusNotFound, "视频分析 Word 报告尚未生成")
					return
				}
			}
			fileName := fmt.Sprintf("AI分析报告_%s.md", order.OrderNo)
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
			c.Header("Content-Type", "text/markdown; charset=utf-8")
			c.String(http.StatusOK, report.Content)
			return
		}
		filePath = "./uploads/reports/" + strings.TrimPrefix(report.AIReportURL, "/uploads/reports/")
		if report.AIReportURL == "" {
			utils.Error(c, http.StatusNotFound, "AI 报告尚未上传")
			return
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	fileName := filepath.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.File(filePath)
}

func (ctrl *AnalystController) findReportForOrder(order *models.Order) (*models.Report, error) {
	if order == nil {
		return nil, nil
	}
	if order.Report != nil {
		return order.Report, nil
	}

	var report models.Report
	if order.ReportID != nil {
		if err := ctrl.db.First(&report, *order.ReportID).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if report.ID > 0 {
			return &report, nil
		}
	}

	if err := ctrl.db.Where("order_id = ?", order.ID).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// UploadAIReport 上传分析师本地编辑后的 AI 报告 Word 终稿
func (ctrl *AnalystController) UploadAIReport(c *gin.Context) {
	analystID, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	order, err := ctrl.analystService.GetOrderByID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusForbidden, err.Error())
		return
	}
	if order.AnalystID == nil || *order.AnalystID != analystID.(uint) {
		utils.Error(c, http.StatusForbidden, "无权操作该订单")
		return
	}

	report, err := ctrl.findReportForOrder(order)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告失败")
		return
	}
	if report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在，请先生成球探报告")
		return
	}
	if report.Status == models.ReportStatusCompleted {
		utils.Error(c, http.StatusBadRequest, "已交付报告不能直接覆盖，请重新提交审核")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "请选择要上传的 Word 报告")
		return
	}
	if file.Size > 50<<20 {
		utils.Error(c, http.StatusBadRequest, "文件大小不能超过50MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".doc" && ext != ".docx" {
		utils.Error(c, http.StatusBadRequest, "仅支持 .doc 或 .docx 文件")
		return
	}

	uploadDir := "./uploads/reports"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	fileName := fmt.Sprintf("analyst_ai_report_%d_%d%s", report.ID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, fileName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	reportURL := "/uploads/reports/" + fileName
	if err := ctrl.db.Model(&models.Report{}).Where("id = ?", report.ID).Update("ai_report_url", reportURL).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新报告失败")
		return
	}
	var analysis models.VideoAnalysis
	if ctrl.db.Migrator().HasTable(&models.VideoAnalysis{}) {
		_ = ctrl.db.Where("order_id = ?", order.ID).First(&analysis).Error
	}
	if analysis.ID > 0 {
		_ = ctrl.db.Model(&models.VideoAnalysis{}).Where("id = ?", analysis.ID).Update("ai_report_status", string(models.ReportVersionStatusAnalystSubmitted)).Error
	}
	if err := models.CreateReportVersion(ctrl.db, &models.ReportVersion{
		ReportID:                report.ID,
		OrderID:                 order.ID,
		AnalysisID:              reportVersionAnalysisID(analysis.ID),
		SourceType:              models.ReportVersionSourceAnalystWord,
		Status:                  models.ReportVersionStatusAnalystSubmitted,
		Content:                 report.Content,
		WordURL:                 reportURL,
		PDFURL:                  report.PdfURL,
		InputSnapshot:           analysis.AIReportInputSnapshot,
		TemplateVersion:         analysis.AIReportTemplateVersion,
		DocumentTemplateVersion: services.VideoAnalysisDocumentTemplateVersion,
		OriginalFileName:        file.Filename,
		CreatedByRole:           "analyst",
	}); err != nil {
		utils.Error(c, http.StatusInternalServerError, "记录报告版本失败")
		return
	}

	utils.Success(c, "AI 报告已上传", gin.H{
		"report_id":      report.ID,
		"ai_report_url":  reportURL,
		"original_name":  file.Filename,
		"order_id":       order.ID,
		"review_status":  report.Status,
		"requires_audit": report.Status != models.ReportStatusCompleted,
	})
}

// CreateInquiry 提交咨询意向
func (ctrl *AnalystController) CreateInquiry(c *gin.Context) {
	idStr := c.Param("id")
	analystID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析师ID")
		return
	}

	analyst, err := ctrl.analystService.GetAnalystByID(uint(analystID))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "分析师不存在")
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Contact string `json:"contact" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请填写完整的咨询信息")
		return
	}

	// 创建通知给分析师对应的用户
	notification := &models.Notification{
		UserID:    analyst.UserID,
		Type:      models.NotificationTypeInquiry,
		Title:     "收到新的咨询意向",
		Content:   req.Name + "（" + req.Contact + "）向您咨询：" + req.Content,
		IsRead:    false,
		Priority:  2,
		CreatedAt: time.Now(),
	}
	if err := ctrl.db.Create(notification).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "提交失败，请重试")
		return
	}

	utils.Success(c, "咨询意向已提交，分析师将尽快与您联系", nil)
}
