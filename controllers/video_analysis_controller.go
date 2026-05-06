package controllers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// VideoAnalysisController 视频分析控制器
type VideoAnalysisController struct {
	db                  *gorm.DB
	analysisRepo        *models.VideoAnalysisRepository
	highlightRepo       *models.AnalysisHighlightRepository
	aiService           *services.AIService
	clipService         *services.VideoClipService
	clipExportJobs      *highlightClipExportJobManager
	reportGen           *services.ReportGenerator
	notificationService *services.NotificationService
}

// NewVideoAnalysisController 创建视频分析控制器
func NewVideoAnalysisController(db *gorm.DB, aiService *services.AIService) *VideoAnalysisController {
	return &VideoAnalysisController{
		db:             db,
		analysisRepo:   models.NewVideoAnalysisRepository(db),
		highlightRepo:  models.NewAnalysisHighlightRepository(db),
		aiService:      aiService,
		clipService:    services.NewVideoClipService(db),
		clipExportJobs: newHighlightClipExportJobManager(db),
		reportGen:      services.NewReportGenerator("./uploads/reports"),
	}
}

// SetNotificationService 注入通知服务
func (ctrl *VideoAnalysisController) SetNotificationService(notificationService *services.NotificationService) {
	ctrl.notificationService = notificationService
}

func getAnalystIDFromContext(c *gin.Context) (uint, bool) {
	analystIDValue, exists := c.Get("analystId")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return 0, false
	}

	analystID, ok := analystIDValue.(uint)
	if !ok || analystID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return 0, false
	}

	return analystID, true
}

func (ctrl *VideoAnalysisController) ensureAnalysisOwner(c *gin.Context, analysis *models.VideoAnalysis) bool {
	analystID, ok := getAnalystIDFromContext(c)
	if !ok {
		return false
	}
	if analysis.AnalystID != analystID {
		utils.Error(c, http.StatusForbidden, "无权操作此分析")
		return false
	}
	return true
}

func (ctrl *VideoAnalysisController) getOwnedAnalysisByID(c *gin.Context, id uint) (*models.VideoAnalysis, bool) {
	analysis, err := ctrl.analysisRepo.FindByID(id)
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return nil, false
	}
	if !ctrl.ensureAnalysisOwner(c, analysis) {
		return nil, false
	}
	return analysis, true
}

func (ctrl *VideoAnalysisController) getOwnedAnalysisFromParam(c *gin.Context) (*models.VideoAnalysis, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return nil, false
	}
	return ctrl.getOwnedAnalysisByID(c, uint(id))
}

func (ctrl *VideoAnalysisController) getOwnedOrder(c *gin.Context, orderID uint) (*models.Order, bool) {
	analystID, ok := getAnalystIDFromContext(c)
	if !ok {
		return nil, false
	}

	var order models.Order
	if err := ctrl.db.First(&order, orderID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return nil, false
	}
	if order.AnalystID == nil || *order.AnalystID != analystID {
		utils.Error(c, http.StatusForbidden, "无权操作此订单")
		return nil, false
	}
	return &order, true
}

func (ctrl *VideoAnalysisController) getOwnedHighlight(c *gin.Context, id uint) (*models.AnalysisHighlight, *models.VideoAnalysis, bool) {
	var highlight models.AnalysisHighlight
	if err := ctrl.db.First(&highlight, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "高光不存在")
		return nil, nil, false
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, highlight.AnalysisID)
	if !ok {
		return nil, nil, false
	}
	return &highlight, analysis, true
}

func (ctrl *VideoAnalysisController) notifyAdminsReportSubmitted(reportID uint, playerName string) {
	if ctrl.notificationService == nil || reportID == 0 {
		return
	}

	var admins []models.User
	if err := ctrl.db.Where("role = ? AND status = ?", models.RoleAdmin, models.StatusActive).Find(&admins).Error; err != nil {
		log.Printf("[VideoAnalysis] query admins for report notification failed: %v", err)
		return
	}

	adminIDs := make([]uint, 0, len(admins))
	for _, admin := range admins {
		adminIDs = append(adminIDs, admin.ID)
	}
	if len(adminIDs) == 0 {
		return
	}

	if err := ctrl.notificationService.NotifyReportPendingReview(adminIDs, reportID, playerName); err != nil {
		log.Printf("[VideoAnalysis] notify admins for report %d failed: %v", reportID, err)
	}
}

func (ctrl *VideoAnalysisController) syncHighlightClip(highlightID uint, mode models.HighlightMode) *models.AnalysisHighlight {
	if ctrl.clipService == nil {
		return nil
	}

	var (
		highlight *models.AnalysisHighlight
		err       error
	)
	if mode == models.HighlightModeRange {
		highlight, err = ctrl.clipService.QueueHighlightClip(highlightID)
	} else {
		highlight, err = ctrl.clipService.ClearHighlightClip(highlightID)
	}
	if err != nil {
		log.Printf("[VideoAnalysis] sync clip for highlight %d failed: %v", highlightID, err)
		return nil
	}
	return highlight
}

func parseHighlightTimeMs(timestamp string) int {
	parts := strings.Split(strings.TrimSpace(timestamp), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0
	}

	totalSeconds := 0
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return 0
		}
		totalSeconds = totalSeconds*60 + n
	}
	return totalSeconds * 1000
}

func formatHighlightTimeMs(ms int) string {
	if ms < 0 {
		ms = 0
	}
	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return strconv.Itoa(minutes) + ":" + twoDigit(seconds)
}

func twoDigit(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func normalizeHighlightTiming(timestamp string, mode models.HighlightMode, startTimeMs int, endTimeMs *int) (string, models.HighlightMode, int, *int, error) {
	if mode == "" {
		mode = models.HighlightModePoint
	}
	if mode != models.HighlightModePoint && mode != models.HighlightModeRange {
		return "", "", 0, nil, strconv.ErrSyntax
	}

	if startTimeMs == 0 && timestamp != "" {
		startTimeMs = parseHighlightTimeMs(timestamp)
	}

	if mode == models.HighlightModePoint {
		endTimeMs = nil
		if timestamp == "" {
			timestamp = formatHighlightTimeMs(startTimeMs)
		}
		return timestamp, mode, startTimeMs, endTimeMs, nil
	}

	if endTimeMs == nil || *endTimeMs <= startTimeMs {
		return "", "", 0, nil, strconv.ErrSyntax
	}
	if timestamp == "" {
		timestamp = formatHighlightTimeMs(startTimeMs) + "-" + formatHighlightTimeMs(*endTimeMs)
	}
	return timestamp, mode, startTimeMs, endTimeMs, nil
}

// UpdateScoresRequest 更新评分请求
type UpdateScoresRequest struct {
	Scores       *models.VideoAnalysisScores `json:"scores"`
	Summary      string                      `json:"summary"`
	Improvements string                      `json:"improvements"`
	AnalystNotes string                      `json:"analyst_notes"`
}

// UpdateScores 更新评分
func (ctrl *VideoAnalysisController) UpdateScores(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req UpdateScoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}
	if req.Scores == nil {
		utils.Error(c, http.StatusBadRequest, "评分不能为空")
		return
	}

	overallScore := req.Scores.CalculateOverallScore()
	potentialLevel := models.GetPotentialLevel(overallScore)

	scoresJSON, err := req.Scores.ToJSON()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "评分序列化失败")
		return
	}

	updates := map[string]interface{}{
		"scores":          scoresJSON,
		"overall_score":   overallScore,
		"potential_level": potentialLevel,
		"summary":         req.Summary,
		"improvements":    req.Improvements,
		"analyst_notes":   req.AnalystNotes,
		"status":          models.AnalysisStatusScoring,
	}

	err = ctrl.analysisRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存评分失败")
		return
	}

	utils.Success(c, "评分保存成功", gin.H{
		"overall_score":   overallScore,
		"potential_level": potentialLevel,
	})
}

