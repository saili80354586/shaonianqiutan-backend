package controllers

import (
	"net/http"
	"net/url"
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

// ReportController 报告控制器
type ReportController struct {
	reportService *services.ReportService
	authService   *services.AuthService
	analysisRepo  *models.VideoAnalysisRepository
	highlightRepo *models.AnalysisHighlightRepository
}

func NewReportController(reportService *services.ReportService, authService *services.AuthService, dbs ...*gorm.DB) *ReportController {
	ctrl := &ReportController{
		reportService: reportService,
		authService:   authService,
	}
	if len(dbs) > 0 && dbs[0] != nil {
		ctrl.analysisRepo = models.NewVideoAnalysisRepository(dbs[0])
		ctrl.highlightRepo = models.NewAnalysisHighlightRepository(dbs[0])
	}
	return ctrl
}

type reportHighlightMarker struct {
	ID              uint                       `json:"id"`
	AnalysisID      uint                       `json:"analysis_id"`
	Timestamp       string                     `json:"timestamp"`
	MarkerType      models.HighlightMarkerType `json:"marker_type"`
	Mode            models.HighlightMode       `json:"mode"`
	StartTimeMs     int                        `json:"start_time_ms"`
	EndTimeMs       *int                       `json:"end_time_ms"`
	TagType         models.HighlightTagType    `json:"tag_type"`
	Description     string                     `json:"description"`
	VideoClipURL    string                     `json:"video_clip_url,omitempty"`
	ClipStatus      models.HighlightClipStatus `json:"clip_status"`
	IncludeInReport bool                       `json:"include_in_report"`
	SortOrder       int                        `json:"sort_order"`
	CreatedAt       time.Time                  `json:"created_at"`
}

type reportDetailResponse struct {
	*models.Report
	VideoAnalysisID  uint                    `json:"video_analysis_id,omitempty"`
	HighlightMarkers []reportHighlightMarker `json:"highlight_markers"`
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

	highlightMarkers, videoAnalysisID := ctrl.getReportHighlightMarkers(report, user.Role)
	utils.Success(c, "", gin.H{"data": reportDetailResponse{
		Report:           report,
		VideoAnalysisID:  videoAnalysisID,
		HighlightMarkers: highlightMarkers,
	}})
}

func (ctrl *ReportController) getReportHighlightMarkers(report *models.Report, userRole models.UserRole) ([]reportHighlightMarker, uint) {
	if ctrl.analysisRepo == nil || ctrl.highlightRepo == nil || report == nil || report.OrderID == 0 {
		return []reportHighlightMarker{}, 0
	}
	if report.Status != models.ReportStatusCompleted && userRole == models.RoleUser {
		return []reportHighlightMarker{}, 0
	}

	analysis, err := ctrl.analysisRepo.FindByOrderID(report.OrderID)
	if err != nil || analysis == nil {
		return []reportHighlightMarker{}, 0
	}
	highlights, err := ctrl.highlightRepo.FindIncludedInReport(analysis.ID)
	if err != nil {
		return []reportHighlightMarker{}, analysis.ID
	}

	markers := make([]reportHighlightMarker, 0, len(highlights))
	for _, highlight := range highlights {
		videoClipURL := ""
		if highlight.Mode == models.HighlightModeRange && highlight.ClipStatus == models.HighlightClipReady {
			videoClipURL = highlight.VideoClipURL
		}
		markers = append(markers, reportHighlightMarker{
			ID:              highlight.ID,
			AnalysisID:      highlight.AnalysisID,
			Timestamp:       highlight.Timestamp,
			MarkerType:      highlight.MarkerType,
			Mode:            highlight.Mode,
			StartTimeMs:     highlight.StartTimeMs,
			EndTimeMs:       highlight.EndTimeMs,
			TagType:         highlight.TagType,
			Description:     highlight.Description,
			VideoClipURL:    videoClipURL,
			ClipStatus:      highlight.ClipStatus,
			IncludeInReport: highlight.IncludeInReport,
			SortOrder:       highlight.SortOrder,
			CreatedAt:       highlight.CreatedAt,
		})
	}
	return markers, analysis.ID
}

// DownloadReport 下载报告。优先正式 PDF，其次分析师上传的 Word，最后回退在线报告正文。
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

	if report.PdfURL != "" {
		filePath := ctrl.reportService.GetPdfFilePath(report)
		if _, err := os.Stat(filePath); err == nil {
			ctrl.sendReportFile(c, filePath, ".pdf", "application/pdf", report)
			return
		}
	}

	if report.AIReportURL != "" {
		filePath := filepath.Join(".", strings.TrimPrefix(report.AIReportURL, "/"))
		if _, err := os.Stat(filePath); err == nil {
			ext := strings.ToLower(filepath.Ext(filePath))
			ctrl.sendReportFile(c, filePath, ext, reportAttachmentContentType(ext), report)
			return
		}
	}

	if strings.TrimSpace(report.Content) != "" {
		fileName := reportDownloadBaseName(report) + ".md"
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+urlEncode(fileName))
		c.String(http.StatusOK, report.Content)
		return
	}

	utils.Error(c, http.StatusNotFound, "报告文件不存在")
}

func (ctrl *ReportController) sendReportFile(c *gin.Context, filePath, ext, contentType string, report *models.Report) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if ext == "" {
		ext = filepath.Ext(filePath)
	}
	fileName := reportDownloadBaseName(report) + ext
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+urlEncode(fileName))
	c.File(filePath)
}

func reportDownloadBaseName(report *models.Report) string {
	if report == nil || strings.TrimSpace(report.PlayerName) == "" {
		return "球探报告"
	}
	return strings.TrimSpace(report.PlayerName) + "_球探报告"
}

func reportAttachmentContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".md":
		return "text/markdown; charset=utf-8"
	default:
		return "application/octet-stream"
	}
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
