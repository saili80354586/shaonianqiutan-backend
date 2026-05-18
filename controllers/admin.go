package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// AdminController 管理后台控制器
type AdminController struct {
	adminService      *services.AdminService
	videoAnalysisRepo *models.VideoAnalysisRepository
}

func NewAdminController(adminService *services.AdminService, videoAnalysisRepo *models.VideoAnalysisRepository) *AdminController {
	return &AdminController{adminService: adminService, videoAnalysisRepo: videoAnalysisRepo}
}

func (ctrl *AdminController) writeAuditLog(c *gin.Context, action, target string, targetID uint, detail string) {
	if ctrl == nil || ctrl.adminService == nil {
		return
	}
	adminID := c.GetUint("userId")
	if adminID == 0 {
		adminID = c.GetUint("adminId")
	}
	adminName := "管理员"
	if userValue, exists := c.Get("user"); exists {
		if user, ok := userValue.(*models.User); ok && user != nil {
			if strings.TrimSpace(user.Nickname) != "" {
				adminName = user.Nickname
			} else if strings.TrimSpace(user.Name) != "" {
				adminName = user.Name
			} else if strings.TrimSpace(user.Phone) != "" {
				adminName = user.Phone
			}
		}
	}
	if err := ctrl.adminService.CreateAdminOperationLog(&models.AdminOperationLog{
		ClubID:    0,
		AdminID:   adminID,
		AdminName: adminName,
		Action:    action,
		Target:    target,
		TargetID:  targetID,
		Detail:    detail,
		IP:        c.ClientIP(),
		CreatedAt: time.Now(),
	}); err != nil {
		fmt.Printf("[AdminAudit] write audit log failed: %v\n", err)
	}
}

func (ctrl *AdminController) GetSystemSettings(c *gin.Context) {
	settings := models.LoadAdminSystemSettings(config.GetDB())
	utils.Success(c, "", settings)
}

func (ctrl *AdminController) GetPublicSystemSettings(c *gin.Context) {
	settings := models.LoadAdminSystemSettings(config.GetDB())
	utils.Success(c, "", settings.Public())
}

func (ctrl *AdminController) UpdateSystemSettings(c *gin.Context) {
	var req models.AdminSystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.SiteName) == "" {
		utils.Error(c, http.StatusBadRequest, "网站名称不能为空")
		return
	}
	if req.DefaultAnalystPrice < 0 {
		utils.Error(c, http.StatusBadRequest, "默认分析师价格不能小于0")
		return
	}
	if req.CommissionRate < 0 || req.CommissionRate > 100 {
		utils.Error(c, http.StatusBadRequest, "平台佣金比例必须在0到100之间")
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "设置序列化失败")
		return
	}

	row := models.SystemSetting{
		Key:       models.AdminSystemSettingKey,
		Value:     string(payload),
		UpdatedAt: time.Now(),
	}
	if err := config.GetDB().Where("key = ?", models.AdminSystemSettingKey).Assign(row).FirstOrCreate(&row).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存系统设置失败")
		return
	}

	ctrl.writeAuditLog(c, "update_system_settings", "system_settings", 0, fmt.Sprintf("更新系统设置：站点=%s，佣金=%.2f%%", req.SiteName, req.CommissionRate))
	utils.Success(c, "系统设置已保存", req)
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

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fileName)))
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
	days := utils.ParseIntQuery(c, "days", 30, 1, 365)

	data, err := ctrl.adminService.GetGrowthData(days)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取增长数据失败")
		return
	}

	utils.Success(c, "", data)
}

// GetUserList 获取用户列表
func (ctrl *AdminController) GetUserList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	filters := models.AdminUserListFilters{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Role:    strings.TrimSpace(c.Query("role")),
		Status:  strings.TrimSpace(c.Query("status")),
		City:    strings.TrimSpace(c.Query("city")),
	}
	if ageMinStr := strings.TrimSpace(c.Query("ageMin")); ageMinStr != "" {
		ageMin, err := strconv.Atoi(ageMinStr)
		if err != nil || ageMin < 0 {
			utils.Error(c, http.StatusBadRequest, "无效的最小年龄")
			return
		}
		filters.AgeMin = &ageMin
	}
	if ageMaxStr := strings.TrimSpace(c.Query("ageMax")); ageMaxStr != "" {
		ageMax, err := strconv.Atoi(ageMaxStr)
		if err != nil || ageMax < 0 {
			utils.Error(c, http.StatusBadRequest, "无效的最大年龄")
			return
		}
		filters.AgeMax = &ageMax
	}
	if filters.AgeMin != nil && filters.AgeMax != nil && *filters.AgeMin > *filters.AgeMax {
		utils.Error(c, http.StatusBadRequest, "最小年龄不能大于最大年龄")
		return
	}

	users, total, err := ctrl.adminService.GetUserList(page, pageSize, filters)
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