// ConfirmAnalysis 确认并生成 MD 文档
func (ctrl *VideoAnalysisController) ConfirmAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	// 获取球员信息
	var user models.User
	if err := ctrl.db.First(&user, analysis.UserID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取球员信息失败")
		return
	}

	// 获取分析师名称
	analystName := "未知分析师"
	var analyst models.Analyst
	if err := ctrl.db.First(&analyst, analysis.AnalystID).Error; err == nil {
		analystName = analyst.Name
	}

	// 生成 MD 文档
	ratingMD, playerInfoMD, err := ctrl.reportGen.GenerateFromVideoAnalysis(analysis, analystName, &user)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成文档失败: "+err.Error())
		return
	}

	// 更新记录
	updates := map[string]interface{}{
		"rating_report_md": ratingMD,
		"player_info_md":   playerInfoMD,
		"status":           models.AnalysisStatusCompleted,
	}
	if err := ctrl.analysisRepo.Update(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新记录失败")
		return
	}

	utils.Success(c, "文档生成成功", gin.H{
		"rating_report_md": ratingMD,
		"player_info_md":   playerInfoMD,
	})
}

// GetAnalysis 获取分析详情
func (ctrl *VideoAnalysisController) GetAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(uint(id))

	utils.Success(c, "", gin.H{
		"analysis":   analysis,
		"scores":     scores,
		"highlights": highlights,
	})
}

// GetAnalysisByOrder 根据订单获取分析
func (ctrl *VideoAnalysisController) GetAnalysisByOrder(c *gin.Context) {
	orderIDStr := c.Query("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	if _, ok := ctrl.getOwnedOrder(c, uint(orderID)); !ok {
		return
	}

	analysis, err := ctrl.analysisRepo.FindByOrderID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	if analysis == nil {
		utils.Success(c, "", nil)
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(analysis.ID)

	utils.Success(c, "", gin.H{
		"analysis":   analysis,
		"scores":     scores,
		"highlights": highlights,
	})
}

// CreateHighlightRequest 创建高光请求
type CreateHighlightRequest struct {
	AnalysisID      uint                       `json:"analysis_id"`
	Timestamp       string                     `json:"timestamp"`
	MarkerType      models.HighlightMarkerType `json:"marker_type"`
	Mode            models.HighlightMode       `json:"mode"`
	StartTimeMs     int                        `json:"start_time_ms"`
	EndTimeMs       *int                       `json:"end_time_ms"`
	TagType         models.HighlightTagType    `json:"tag_type"`
	Description     string                     `json:"description"`
	VideoClipURL    string                     `json:"video_clip_url"`
	IncludeInReport *bool                      `json:"include_in_report"`
}

// ExportHighlightClipsRequest 批量导出片段请求
type ExportHighlightClipsRequest struct {
	MarkerIDs  []uint                     `json:"marker_ids"`
	MarkerType models.HighlightMarkerType `json:"marker_type"`
	TagType    models.HighlightTagType    `json:"tag_type"`
}

type highlightClipExportJobStatus = models.VideoClipExportJobStatus

const (
	highlightClipExportQueued     = models.VideoClipExportQueued
	highlightClipExportProcessing = models.VideoClipExportProcessing
	highlightClipExportReady      = models.VideoClipExportReady
	highlightClipExportFailed     = models.VideoClipExportFailed
)

// HighlightClipExportJobResponse 批量导出任务状态
type HighlightClipExportJobResponse struct {
	ID          string                       `json:"id"`
	AnalysisID  uint                         `json:"analysis_id"`
	Status      highlightClipExportJobStatus `json:"status"`
	Progress    int                          `json:"progress"`
	Processed   int                          `json:"processed"`
	Total       int                          `json:"total"`
	FileName    string                       `json:"filename"`
	Error       string                       `json:"error,omitempty"`
	DownloadURL string                       `json:"download_url,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
	ExpiresAt   *time.Time                   `json:"expires_at,omitempty"`
}

type highlightClipExportJob struct {
	ID         string
	AnalysisID uint
	AnalystID  uint
	Analysis   models.VideoAnalysis
	Request    ExportHighlightClipsRequest
	Items      []highlightClipExportItem
	Status     highlightClipExportJobStatus
	Progress   int
	Processed  int
	Total      int
	FileName   string
	ZipPath    string
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  *time.Time
}

type highlightClipExportJobManager struct {
	mu    sync.Mutex
	jobs  map[string]*highlightClipExportJob
	queue chan struct{}
	seq   uint64
	ttl   time.Duration
	db    *gorm.DB
}

func newHighlightClipExportJobManager(db *gorm.DB) *highlightClipExportJobManager {
	return &highlightClipExportJobManager{
		jobs:  make(map[string]*highlightClipExportJob),
		queue: make(chan struct{}, 1),
		ttl:   30 * time.Minute,
		db:    db,
	}
}

func (m *highlightClipExportJobManager) start(analysis *models.VideoAnalysis, analystID uint, req ExportHighlightClipsRequest, items []highlightClipExportItem) (HighlightClipExportJobResponse, error) {
	now := time.Now()
	id := fmt.Sprintf("%d-%d", now.UnixNano(), atomic.AddUint64(&m.seq, 1))
	copiedItems := append([]highlightClipExportItem(nil), items...)
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return HighlightClipExportJobResponse{}, err
	}
	job := &highlightClipExportJob{
		ID:         id,
		AnalysisID: analysis.ID,
		AnalystID:  analystID,
		Analysis:   *analysis,
		Request:    req,
		Items:      copiedItems,
		Status:     highlightClipExportQueued,
		Progress:   0,
		Processed:  0,
		Total:      len(copiedItems) + 1,
		FileName:   buildHighlightClipsZipName(analysis),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	record := models.VideoClipExportJob{
		JobID:       id,
		AnalysisID:  analysis.ID,
		AnalystID:   analystID,
		Status:      models.VideoClipExportQueued,
		Progress:    0,
		Processed:   0,
		Total:       len(copiedItems) + 1,
		FileName:    job.FileName,
		RequestJSON: string(requestJSON),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := m.db.Create(&record).Error; err != nil {
		return HighlightClipExportJobResponse{}, err
	}

	m.mu.Lock()
	m.cleanupExpiredLocked(now, true)
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(id)
	return responseFromClipExportRecord(record), nil
}

func (m *highlightClipExportJobManager) get(analysisID, analystID uint, id string) (HighlightClipExportJobResponse, bool) {
	record, ok := m.getRecord(analysisID, analystID, id)
	if !ok {
		return HighlightClipExportJobResponse{}, false
	}
	m.normalizeRecordState(&record)
	return responseFromClipExportRecord(record), true
}

func (m *highlightClipExportJobManager) list(analysisID, analystID uint) []HighlightClipExportJobResponse {
	now := time.Now()
	m.cleanupExpired(now)

	var records []models.VideoClipExportJob
	if err := m.db.Where("analysis_id = ? AND analyst_id = ?", analysisID, analystID).
		Order("created_at DESC").
		Limit(10).
		Find(&records).Error; err != nil {
		return []HighlightClipExportJobResponse{}
	}

	responses := make([]HighlightClipExportJobResponse, 0, len(records))
	for i := range records {
		m.normalizeRecordState(&records[i])
		responses = append(responses, responseFromClipExportRecord(records[i]))
	}
	return responses
}

func (m *highlightClipExportJobManager) retry(analysis *models.VideoAnalysis, analystID uint, id string, req ExportHighlightClipsRequest, items []highlightClipExportItem) (HighlightClipExportJobResponse, bool, string, error) {
	now := time.Now()
	record, ok := m.getRecord(analysis.ID, analystID, id)
	if !ok {
		return HighlightClipExportJobResponse{}, false, "", nil
	}
	m.normalizeRecordState(&record)
	if record.Status != models.VideoClipExportFailed {
		return responseFromClipExportRecord(record), true, "只有失败的导出任务可以重试", nil
	}
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return HighlightClipExportJobResponse{}, true, "", err
	}
	if record.ZipPath != "" {
		_ = os.Remove(record.ZipPath)
	}

	copiedItems := append([]highlightClipExportItem(nil), items...)
	job := &highlightClipExportJob{
		ID:         id,
		AnalysisID: analysis.ID,
		AnalystID:  analystID,
		Analysis:   *analysis,
		Request:    req,
		Items:      copiedItems,
		Status:     highlightClipExportQueued,
		Progress:   0,
		Processed:  0,
		Total:      len(copiedItems) + 1,
		FileName:   record.FileName,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  now,
	}

	updates := map[string]interface{}{
		"status":       models.VideoClipExportQueued,
		"progress":     0,
		"processed":    0,
		"total":        len(copiedItems) + 1,
		"zip_path":     "",
		"request_json": string(requestJSON),
		"error":        "",
		"expires_at":   nil,
		"updated_at":   now,
	}
	if err := m.db.Model(&models.VideoClipExportJob{}).Where("job_id = ?", id).Updates(updates).Error; err != nil {
		return HighlightClipExportJobResponse{}, true, "", err
	}
	record.Status = models.VideoClipExportQueued
	record.Progress = 0
	record.Processed = 0
	record.Total = len(copiedItems) + 1
	record.ZipPath = ""
	record.RequestJSON = string(requestJSON)
	record.Error = ""
	record.ExpiresAt = nil
	record.UpdatedAt = now

	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(id)
	return responseFromClipExportRecord(record), true, "", nil
}

func (m *highlightClipExportJobManager) download(analysisID, analystID uint, id string) (string, string, HighlightClipExportJobResponse, bool, string) {
	record, ok := m.getRecord(analysisID, analystID, id)
	if !ok {
		return "", "", HighlightClipExportJobResponse{}, false, ""
	}
	m.normalizeRecordState(&record)
	if record.Status != models.VideoClipExportReady {
		return "", "", responseFromClipExportRecord(record), true, "下载包尚未生成"
	}

	if _, err := os.Stat(record.ZipPath); err != nil {
		m.markFailed(id, "下载包文件已过期，请重新生成")
		record.Status = models.VideoClipExportFailed
		record.Error = "下载包文件已过期，请重新生成"
		return "", "", responseFromClipExportRecord(record), true, "下载包文件已过期，请重新生成"
	}
	return record.ZipPath, record.FileName, responseFromClipExportRecord(record), true, ""
}

func (m *highlightClipExportJobManager) run(id string) {
	m.queue <- struct{}{}
	defer func() { <-m.queue }()

	m.mu.Lock()
	job, ok := m.jobs[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	job.Status = highlightClipExportProcessing
	job.Progress = 5
	job.UpdatedAt = time.Now()
	analysis := job.Analysis
	items := append([]highlightClipExportItem(nil), job.Items...)
	m.mu.Unlock()
	_ = m.db.Model(&models.VideoClipExportJob{}).Where("job_id = ?", id).Updates(map[string]interface{}{
		"status":     models.VideoClipExportProcessing,
		"progress":   5,
		"updated_at": time.Now(),
	}).Error

	zipPath, err := createHighlightClipsZipWithProgress(&analysis, items, func(processed, total int) {
		m.mu.Lock()
		if job, ok := m.jobs[id]; ok && job.Status == highlightClipExportProcessing {
			job.Processed = processed
			job.Total = total
			if total > 0 {
				job.Progress = (processed * 100) / total
				if job.Progress < 5 {
					job.Progress = 5
				}
				if job.Progress > 99 {
					job.Progress = 99
				}
			}
			job.UpdatedAt = time.Now()
		}
		m.mu.Unlock()
		_ = m.db.Model(&models.VideoClipExportJob{}).Where("job_id = ?", id).Updates(map[string]interface{}{
			"processed":  processed,
			"total":      total,
			"progress":   exportProgress(processed, total),
			"updated_at": time.Now(),
		}).Error
	})
	if err != nil {
		m.markFailed(id, "生成下载包失败: "+err.Error())
		return
	}

	now := time.Now()
	expiresAt := now.Add(m.ttl)
	_ = m.db.Model(&models.VideoClipExportJob{}).Where("job_id = ?", id).Updates(map[string]interface{}{
		"status":     models.VideoClipExportReady,
		"progress":   100,
		"processed":  len(items) + 1,
		"total":      len(items) + 1,
		"zip_path":   zipPath,
		"error":      "",
		"expires_at": &expiresAt,
		"updated_at": now,
	}).Error
	m.mu.Lock()
	if job, ok := m.jobs[id]; ok {
		if job.ZipPath != "" && job.ZipPath != zipPath {
			_ = os.Remove(job.ZipPath)
		}
		job.Status = highlightClipExportReady
		job.Progress = 100
		job.Processed = job.Total
		job.ZipPath = zipPath
		job.Error = ""
		job.ExpiresAt = &expiresAt
		job.UpdatedAt = now
	}
	delete(m.jobs, id)
	m.cleanupExpiredLocked(now, true)
	m.mu.Unlock()
}

func (m *highlightClipExportJobManager) markFailed(id string, message string) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	m.mu.Lock()
	if job, ok := m.jobs[id]; ok {
		if job.ZipPath != "" {
			_ = os.Remove(job.ZipPath)
			job.ZipPath = ""
		}
		job.Status = highlightClipExportFailed
		job.Progress = 100
		job.Error = message
		job.ExpiresAt = &expiresAt
		job.UpdatedAt = now
		delete(m.jobs, id)
	}
	m.cleanupExpiredLocked(now, true)
	m.mu.Unlock()
	_ = m.db.Model(&models.VideoClipExportJob{}).Where("job_id = ?", id).Updates(map[string]interface{}{
		"status":     models.VideoClipExportFailed,
		"progress":   100,
		"zip_path":   "",
		"error":      message,
		"expires_at": &expiresAt,
		"updated_at": now,
	}).Error
}

func (m *highlightClipExportJobManager) cleanupExpired(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupExpiredLocked(now, false)
}

func (m *highlightClipExportJobManager) cleanupExpiredLocked(now time.Time, dbOnly bool) {
	for id, job := range m.jobs {
		if job.ExpiresAt != nil && now.After(*job.ExpiresAt) {
			if job.ZipPath != "" {
				_ = os.Remove(job.ZipPath)
			}
			delete(m.jobs, id)
		}
	}
	var records []models.VideoClipExportJob
	if err := m.db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", models.VideoClipExportReady, now).Find(&records).Error; err == nil {
		for _, record := range records {
			if record.ZipPath != "" {
				_ = os.Remove(record.ZipPath)
			}
			updates := map[string]interface{}{
				"status":     models.VideoClipExportFailed,
				"progress":   100,
				"zip_path":   "",
				"error":      "下载包已过期，请重新生成",
				"updated_at": now,
			}
			_ = m.db.Model(&models.VideoClipExportJob{}).Where("id = ?", record.ID).Updates(updates).Error
		}
	}
	if !dbOnly {
		m.markInterruptedLocked(now)
	}
}

func (job *highlightClipExportJob) snapshotLocked() HighlightClipExportJobResponse {
	record := models.VideoClipExportJob{
		JobID:      job.ID,
		AnalysisID: job.AnalysisID,
		Status:     job.Status,
		Progress:   job.Progress,
		Processed:  job.Processed,
		Total:      job.Total,
		FileName:   job.FileName,
		Error:      job.Error,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		ExpiresAt:  job.ExpiresAt,
	}
	return responseFromClipExportRecord(record)
}

func (m *highlightClipExportJobManager) getRecord(analysisID, analystID uint, id string) (models.VideoClipExportJob, bool) {
	m.cleanupExpired(time.Now())
	var record models.VideoClipExportJob
	if err := m.db.Where("job_id = ? AND analysis_id = ? AND analyst_id = ?", id, analysisID, analystID).First(&record).Error; err != nil {
		return models.VideoClipExportJob{}, false
	}
	return record, true
}

func (m *highlightClipExportJobManager) normalizeRecordState(record *models.VideoClipExportJob) {
	if record.Status != models.VideoClipExportQueued && record.Status != models.VideoClipExportProcessing {
		return
	}
	m.mu.Lock()
	_, running := m.jobs[record.JobID]
	m.mu.Unlock()
	if running {
		return
	}
	now := time.Now()
	record.Status = models.VideoClipExportFailed
	record.Progress = 100
	record.Error = "导出任务已中断，请重试"
	record.UpdatedAt = now
	_ = m.db.Model(&models.VideoClipExportJob{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
		"status":     record.Status,
		"progress":   record.Progress,
		"error":      record.Error,
		"updated_at": now,
	}).Error
}

func (m *highlightClipExportJobManager) markInterruptedLocked(now time.Time) {
	var records []models.VideoClipExportJob
	if err := m.db.Where("status IN ?", []models.VideoClipExportJobStatus{
		models.VideoClipExportQueued,
		models.VideoClipExportProcessing,
	}).Find(&records).Error; err != nil {
		return
	}
	for _, record := range records {
		if _, ok := m.jobs[record.JobID]; ok {
			continue
		}
		_ = m.db.Model(&models.VideoClipExportJob{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
			"status":     models.VideoClipExportFailed,
			"progress":   100,
			"error":      "导出任务已中断，请重试",
			"updated_at": now,
		}).Error
	}
}

func responseFromClipExportRecord(record models.VideoClipExportJob) HighlightClipExportJobResponse {
	response := HighlightClipExportJobResponse{
		ID:         record.JobID,
		AnalysisID: record.AnalysisID,
		Status:     record.Status,
		Progress:   record.Progress,
		Processed:  record.Processed,
		Total:      record.Total,
		FileName:   record.FileName,
		Error:      record.Error,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
		ExpiresAt:  record.ExpiresAt,
	}
	if record.Status == models.VideoClipExportReady {
		response.DownloadURL = fmt.Sprintf("/api/video-analysis/%d/clips/export/jobs/%s/download", record.AnalysisID, record.JobID)
	}
	return response
}

func exportProgress(processed int, total int) int {
	if total <= 0 {
		return 5
	}
	progress := (processed * 100) / total
	if progress < 5 {
		return 5
	}
	if progress > 99 {
		return 99
	}
	return progress
}

// CreateHighlight 创建高光标记
func (ctrl *VideoAnalysisController) CreateHighlight(c *gin.Context) {
	var req CreateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if _, ok := ctrl.getOwnedAnalysisByID(c, req.AnalysisID); !ok {
		return
	}
	timestamp, mode, startTimeMs, endTimeMs, err := normalizeHighlightTiming(req.Timestamp, req.Mode, req.StartTimeMs, req.EndTimeMs)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的标记时间")
		return
	}
	markerType := req.MarkerType
	if markerType == "" {
		markerType = models.HighlightMarkerHighlight
	}
	includeInReport := true
	if req.IncludeInReport != nil {
		includeInReport = *req.IncludeInReport
	}

	highlight := &models.AnalysisHighlight{
		AnalysisID:      req.AnalysisID,
		Timestamp:       timestamp,
		MarkerType:      markerType,
		Mode:            mode,
		StartTimeMs:     startTimeMs,
		EndTimeMs:       endTimeMs,
		TagType:         req.TagType,
		Description:     req.Description,
		VideoClipURL:    req.VideoClipURL,
		IncludeInReport: includeInReport,
	}

	if err := ctrl.highlightRepo.Create(highlight); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建高光失败")
		return
	}
	if updatedHighlight := ctrl.syncHighlightClip(highlight.ID, mode); updatedHighlight != nil {
		highlight = updatedHighlight
	}

	utils.Success(c, "高光标记成功", highlight)
}

// UpdateHighlight 更新高光
func (ctrl *VideoAnalysisController) UpdateHighlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的高光ID")
		return
	}

	var req CreateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if _, _, ok := ctrl.getOwnedHighlight(c, uint(id)); !ok {
		return
	}
	timestamp, mode, startTimeMs, endTimeMs, err := normalizeHighlightTiming(req.Timestamp, req.Mode, req.StartTimeMs, req.EndTimeMs)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的标记时间")
		return
	}
	markerType := req.MarkerType
	if markerType == "" {
		markerType = models.HighlightMarkerHighlight
	}
	includeInReport := true
	if req.IncludeInReport != nil {
		includeInReport = *req.IncludeInReport
	}

	updates := map[string]interface{}{
		"timestamp":         timestamp,
		"marker_type":       markerType,
		"mode":              mode,
		"start_time_ms":     startTimeMs,
		"end_time_ms":       endTimeMs,
		"tag_type":          req.TagType,
		"description":       req.Description,
		"video_clip_url":    req.VideoClipURL,
		"include_in_report": includeInReport,
	}

	if err := ctrl.highlightRepo.Update(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新高光失败")
		return
	}
	updatedHighlight := ctrl.syncHighlightClip(uint(id), mode)
	if updatedHighlight == nil {
		updatedHighlight, err = ctrl.highlightRepo.FindByID(uint(id))
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "查询更新结果失败")
			return
		}
	}

	utils.Success(c, "更新成功", updatedHighlight)
}

// RetryHighlightClip 重新生成标记片段
func (ctrl *VideoAnalysisController) RetryHighlightClip(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的标记ID")
		return
	}

	highlight, _, ok := ctrl.getOwnedHighlight(c, uint(id))
	if !ok {
		return
	}
	if highlight.Mode != models.HighlightModeRange {
		utils.Error(c, http.StatusBadRequest, "单点标记不支持生成视频片段")
		return
	}
	updatedHighlight := ctrl.syncHighlightClip(highlight.ID, highlight.Mode)
	if updatedHighlight == nil {
		utils.Error(c, http.StatusInternalServerError, "创建剪辑任务失败")
		return
	}
	utils.Success(c, "剪辑任务已提交", updatedHighlight)
}

// GetHighlightClip 查询标记片段状态
func (ctrl *VideoAnalysisController) GetHighlightClip(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的标记ID")
		return
	}

	highlight, _, ok := ctrl.getOwnedHighlight(c, uint(id))
	if !ok {
		return
	}
	utils.Success(c, "", highlight)
}

// DownloadHighlightClip 下载单个标记片段
func (ctrl *VideoAnalysisController) DownloadHighlightClip(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的标记ID")
		return
	}

	highlight, _, ok := ctrl.getOwnedHighlight(c, uint(id))
	if !ok {
		return
	}
	if highlight.ClipStatus != models.HighlightClipReady || strings.TrimSpace(highlight.VideoClipURL) == "" {
		utils.Error(c, http.StatusBadRequest, "视频片段尚未生成")
		return
	}
	if ctrl.clipService == nil {
		utils.Error(c, http.StatusInternalServerError, "剪辑服务不可用")
		return
	}
	localPath, err := ctrl.clipService.ResolveClipFilePath(highlight.VideoClipURL)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}
	c.FileAttachment(localPath, filepath.Base(localPath))
}

type highlightClipExportItem struct {
	Highlight models.AnalysisHighlight
	LocalPath string
	FileName  string
}

func (ctrl *VideoAnalysisController) collectHighlightClipExportItems(analysis *models.VideoAnalysis, req ExportHighlightClipsRequest) ([]highlightClipExportItem, int, string) {
	if ctrl.clipService == nil {
		return nil, http.StatusInternalServerError, "剪辑服务不可用"
	}

	highlights, err := ctrl.highlightRepo.FindByAnalysisID(analysis.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, "查询标记失败"
	}

	selectedIDs := make(map[uint]bool, len(req.MarkerIDs))
	for _, markerID := range req.MarkerIDs {
		if markerID > 0 {
			selectedIDs[markerID] = true
		}
	}

	var (
		items           []highlightClipExportItem
		matchedIDs      int
		pendingCount    int
		failedCount     int
		brokenFileCount int
	)
	for _, highlight := range highlights {
		if len(selectedIDs) > 0 {
			if !selectedIDs[highlight.ID] {
				continue
			}
			matchedIDs++
		}
		if req.MarkerType != "" && highlight.MarkerType != req.MarkerType {
			continue
		}
		if req.TagType != "" && highlight.TagType != req.TagType {
			continue
		}
		if highlight.Mode != models.HighlightModeRange {
			continue
		}

		switch highlight.ClipStatus {
		case models.HighlightClipReady:
			if strings.TrimSpace(highlight.VideoClipURL) == "" {
				brokenFileCount++
				continue
			}
			localPath, err := ctrl.clipService.ResolveClipFilePath(highlight.VideoClipURL)
			if err != nil {
				brokenFileCount++
				continue
			}
			items = append(items, highlightClipExportItem{
				Highlight: highlight,
				LocalPath: localPath,
				FileName:  buildHighlightClipExportFileName(len(items)+1, highlight, localPath),
			})
		case models.HighlightClipFailed:
			failedCount++
		default:
			pendingCount++
		}
	}

	if len(selectedIDs) > 0 && matchedIDs != len(selectedIDs) {
		return nil, http.StatusBadRequest, "所选片段不属于当前分析"
	}
	if brokenFileCount > 0 {
		return nil, http.StatusConflict, "部分已生成片段文件缺失，请重试生成后下载"
	}
	if len(items) == 0 {
		return nil, http.StatusBadRequest, fmt.Sprintf("没有可下载的已生成片段（未完成 %d 个，失败 %d 个）", pendingCount, failedCount)
	}

	return items, 0, ""
}

// ExportHighlightClips 批量导出已生成的标记片段
func (ctrl *VideoAnalysisController) ExportHighlightClips(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req ExportHighlightClipsRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	items, status, message := ctrl.collectHighlightClipExportItems(analysis, req)
	if status != 0 {
		utils.Error(c, status, message)
		return
	}

	zipPath, err := createHighlightClipsZip(analysis, items)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成下载包失败: "+err.Error())
		return
	}
	defer os.Remove(zipPath)

	downloadName := buildHighlightClipsZipName(analysis)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(downloadName)))
	c.File(zipPath)
}

// CreateHighlightClipsExportJob 创建后台批量导出任务
func (ctrl *VideoAnalysisController) CreateHighlightClipsExportJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req ExportHighlightClipsRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}
	items, status, message := ctrl.collectHighlightClipExportItems(analysis, req)
	if status != 0 {
		utils.Error(c, status, message)
		return
	}

	job, err := ctrl.clipExportJobs.start(analysis, analysis.AnalystID, req, items)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建导出任务失败: "+err.Error())
		return
	}
	utils.Success(c, "导出任务已创建", job)
}

// ListHighlightClipsExportJobs 查询后台批量导出任务列表
func (ctrl *VideoAnalysisController) ListHighlightClipsExportJobs(c *gin.Context) {
	analysis, ok := ctrl.getOwnedAnalysisFromParam(c)
	if !ok {
		return
	}
	jobs := ctrl.clipExportJobs.list(analysis.ID, analysis.AnalystID)
	utils.Success(c, "", gin.H{"list": jobs})
}

// GetHighlightClipsExportJob 查询后台批量导出任务状态
func (ctrl *VideoAnalysisController) GetHighlightClipsExportJob(c *gin.Context) {
	analysis, ok := ctrl.getOwnedAnalysisFromParam(c)
	if !ok {
		return
	}
	job, exists := ctrl.clipExportJobs.get(analysis.ID, analysis.AnalystID, c.Param("job_id"))
	if !exists {
		utils.Error(c, http.StatusNotFound, "导出任务不存在")
		return
	}
	utils.Success(c, "", job)
}

// RetryHighlightClipsExportJob 重试失败的后台批量导出任务
func (ctrl *VideoAnalysisController) RetryHighlightClipsExportJob(c *gin.Context) {
	analysis, ok := ctrl.getOwnedAnalysisFromParam(c)
	if !ok {
		return
	}
	record, exists := ctrl.clipExportJobs.getRecord(analysis.ID, analysis.AnalystID, c.Param("job_id"))
	if !exists {
		utils.Error(c, http.StatusNotFound, "导出任务不存在")
		return
	}
	var req ExportHighlightClipsRequest
	if strings.TrimSpace(record.RequestJSON) != "" {
		if err := json.Unmarshal([]byte(record.RequestJSON), &req); err != nil {
			utils.Error(c, http.StatusInternalServerError, "导出任务参数损坏")
			return
		}
	}
	items, status, collectMessage := ctrl.collectHighlightClipExportItems(analysis, req)
	if status != 0 {
		utils.Error(c, status, collectMessage)
		return
	}
	job, exists, message, err := ctrl.clipExportJobs.retry(analysis, analysis.AnalystID, c.Param("job_id"), req, items)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "重试导出任务失败: "+err.Error())
		return
	}
	if !exists {
		utils.Error(c, http.StatusNotFound, "导出任务不存在")
		return
	}
	if message != "" {
		utils.Error(c, http.StatusConflict, message)
		return
	}
	utils.Success(c, "导出任务已重试", job)
}

// DownloadHighlightClipsExportJob 下载后台批量导出任务生成的 ZIP
func (ctrl *VideoAnalysisController) DownloadHighlightClipsExportJob(c *gin.Context) {
	analysis, ok := ctrl.getOwnedAnalysisFromParam(c)
	if !ok {
		return
	}
	zipPath, fileName, _, exists, message := ctrl.clipExportJobs.download(analysis.ID, analysis.AnalystID, c.Param("job_id"))
	if !exists {
		utils.Error(c, http.StatusNotFound, "导出任务不存在")
		return
	}
	if message != "" {
		utils.Error(c, http.StatusConflict, message)
		return
	}
	c.Header("Content-Type", "application/zip")
	c.FileAttachment(zipPath, fileName)
}

// DeleteHighlight 删除高光
func (ctrl *VideoAnalysisController) DeleteHighlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的高光ID")
		return
	}

	if _, _, ok := ctrl.getOwnedHighlight(c, uint(id)); !ok {
		return
	}

	if err := ctrl.highlightRepo.Delete(uint(id)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "删除高光失败")
		return
	}

	utils.Success(c, "删除成功", nil)
}

// GetHighlights 获取分析的所有高光
func (ctrl *VideoAnalysisController) GetHighlights(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	if _, ok := ctrl.getOwnedAnalysisByID(c, uint(id)); !ok {
		return
	}

	highlights, err := ctrl.highlightRepo.FindByAnalysisID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.Success(c, "", highlights)
}

func createHighlightClipsZip(analysis *models.VideoAnalysis, items []highlightClipExportItem) (string, error) {
	return createHighlightClipsZipWithProgress(analysis, items, nil)
}

func createHighlightClipsZipWithProgress(analysis *models.VideoAnalysis, items []highlightClipExportItem, onProgress func(processed int, total int)) (string, error) {
	tmpFile, err := os.CreateTemp("", "analysis-clips-*.zip")
	if err != nil {
		return "", err
	}
	zipPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(zipPath)
		}
	}()

	zipWriter := zip.NewWriter(tmpFile)
	total := len(items) + 1
	for index, item := range items {
		if err := addClipFileToZip(zipWriter, item); err != nil {
			_ = zipWriter.Close()
			_ = tmpFile.Close()
			return "", err
		}
		if onProgress != nil {
			onProgress(index+1, total)
		}
	}
	if err := addClipManifestToZip(zipWriter, items); err != nil {
		_ = zipWriter.Close()
		_ = tmpFile.Close()
		return "", err
	}
	if onProgress != nil {
		onProgress(total, total)
	}
	if err := zipWriter.Close(); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	cleanup = false
	_ = analysis
	return zipPath, nil
}

func addClipFileToZip(zipWriter *zip.Writer, item highlightClipExportItem) error {
	file, err := os.Open(item.LocalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = item.FileName
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}

func addClipManifestToZip(zipWriter *zip.Writer, items []highlightClipExportItem) error {
	writer, err := zipWriter.Create("markers_manifest.csv")
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{"序号", "类型", "标签", "开始时间", "结束时间", "描述", "是否纳入报告", "片段文件名"}); err != nil {
		return err
	}
	for index, item := range items {
		highlight := item.Highlight
		endTime := ""
		if highlight.EndTimeMs != nil {
			endTime = formatHighlightTimeMs(*highlight.EndTimeMs)
		}
		includeInReport := "否"
		if highlight.IncludeInReport {
			includeInReport = "是"
		}
		if err := csvWriter.Write([]string{
			strconv.Itoa(index + 1),
			markerTypeLabel(highlight.MarkerType),
			highlightTagLabel(highlight.TagType),
			formatHighlightTimeMs(highlight.StartTimeMs),
			endTime,
			highlight.Description,
			includeInReport,
			item.FileName,
		}); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func buildHighlightClipExportFileName(index int, highlight models.AnalysisHighlight, localPath string) string {
	ext := strings.TrimSpace(filepath.Ext(localPath))
	if ext == "" {
		ext = ".mp4"
	}
	endTime := highlight.StartTimeMs
	if highlight.EndTimeMs != nil {
		endTime = *highlight.EndTimeMs
	}
	return fmt.Sprintf(
		"%02d_%s_%s_%s-%s%s",
		index,
		safeArchiveNamePart(markerTypeLabel(highlight.MarkerType)),
		safeArchiveNamePart(highlightTagLabel(highlight.TagType)),
		formatExportTime(highlight.StartTimeMs),
		formatExportTime(endTime),
		ext,
	)
}

func buildHighlightClipsZipName(analysis *models.VideoAnalysis) string {
	playerName := safeArchiveNamePart(analysis.PlayerName)
	if playerName == "" {
		playerName = "未知球员"
	}
	return fmt.Sprintf(
		"少年球探_视频片段_%s_订单%d_%s.zip",
		playerName,
		analysis.OrderID,
		time.Now().Format("20060102"),
	)
}

func markerTypeLabel(markerType models.HighlightMarkerType) string {
	switch markerType {
	case models.HighlightMarkerIssue:
		return "待改进问题"
	case models.HighlightMarkerObservation:
		return "战术观察"
	case models.HighlightMarkerHighlight:
		return "精彩表现"
	default:
		return "精彩表现"
	}
}

func highlightTagLabel(tagType models.HighlightTagType) string {
	switch tagType {
	case models.HighlightGoal:
		return "进球"
	case models.HighlightAssist:
		return "助攻"
	case models.HighlightSteal:
		return "抢断"
	case models.HighlightSave:
		return "扑救"
	case models.HighlightDribble:
		return "过人"
	case models.HighlightPass:
		return "关键传球"
	case models.HighlightDefense:
		return "防守关键"
	case models.HighlightPositioningError:
		return "站位问题"
	case models.HighlightDecisionError:
		return "决策问题"
	case models.HighlightTurnover:
		return "失误"
	case models.HighlightRecoverySlow:
		return "回防不及时"
	case models.HighlightTacticalNote:
		return "战术观察"
	case models.HighlightOffBallRun:
		return "跑位亮点"
	default:
		return "未分类"
	}
}

func formatExportTime(ms int) string {
	if ms < 0 {
		ms = 0
	}
	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%dm%02ds", minutes, seconds)
}

func safeArchiveNamePart(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"\n", "_",
		"\r", "_",
		"\t", "_",
	)
	value = replacer.Replace(value)
	runes := []rune(value)
	if len(runes) > 36 {
		return string(runes[:36])
	}
	return value
}

// GenerateAIReportRequest AI报告生成请求
type GenerateAIReportRequest struct {
	AnalysisID uint `json:"analysis_id"`
}

// GenerateAIReport 触发AI生成报告（异步）
func (ctrl *VideoAnalysisController) GenerateAIReport(c *gin.Context) {
	var req GenerateAIReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, req.AnalysisID)
	if !ok {
		return
	}

	if analysis.OverallScore == 0 {
		utils.Error(c, http.StatusBadRequest, "请先完成评分")
		return
	}

	// 如果已经在生成中，直接返回
	if analysis.AIReportStatus == "generating" {
		utils.Success(c, "AI报告生成中，请耐心等待", nil)
		return
	}

	// 更新状态为生成中
	ctrl.analysisRepo.Update(req.AnalysisID, map[string]interface{}{
		"ai_report_status": "generating",
	})

	// 异步在后台生成报告，避免前端请求超时
	go func(analysisID uint, currentVersion int) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[AIReport] panic recovered for analysis %d: %v", analysisID, r)
				ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
					"ai_report_status": "failed",
				})
			}
		}()

		analysis, err := ctrl.analysisRepo.FindByID(analysisID)
		if err != nil || analysis == nil {
			ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
				"ai_report_status": "failed",
			})
			return
		}

		scores, _ := models.ParseScoresFromJSON(analysis.Scores)
		highlights, _ := ctrl.highlightRepo.FindIncludedInReport(analysis.ID)

		var highlightInputs []services.HighlightInput
		for _, h := range highlights {
			endTime := ""
			if h.EndTimeMs != nil {
				endTime = formatHighlightTimeMs(*h.EndTimeMs)
			}
			highlightInputs = append(highlightInputs, services.HighlightInput{
				Timestamp:   h.Timestamp,
				MarkerType:  string(h.MarkerType),
				Mode:        string(h.Mode),
				StartTime:   formatHighlightTimeMs(h.StartTimeMs),
				EndTime:     endTime,
				TagType:     string(h.TagType),
				Description: h.Description,
			})
		}

		reportInput := &services.VideoAnalysisReportInput{
			PlayerName:     analysis.PlayerName,
			PlayerAge:      analysis.PlayerAge,
			PlayerPosition: analysis.PlayerPosition,
			PlayerFoot:     analysis.PlayerFoot,
			PlayerHeight:   analysis.PlayerHeight,
			PlayerWeight:   analysis.PlayerWeight,
			PlayerTeam:     analysis.PlayerTeam,
			MatchName:      analysis.MatchName,
			MatchDate:      analysis.MatchDate,
			MatchType:      analysis.MatchType,
			OpponentLevel:  analysis.OpponentLevel,
			Opponent:       analysis.Opponent,
			PlayTime:       analysis.PlayTime,
			Goals:          analysis.Goals,
			Assists:        analysis.Assists,
			OverallScore:   analysis.OverallScore,
			PotentialLevel: string(analysis.PotentialLevel),
			Highlights:     highlightInputs,
			Summary:        analysis.Summary,
			Improvements:   analysis.Improvements,
			AnalystNotes:   analysis.AnalystNotes,
		}

		scoresInput := services.ScoresInput{
			BallControl:          services.ScoreInput{Score: scores.BallControl.Score, Weight: scores.BallControl.Weight, Comment: scores.BallControl.Comment},
			OffBallMovement:      services.ScoreInput{Score: scores.OffBallMovement.Score, Weight: scores.OffBallMovement.Weight, Comment: scores.OffBallMovement.Comment},
			PressingAwareness:    services.ScoreInput{Score: scores.PressingAwareness.Score, Weight: scores.PressingAwareness.Weight, Comment: scores.PressingAwareness.Comment},
			Positioning:          services.ScoreInput{Score: scores.Positioning.Score, Weight: scores.Positioning.Weight, Comment: scores.Positioning.Comment},
			WidthParticipation:   services.ScoreInput{Score: scores.WidthParticipation.Score, Weight: scores.WidthParticipation.Weight, Comment: scores.WidthParticipation.Comment},
			OffBallSupport:       services.ScoreInput{Score: scores.OffBallSupport.Score, Weight: scores.OffBallSupport.Weight, Comment: scores.OffBallSupport.Comment},
			OneVOne:              services.ScoreInput{Score: scores.OneVOne.Score, Weight: scores.OneVOne.Weight, Comment: scores.OneVOne.Comment},
			CrossingAssist:       services.ScoreInput{Score: scores.CrossingAssist.Score, Weight: scores.CrossingAssist.Weight, Comment: scores.CrossingAssist.Comment},
			CombatAbility:        services.ScoreInput{Score: scores.CombatAbility.Score, Weight: scores.CombatAbility.Weight, Comment: scores.CombatAbility.Comment},
			PaceRhythm:           services.ScoreInput{Score: scores.PaceRhythm.Score, Weight: scores.PaceRhythm.Weight, Comment: scores.PaceRhythm.Comment},
			PassVision:           services.ScoreInput{Score: scores.PassVision.Score, Weight: scores.PassVision.Weight, Comment: scores.PassVision.Comment},
			BodyPosture:          services.ScoreInput{Score: scores.BodyPosture.Score, Weight: scores.BodyPosture.Weight, Comment: scores.BodyPosture.Comment},
			DefensiveCommitment:  services.ScoreInput{Score: scores.DefensiveCommitment.Score, Weight: scores.DefensiveCommitment.Weight, Comment: scores.DefensiveCommitment.Comment},
			LossRecovery:         services.ScoreInput{Score: scores.LossRecovery.Score, Weight: scores.LossRecovery.Weight, Comment: scores.LossRecovery.Comment},
			TeammateCoordination: services.ScoreInput{Score: scores.TeammateCoordination.Score, Weight: scores.TeammateCoordination.Weight, Comment: scores.TeammateCoordination.Comment},
			SecondBall:           services.ScoreInput{Score: scores.SecondBall.Score, Weight: scores.SecondBall.Weight, Comment: scores.SecondBall.Comment},
			AerialDuel:           services.ScoreInput{Score: scores.AerialDuel.Score, Weight: scores.AerialDuel.Weight, Comment: scores.AerialDuel.Comment},
			DefensiveShape:       services.ScoreInput{Score: scores.DefensiveShape.Score, Weight: scores.DefensiveShape.Weight, Comment: scores.DefensiveShape.Comment},
			RoleAdjustment:       services.ScoreInput{Score: scores.RoleAdjustment.Score, Weight: scores.RoleAdjustment.Weight, Comment: scores.RoleAdjustment.Comment},
			DefensiveRhythm:      services.ScoreInput{Score: scores.DefensiveRhythm.Score, Weight: scores.DefensiveRhythm.Weight, Comment: scores.DefensiveRhythm.Comment},
		}
		reportInput.Scores = scoresInput

		prompt := services.BuildReportPrompt(reportInput)
		aiReport, err := ctrl.aiService.GenerateReport(prompt)
		if err != nil {
			log.Printf("[AIReport] generation failed for analysis %d: %v", analysisID, err)
			ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
				"ai_report_status": "failed",
			})
			return
		}

		newVersion := currentVersion + 1
		ctrl.analysisRepo.Update(analysisID, map[string]interface{}{
			"ai_report":         aiReport,
			"ai_report_status":  "draft",
			"ai_report_version": newVersion,
		})
	}(analysis.ID, analysis.AIReportVersion)

	utils.Success(c, "AI报告生成任务已提交，预计需要3-5分钟", gin.H{
		"status": "generating",
	})
}

// UpdateAIReport 手动修改AI报告
func (ctrl *VideoAnalysisController) UpdateAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	var req struct {
		Report string `json:"report"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	updates := map[string]interface{}{
		"ai_report":        req.Report,
		"ai_report_status": "draft",
	}

	err = ctrl.analysisRepo.Update(uint(id), updates)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新报告失败")
		return
	}

	utils.Success(c, "报告已保存", nil)
}