// UpdateUserRequest 更新用户基础资料请求
type UpdateUserRequest struct {
	Phone       *string `json:"phone"`
	Nickname    *string `json:"nickname"`
	Name        *string `json:"name"`
	Role        *string `json:"role"`
	CurrentRole *string `json:"current_role"`
	Status      *string `json:"status"`
	Province    *string `json:"province"`
	City        *string `json:"city"`
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

	ctrl.writeAuditLog(c, "update_user_status", "user", uint(id), "更新用户状态为："+req.Status)
	utils.Success(c, "状态更新成功", nil)
}

// UpdateUser 更新用户基础资料
func (ctrl *AdminController) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Phone != nil {
		updates["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Nickname != nil {
		updates["nickname"] = strings.TrimSpace(*req.Nickname)
	}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Role != nil {
		updates["role"] = strings.TrimSpace(*req.Role)
	}
	if req.CurrentRole != nil {
		updates["current_role"] = strings.TrimSpace(*req.CurrentRole)
	}
	if req.Status != nil {
		updates["status"] = strings.TrimSpace(*req.Status)
	}
	if req.Province != nil {
		updates["province"] = strings.TrimSpace(*req.Province)
	}
	if req.City != nil {
		updates["city"] = strings.TrimSpace(*req.City)
	}

	user, err := ctrl.adminService.UpdateUser(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "更新失败: "+err.Error())
		return
	}

	ctrl.writeAuditLog(c, "update_user", "user", uint(id), "更新用户基础资料")
	utils.Success(c, "用户已更新", user)
}

// GetAllOrders 获取所有订单列表
func (ctrl *AdminController) GetAllOrders(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := c.DefaultQuery("status", "")

	var list interface{}
	var total int64
	var err error
	if c.Query("include_progress") == "true" {
		list, total, err = ctrl.adminService.GetAllOrdersWithProgress(page, pageSize, status)
	} else {
		list, total, err = ctrl.adminService.GetAllOrders(page, pageSize, status)
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetAssignmentRecords 获取订单派发记录
func (ctrl *AdminController) GetAssignmentRecords(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := c.DefaultQuery("status", "")

	assignments, total, err := ctrl.adminService.GetAssignmentRecords(page, pageSize, status)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "获取派发记录失败: "+err.Error())
		return
	}

	utils.Success(c, "", gin.H{
		"list":     assignments,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetOrderStatusHistory 获取订单状态流转历史
func (ctrl *AdminController) GetOrderStatusHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	histories, err := ctrl.adminService.GetOrderStatusHistory(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取状态历史失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list": histories,
	})
}

// GetOrderAnalysisProgress 获取单个订单的分析师分析进度详情
func (ctrl *AdminController) GetOrderAnalysisProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	detail, err := ctrl.adminService.GetOrderAnalysisProgressDetail(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "获取分析进度失败")
		return
	}

	utils.Success(c, "", detail)
}

type OrderProgressReminderRequest struct {
	Message string `json:"message"`
}

// SendOrderProgressReminder 管理员催办订单分析进度
func (ctrl *AdminController) SendOrderProgressReminder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}
	var req OrderProgressReminderRequest
	_ = c.ShouldBindJSON(&req)
	if err := ctrl.adminService.SendOrderProgressReminder(uint(id), c.GetUint("userId"), req.Message); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ctrl.writeAuditLog(c, "send_order_progress_reminder", "order", uint(id), "管理员催办订单分析进度")
	utils.Success(c, "已发送催办提醒", nil)
}

type OrderProgressExceptionRequest struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MarkOrderProgressException 管理员标记订单分析异常
func (ctrl *AdminController) MarkOrderProgressException(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}
	var req OrderProgressExceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.adminService.MarkOrderProgressException(uint(id), c.GetUint("userId"), req.Code, req.Message); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ctrl.writeAuditLog(c, "mark_order_progress_exception", "order", uint(id), "管理员标记订单分析异常："+req.Message)
	utils.Success(c, "已标记异常", nil)
}