// ConfirmAIReport 确认AI报告并提交管理员审核
// 核心操作：1.更新video_analyses状态 2.创建/更新待审核reports记录
func (ctrl *VideoAnalysisController) ConfirmAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	// 1. 查询分析记录并校验归属
	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	// 校验：AI报告必须已生成
	if analysis.AIReport == "" {
		utils.Error(c, http.StatusBadRequest, "请先生成AI报告")
		return
	}

	// 2. 更新 video_analyses 状态
	updates := map[string]interface{}{
		"ai_report_status": "confirmed",
		"status":           models.AnalysisStatusSubmitted,
	}
	if err := ctrl.analysisRepo.Update(uint(id), updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "确认报告失败")
		return
	}

	// 3. 创建/更新 reports 记录（桥接 video_analyses → reports），等待管理员审核
	reportRepo := models.NewReportRepository(ctrl.db)
	existingReport, _ := reportRepo.FindByOrderID(analysis.OrderID)
	var reportID uint

	if existingReport != nil {
		reportID = existingReport.ID
		if err := reportRepo.Update(existingReport.ID, map[string]interface{}{
			"content":        analysis.AIReport,
			"status":         models.ReportStatusProcessing,
			"overall_rating": analysis.OverallScore,
			"potential":      string(analysis.PotentialLevel),
			"summary":        analysis.Summary,
			"suggestions":    analysis.Improvements,
			"rating_details": analysis.Scores,
			"review_remark":  "",
		}); err != nil {
			utils.Error(c, http.StatusInternalServerError, "提交审核失败")
			return
		}
	} else {
		report := &models.Report{
			OrderID:        analysis.OrderID,
			UserID:         analysis.UserID,
			AnalystID:      analysis.AnalystID,
			PlayerName:     analysis.PlayerName,
			PlayerPosition: analysis.PlayerPosition,
			Content:        analysis.AIReport,
			Status:         models.ReportStatusProcessing,
			OverallRating:  analysis.OverallScore,
			Potential:      string(analysis.PotentialLevel),
			Summary:        analysis.Summary,
			Suggestions:    analysis.Improvements,
			RatingDetails:  analysis.Scores,
		}
		if err := reportRepo.Create(report); err != nil {
			utils.Error(c, http.StatusInternalServerError, "提交审核失败")
			return
		}
		reportID = report.ID
	}

	if err := ctrl.db.Model(&models.Order{}).Where("id = ?", analysis.OrderID).
		Update("report_id", reportID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "关联报告失败")
		return
	}

	// 生成 MD 文档
	if ctrl.reportGen != nil {
		var user models.User
		_ = ctrl.db.First(&user, analysis.UserID).Error
		var analyst models.Analyst
		_ = ctrl.db.First(&analyst, analysis.AnalystID).Error
		ratingMD, playerMD, _ := ctrl.reportGen.GenerateFromVideoAnalysis(analysis, analyst.Name, &user)
		if ratingMD != "" {
			ctrl.analysisRepo.Update(analysis.ID, map[string]interface{}{
				"rating_report_md": ratingMD,
				"player_info_md":   playerMD,
			})
		}
	}

	ctrl.notifyAdminsReportSubmitted(reportID, analysis.PlayerName)

	utils.Success(c, "报告已提交审核，文档已生成", gin.H{
		"order_id":    analysis.OrderID,
		"analysis_id": id,
		"report_id":   reportID,
	})
}