// ResolveOrderProgressException 管理员解除订单分析异常
func (ctrl *AdminController) ResolveOrderProgressException(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}
	var req OrderProgressExceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.adminService.ResolveOrderProgressException(uint(id), c.GetUint("userId"), req.Code, req.Message); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ctrl.writeAuditLog(c, "resolve_order_progress_exception", "order", uint(id), "管理员解除订单分析异常："+req.Code)
	utils.Success(c, "已解除异常", nil)
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

	err = ctrl.adminService.ReviewReport(uint(id), req.Status, req.Remark, c.GetUint("userId"))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "审核失败: "+err.Error())
		return
	}

	ctrl.writeAuditLog(c, "review_report", "report", uint(id), "审核报告状态为："+string(req.Status))
	utils.Success(c, "审核完成", nil)
}

// DownloadReportDoc 管理员下载报告 MD 文档、正式 Word 或 PDF 报告
func (ctrl *AdminController) DownloadReportDoc(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的报告ID")
		return
	}

	docType := c.DefaultQuery("type", "rating") // rating | player-info | report | pdf
	if docType != "rating" && docType != "player-info" && docType != "report" && docType != "pdf" {
		utils.Error(c, http.StatusBadRequest, "无效的文档类型")
		return
	}

	report, err := ctrl.adminService.GetReportByID(uint(reportID))
	if err != nil || report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	var filePath string
	var fileName string
	var contentType string
	if docType == "rating" {
		filePath = report.RatingReportMD
		fileName = fmt.Sprintf("评分报告_%d.md", report.OrderID)
		contentType = "text/markdown; charset=utf-8"
	} else if docType == "player-info" {
		filePath = report.PlayerInfoMD
		fileName = fmt.Sprintf("球员基础信息_%s.md", report.PlayerName)
		contentType = "text/markdown; charset=utf-8"
	} else if docType == "pdf" {
		if strings.TrimSpace(report.PdfURL) == "" {
			utils.Error(c, http.StatusNotFound, "PDF 报告尚未生成")
			return
		}
		filePath = adminReportFileRef(report.PdfURL)
		fileName = adminReportFileName(report.PdfURL, fmt.Sprintf("正式报告_%s.pdf", report.PlayerName))
		contentType = "application/pdf"
	} else {
		if strings.TrimSpace(report.AIReportURL) == "" {
			utils.Error(c, http.StatusNotFound, "Word 报告尚未生成")
			return
		}
		filePath = adminReportFileRef(report.AIReportURL)
		fileName = adminReportFileName(report.AIReportURL, fmt.Sprintf("正式报告_%s.docx", report.PlayerName))
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if filePath == "" {
		utils.Error(c, http.StatusNotFound, "文档不存在")
		return
	}

	if isRemoteReportFile(filePath) {
		streamRemoteReportFile(c, filePath, fileName, contentType)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fileName)))
	c.Header("Content-Type", firstNonEmptyContentType(contentType, adminReportAttachmentContentType(filePath)))
	c.File(filePath)
}

func adminReportFileRef(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || isRemoteReportFile(value) {
		return value
	}
	if strings.HasPrefix(value, "/uploads/") {
		return filepath.Join(".", strings.TrimPrefix(value, "/"))
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(".", strings.TrimPrefix(value, "/"))
}

func adminReportFileName(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if parsedURL, err := url.Parse(value); err == nil && parsedURL.Path != "" {
		value = parsedURL.Path
	}
	if base := filepath.Base(strings.TrimPrefix(value, "/")); base != "" && base != "." && base != "/" {
		return base
	}
	return fallback
}

func isRemoteReportFile(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func streamRemoteReportFile(c *gin.Context, fileURL string, fileName string, contentType string) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		utils.Error(c, http.StatusBadGateway, "远程报告文件下载失败")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		utils.Error(c, http.StatusNotFound, "远程报告文件不存在")
		return
	}
	if contentType == "" {
		contentType = resp.Header.Get("Content-Type")
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fileName)))
	c.Header("Content-Type", firstNonEmptyContentType(contentType, "application/octet-stream"))
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func firstNonEmptyContentType(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "application/octet-stream"
}

func adminReportAttachmentContentType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".pdf":
		return "application/pdf"
	case ".md":
		return "text/markdown; charset=utf-8"
	default:
		return "application/octet-stream"
	}
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
	db := ctrl.adminService.GetDB()
	var analysis models.VideoAnalysis
	if db != nil && db.Migrator().HasTable(&models.VideoAnalysis{}) {
		_ = db.Where("order_id = ?", report.OrderID).First(&analysis).Error
	}
	adminID := c.GetUint("userId")
	if err := models.CreateReportVersion(db, &models.ReportVersion{
		ReportID:                report.ID,
		OrderID:                 report.OrderID,
		AnalysisID:              reportVersionAnalysisID(analysis.ID),
		SourceType:              models.ReportVersionSourceAdminWord,
		Status:                  models.ReportVersionStatusAnalystSubmitted,
		Content:                 report.Content,
		WordURL:                 fileURL,
		PDFURL:                  report.PdfURL,
		InputSnapshot:           analysis.AIReportInputSnapshot,
		TemplateVersion:         analysis.AIReportTemplateVersion,
		DocumentTemplateVersion: services.VideoAnalysisDocumentTemplateVersion,
		OriginalFileName:        file.Filename,
		CreatedByUserID:         &adminID,
		CreatedByRole:           "admin",
	}); err != nil {
		utils.Error(c, http.StatusInternalServerError, "记录报告版本失败")
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

	ctrl.writeAuditLog(c, "delete_user", "user", uint(id), "删除用户")
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

	adminID := c.GetUint("adminId")
	order, err := ctrl.adminService.AssignOrderWithRequest(uint(id), req, adminID)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "派单失败: "+err.Error())
		return
	}

	detail := fmt.Sprintf("订单派发给分析师ID=%d", req.AnalystID)
	if req.Deadline != nil {
		detail += "，截止时间=" + req.Deadline.Format("2006-01-02 15:04:05")
	}
	if note := strings.TrimSpace(req.Note); note != "" {
		detail += "，备注=" + note
	}
	ctrl.writeAuditLog(c, "assign_order", "order", uint(id), detail)
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

	adminID := c.GetUint("adminId")
	err = ctrl.adminService.CancelOrder(uint(id), adminID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "取消订单失败: "+err.Error())
		return
	}

	ctrl.writeAuditLog(c, "cancel_order", "order", uint(id), "管理员取消订单")
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

	ctrl.writeAuditLog(c, "audit_analyst", "analyst", uint(id), "审核分析师状态为："+req.Status)
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

	ctrl.writeAuditLog(c, "update_analyst_status", "analyst", uint(id), "更新分析师状态为："+req.Status)
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
	days := utils.ParseIntQuery(c, "days", 30, 1, 365)
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
	days := utils.ParseIntQuery(c, "days", 30, 1, 365)
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
	ctrl.writeAuditLog(c, "process_settlement", "settlement", 0, fmt.Sprintf("处理结算订单数=%d", len(req.OrderIDs)))
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
	ctrl.writeAuditLog(c, "handle_content_report", "content_report", uint(id), "处理举报状态为："+req.Status)
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
	ctrl.writeAuditLog(c, "create_sensitive_word", "sensitive_word", word.ID, "创建敏感词："+req.Word)
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
	ctrl.writeAuditLog(c, "update_sensitive_word", "sensitive_word", uint(id), "更新敏感词配置")
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
	ctrl.writeAuditLog(c, "delete_sensitive_word", "sensitive_word", uint(id), "删除敏感词")
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
	ctrl.writeAuditLog(c, "create_announcement", "announcement", ann.ID, "发布平台公告："+req.Title)
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
	ctrl.writeAuditLog(c, "update_announcement", "announcement", uint(id), "更新平台公告")
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
	ctrl.writeAuditLog(c, "delete_announcement", "announcement", uint(id), "删除平台公告")
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
	ctrl.writeAuditLog(c, "create_banner", "banner", banner.ID, "创建轮播图："+req.Title)
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
	ctrl.writeAuditLog(c, "update_banner", "banner", uint(id), "更新轮播图")
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
	ctrl.writeAuditLog(c, "delete_banner", "banner", uint(id), "删除轮播图")
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

// GetPublicFAQs 获取前台帮助中心 FAQ 列表
func (ctrl *AdminController) GetPublicFAQs(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 50)
	category := c.DefaultQuery("category", "")
	enabled := true
	faqs, total, err := ctrl.adminService.GetFAQs(pagination.Page, pagination.PageSize, category, &enabled)
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
	ctrl.writeAuditLog(c, "create_faq", "faq", faq.ID, "创建FAQ："+req.Question)
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
	ctrl.writeAuditLog(c, "update_faq", "faq", uint(id), "更新FAQ")
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
	ctrl.writeAuditLog(c, "delete_faq", "faq", uint(id), "删除FAQ")
	utils.Success(c, "删除成功", nil)
}