// GetAIReport 获取AI报告
func (ctrl *VideoAnalysisController) GetAIReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, ok := ctrl.getOwnedAnalysisByID(c, uint(id))
	if !ok {
		return
	}

	utils.Success(c, "", gin.H{
		"report":  analysis.AIReport,
		"status":  analysis.AIReportStatus,
		"version": analysis.AIReportVersion,
	})
}

// CreateAnalysisFromOrderRequest 创建分析请求
type CreateAnalysisFromOrderRequest struct {
	OrderID uint `json:"order_id"`
}

// CreateFromOrder 从订单创建分析记录
func (ctrl *VideoAnalysisController) CreateFromOrder(c *gin.Context) {
	var req CreateAnalysisFromOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 查询订单信息，补全 analyst_id 和 user_id
	order, ok := ctrl.getOwnedOrder(c, req.OrderID)
	if !ok {
		return
	}

	existing, _ := ctrl.analysisRepo.FindByOrderID(req.OrderID)
	if existing != nil {
		utils.Error(c, http.StatusBadRequest, "该订单已有分析记录")
		return
	}

	analysis := &models.VideoAnalysis{
		OrderID:        req.OrderID,
		AnalystID:      *order.AnalystID,
		UserID:         order.UserID,
		PlayerName:     order.PlayerName,
		PlayerAge:      order.PlayerAge,
		PlayerPosition: order.PlayerPosition,
		MatchName:      order.MatchName,
		Opponent:       order.Opponent,
		Status:         models.AnalysisStatusScoring,
	}

	err := ctrl.analysisRepo.Create(analysis)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}

	utils.Success(c, "分析记录已创建", analysis)
}

// ========== 球员端 API ==========

// GetMyAnalyses 获取当前用户的视频分析列表（球员视角）
func (ctrl *VideoAnalysisController) GetMyAnalyses(c *gin.Context) {
	userID := c.GetUint("userId")

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	analyses, total, err := ctrl.analysisRepo.FindByUserID(userID, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":      analyses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMyAnalysisDetail 获取当前用户的某条视频分析详情（含评分+报告）
func (ctrl *VideoAnalysisController) GetMyAnalysisDetail(c *gin.Context) {
	userID := c.GetUint("userId")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的分析ID")
		return
	}

	analysis, err := ctrl.analysisRepo.FindByID(uint(id))
	if err != nil || analysis == nil {
		utils.Error(c, http.StatusNotFound, "分析记录不存在")
		return
	}

	// 权限校验：只能查看自己的分析
	if analysis.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权查看此分析")
		return
	}
	if analysis.Status != models.AnalysisStatusCompleted {
		utils.Error(c, http.StatusBadRequest, "报告尚未审核完成")
		return
	}

	scores, _ := models.ParseScoresFromJSON(analysis.Scores)
	highlights, _ := ctrl.highlightRepo.FindByAnalysisID(analysis.ID)

	utils.Success(c, "", gin.H{
		"analysis":          analysis,
		"scores":            scores,
		"highlights":        highlights,
		"ai_report":         analysis.AIReport,
		"ai_report_status":  analysis.AIReportStatus,
		"ai_report_version": analysis.AIReportVersion,
	})
}