// GetHelpGuides 获取使用指南列表
func (ctrl *AdminController) GetHelpGuides(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	role := c.DefaultQuery("role", "")
	enabledStr := c.DefaultQuery("enabled", "")
	var enabled *bool
	if enabledStr != "" {
		e := enabledStr == "true"
		enabled = &e
	}
	guides, total, err := ctrl.adminService.GetHelpGuides(pagination.Page, pagination.PageSize, role, enabled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取使用指南失败")
		return
	}
	utils.Success(c, "", gin.H{"list": guides, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// GetPublicHelpGuides 获取前台帮助中心使用指南
func (ctrl *AdminController) GetPublicHelpGuides(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 50)
	role := c.DefaultQuery("role", "")
	enabled := true
	guides, total, err := ctrl.adminService.GetHelpGuides(pagination.Page, pagination.PageSize, role, &enabled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取使用指南失败")
		return
	}
	utils.Success(c, "", gin.H{"list": guides, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

type CreateHelpGuideRequest struct {
	Role      string `json:"role" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Summary   string `json:"summary"`
	Content   string `json:"content" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
}

// CreateHelpGuide 创建使用指南
func (ctrl *AdminController) CreateHelpGuide(c *gin.Context) {
	adminID := c.GetUint("userId")
	var req CreateHelpGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	guide := &models.HelpGuide{
		Role: req.Role, Title: req.Title, Summary: req.Summary, Content: req.Content,
		SortOrder: req.SortOrder, Enabled: req.Enabled, CreatedBy: adminID,
	}
	if err := ctrl.adminService.CreateHelpGuide(guide); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	ctrl.writeAuditLog(c, "create_help_guide", "help_guide", guide.ID, "创建使用指南："+req.Title)
	utils.Success(c, "创建成功", guide)
}

type UpdateHelpGuideRequest struct {
	Role      string `json:"role"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}

// UpdateHelpGuide 更新使用指南
func (ctrl *AdminController) UpdateHelpGuide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的使用指南ID")
		return
	}
	var req UpdateHelpGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if err := ctrl.adminService.UpdateHelpGuide(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	ctrl.writeAuditLog(c, "update_help_guide", "help_guide", uint(id), "更新使用指南")
	utils.Success(c, "更新成功", nil)
}

// DeleteHelpGuide 删除使用指南
func (ctrl *AdminController) DeleteHelpGuide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的使用指南ID")
		return
	}
	if err := ctrl.adminService.DeleteHelpGuide(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	ctrl.writeAuditLog(c, "delete_help_guide", "help_guide", uint(id), "删除使用指南")
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

// GetAuditLogs 获取管理员操作审计日志
func (ctrl *AdminController) GetAuditLogs(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	action := c.DefaultQuery("action", "")
	target := c.DefaultQuery("target", "")
	keyword := c.DefaultQuery("keyword", c.DefaultQuery("search", ""))

	logs, total, err := ctrl.adminService.GetAdminOperationLogs(pagination.Page, pagination.PageSize, action, target, keyword)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取审计日志失败")
		return
	}

	items := make([]*models.AdminOperationLogResponse, 0, len(logs))
	for i := range logs {
		items = append(items, logs[i].ToResponse())
	}
	utils.Success(c, "", gin.H{"list": items, "total": total, "page": pagination.Page, "pageSize": pagination.PageSize})
}

// GetRolePermissions 获取管理员子角色权限矩阵
func (ctrl *AdminController) GetRolePermissions(c *gin.Context) {
	data, err := ctrl.adminService.GetAdminRolePermissions(c.GetUint("userId"))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取角色权限失败")
		return
	}
	utils.Success(c, "", data)
}

// GetMyAdminPermissions 获取当前管理员自身权限，用于前端菜单裁剪。
func (ctrl *AdminController) GetMyAdminPermissions(c *gin.Context) {
	data, err := ctrl.adminService.GetCurrentAdminPermissions(c.GetUint("userId"))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取当前权限失败")
		return
	}
	utils.Success(c, "", data)
}

type AssignAdminRoleRequest struct {
	RoleKey string `json:"role_key" binding:"required"`
}

type AdminRoleMutationRequest struct {
	Key         string   `json:"key"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

type UpdateAdminRoleStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type BatchAssignAdminRoleRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
	RoleKey string `json:"role_key" binding:"required"`
}

// CreateAdminRole 创建自定义管理员子角色
func (ctrl *AdminController) CreateAdminRole(c *gin.Context) {
	var req AdminRoleMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	role, err := ctrl.adminService.CreateAdminRole(services.AdminRoleMutationRequest{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "创建角色失败: "+err.Error())
		return
	}

	detailBytes, _ := json.Marshal(role)
	ctrl.writeAuditLog(c, "create_admin_role", "admin_role", 0, string(detailBytes))
	utils.Success(c, "角色已创建", role)
}

// UpdateAdminRole 更新自定义管理员子角色权限
func (ctrl *AdminController) UpdateAdminRole(c *gin.Context) {
	roleKey := strings.TrimSpace(c.Param("roleKey"))
	if roleKey == "" {
		utils.Error(c, http.StatusBadRequest, "无效的角色标识")
		return
	}

	var req AdminRoleMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	role, err := ctrl.adminService.UpdateAdminRole(roleKey, services.AdminRoleMutationRequest{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "更新角色失败: "+err.Error())
		return
	}

	detailBytes, _ := json.Marshal(role)
	ctrl.writeAuditLog(c, "update_admin_role", "admin_role", 0, string(detailBytes))
	utils.Success(c, "角色已更新", role)
}

// UpdateAdminRoleStatus 启用或禁用自定义管理员子角色
func (ctrl *AdminController) UpdateAdminRoleStatus(c *gin.Context) {
	roleKey := strings.TrimSpace(c.Param("roleKey"))
	if roleKey == "" {
		utils.Error(c, http.StatusBadRequest, "无效的角色标识")
		return
	}

	var req UpdateAdminRoleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	role, err := ctrl.adminService.UpdateAdminRoleStatus(roleKey, req.Enabled)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "更新角色状态失败: "+err.Error())
		return
	}

	detailBytes, _ := json.Marshal(role)
	ctrl.writeAuditLog(c, "update_admin_role_status", "admin_role", 0, string(detailBytes))
	utils.Success(c, "角色状态已更新", role)
}

// AssignAdminRole 给管理员账号分配子角色
func (ctrl *AdminController) AssignAdminRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil || userID == 0 {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var req AssignAdminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.adminService.AssignAdminRole(uint(userID), req.RoleKey, c.GetUint("userId"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "分配角色失败: "+err.Error())
		return
	}
	detailBytes, _ := json.Marshal(result)
	ctrl.writeAuditLog(c, "assign_admin_role", "admin_user_role", uint(userID), string(detailBytes))
	utils.Success(c, "角色已分配", result)
}

// BatchAssignAdminRole 批量给管理员账号分配子角色
func (ctrl *AdminController) BatchAssignAdminRole(c *gin.Context) {
	var req BatchAssignAdminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	results, err := ctrl.adminService.BatchAssignAdminRole(req.UserIDs, req.RoleKey, c.GetUint("userId"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "批量分配角色失败: "+err.Error())
		return
	}

	for _, result := range results {
		detailBytes, _ := json.Marshal(result)
		ctrl.writeAuditLog(c, "assign_admin_role", "admin_user_role", result.TargetUserID, string(detailBytes))
	}

	utils.Success(c, "角色已批量分配", gin.H{"list": results, "total": len(results)})
}

// GetAdminRoleAssignmentHistory 获取管理员子角色授权历史
func (ctrl *AdminController) GetAdminRoleAssignmentHistory(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	var targetUserID uint
	if rawTargetUserID := strings.TrimSpace(c.Query("target_user_id")); rawTargetUserID != "" {
		parsed, err := strconv.ParseUint(rawTargetUserID, 10, 32)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "无效的管理员用户ID")
			return
		}
		targetUserID = uint(parsed)
	}

	data, err := ctrl.adminService.GetAdminRoleAssignmentHistory(pagination.Page, pagination.PageSize, targetUserID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取授权历史失败")
		return
	}
	utils.Success(c, "", data)
}

// GetExceptions 获取异常管控列表
func (ctrl *AdminController) GetExceptions(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	status := c.DefaultQuery("status", "")
	source := c.DefaultQuery("source", "")
	severity := c.DefaultQuery("severity", "")

	data, err := ctrl.adminService.GetExceptions(pagination.Page, pagination.PageSize, status, source, severity)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取异常列表失败")
		return
	}
	utils.Success(c, "", data)
}

// GetLoginLogStats 获取登录日志统计
func (ctrl *AdminController) GetLoginLogStats(c *gin.Context) {
	days := utils.ParseIntQuery(c, "days", 7, 1, 365)
	stats, err := ctrl.adminService.GetLoginLogStats(days)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取登录日志统计失败")
		return
	}
	utils.Success(c, "", stats)
}
