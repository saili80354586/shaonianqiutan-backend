package services

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

type AdminOrderAnalysisProgressDTO struct {
	Stage      string                           `json:"stage"`
	StageLabel string                           `json:"stage_label"`
	Percent    int                              `json:"percent"`
	RiskLevel  string                           `json:"risk_level"`
	RiskLabel  string                           `json:"risk_label"`
	SLA        AdminOrderProgressSLADTO         `json:"sla"`
	Assignment *AdminOrderProgressAssignmentDTO `json:"assignment,omitempty"`
	Analysis   *AdminOrderProgressAnalysisDTO   `json:"analysis,omitempty"`
	Summary    AdminOrderProgressSummaryDTO     `json:"summary"`
	Exception  *AdminOrderProgressExceptionDTO  `json:"exception,omitempty"`
}

type AdminOrderWithProgressDTO struct {
	models.Order
	AnalysisProgress *AdminOrderAnalysisProgressDTO `json:"analysis_progress,omitempty"`
}

type AdminOrderProgressSLADTO struct {
	Deadline         *time.Time `json:"deadline,omitempty"`
	RemainingSeconds int64      `json:"remaining_seconds"`
	IsNearDeadline   bool       `json:"is_near_deadline"`
	IsOverdue        bool       `json:"is_overdue"`
	Label            string     `json:"label"`
}

type AdminOrderProgressAssignmentDTO struct {
	Status         string     `json:"status"`
	AssignedAt     *time.Time `json:"assigned_at,omitempty"`
	RespondedAt    *time.Time `json:"responded_at,omitempty"`
	RejectedReason string     `json:"rejected_reason,omitempty"`
}

type AdminOrderProgressAnalysisDTO struct {
	ID              uint       `json:"id"`
	Status          string     `json:"status"`
	AIReportStatus  string     `json:"ai_report_status"`
	AIReportVersion int        `json:"ai_report_version"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type AdminOrderProgressSummaryDTO struct {
	ScoreCompleted         int  `json:"score_completed"`
	ScoreTotal             int  `json:"score_total"`
	ScoreCommentCompleted  int  `json:"score_comment_completed"`
	ScoreValueChanged      int  `json:"score_value_changed"`
	TextSectionsCompleted  int  `json:"text_sections_completed"`
	TextSectionsTotal      int  `json:"text_sections_total"`
	HighlightCount         int  `json:"highlight_count"`
	IncludedHighlightCount int  `json:"included_highlight_count"`
	ClipReadyCount         int  `json:"clip_ready_count"`
	ClipFailedCount        int  `json:"clip_failed_count"`
	HasSummary             bool `json:"has_summary"`
	HasReport              bool `json:"has_report"`
}

type AdminOrderProgressExceptionDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AdminOrderAnalysisProgressDetailDTO struct {
	Order            *models.Order                    `json:"order"`
	Analyst          *models.Analyst                  `json:"analyst,omitempty"`
	AnalysisProgress AdminOrderAnalysisProgressDTO    `json:"analysis_progress"`
	Timeline         []AdminProgressTimelineItemDTO   `json:"timeline"`
	Completion       AdminOrderAnalysisCompletionDTO  `json:"completion"`
	OperationEvents  []AdminAnalysisOperationEventDTO `json:"operation_events"`
	Exceptions       []AdminOrderProgressExceptionDTO `json:"exceptions"`
}

type AdminProgressTimelineItemDTO struct {
	Key    string     `json:"key"`
	Label  string     `json:"label"`
	Status string     `json:"status"`
	Time   *time.Time `json:"time,omitempty"`
	Actor  string     `json:"actor,omitempty"`
	Remark string     `json:"remark,omitempty"`
}

type AdminOrderAnalysisCompletionDTO struct {
	ScoreOverview     AdminScoreOverviewDTO     `json:"score_overview"`
	ScoreGroups       []AdminScoreGroupDTO      `json:"score_groups"`
	TextSections      []AdminTextSectionDTO     `json:"text_sections"`
	HighlightOverview AdminHighlightOverviewDTO `json:"highlight_overview"`
	HighlightItems    []AdminHighlightItemDTO   `json:"highlight_items"`
	ReportOverview    AdminReportOverviewDTO    `json:"report_overview"`
}

type AdminScoreOverviewDTO struct {
	ScoreTotal        int        `json:"score_total"`
	CompletedCount    int        `json:"completed_count"`
	ScoreOnlyCount    int        `json:"score_only_count"`
	CommentOnlyCount  int        `json:"comment_only_count"`
	NotStartedCount   int        `json:"not_started_count"`
	CommentTotalWords int        `json:"comment_total_words"`
	LastSavedAt       *time.Time `json:"last_saved_at,omitempty"`
}

type AdminScoreGroupDTO struct {
	GroupKey       string                   `json:"group_key"`
	GroupLabel     string                   `json:"group_label"`
	CompletedCount int                      `json:"completed_count"`
	Total          int                      `json:"total"`
	Items          []AdminScoreDimensionDTO `json:"items"`
}

type AdminScoreDimensionDTO struct {
	FieldKey      string     `json:"field_key"`
	FieldLabel    string     `json:"field_label"`
	Score         float64    `json:"score"`
	CommentWords  int        `json:"comment_words"`
	Status        string     `json:"status"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	UpdateCount   int        `json:"update_count,omitempty"`
}

type AdminTextSectionDTO struct {
	FieldKey      string     `json:"field_key"`
	FieldLabel    string     `json:"field_label"`
	Filled        bool       `json:"filled"`
	WordCount     int        `json:"word_count"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	Preview       string     `json:"preview,omitempty"`
}

type AdminHighlightOverviewDTO struct {
	HighlightCount         int `json:"highlight_count"`
	IncludedHighlightCount int `json:"included_highlight_count"`
	RangeCount             int `json:"range_count"`
	PointCount             int `json:"point_count"`
	ClipReadyCount         int `json:"clip_ready_count"`
	ClipProcessingCount    int `json:"clip_processing_count"`
	ClipFailedCount        int `json:"clip_failed_count"`
}

type AdminHighlightItemDTO struct {
	ID              uint       `json:"id"`
	Timestamp       string     `json:"timestamp"`
	MarkerType      string     `json:"marker_type"`
	Mode            string     `json:"mode"`
	StartTimeMs     int        `json:"start_time_ms"`
	EndTimeMs       *int       `json:"end_time_ms,omitempty"`
	TagType         string     `json:"tag_type"`
	Description     string     `json:"description"`
	VideoClipURL    string     `json:"video_clip_url,omitempty"`
	ClipStatus      string     `json:"clip_status"`
	ClipError       string     `json:"clip_error,omitempty"`
	ClipVersion     int        `json:"clip_version"`
	ClipGeneratedAt *time.Time `json:"clip_generated_at,omitempty"`
	IncludeInReport bool       `json:"include_in_report"`
	SortOrder       int        `json:"sort_order"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AdminReportOverviewDTO struct {
	AIReportStatus  string     `json:"ai_report_status"`
	AIReportVersion int        `json:"ai_report_version"`
	TemplateVersion string     `json:"template_version,omitempty"`
	InputSnapshot   bool       `json:"input_snapshot"`
	WordReportReady bool       `json:"word_report_ready"`
	PDFReportReady  bool       `json:"pdf_report_ready"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	ReviewStatus    string     `json:"review_status,omitempty"`
	ReviewRemark    string     `json:"review_remark,omitempty"`
}

type AdminAnalysisOperationEventDTO struct {
	EventType     string    `json:"event_type"`
	Section       string    `json:"section"`
	FieldKey      string    `json:"field_key,omitempty"`
	FieldLabel    string    `json:"field_label,omitempty"`
	Summary       string    `json:"summary"`
	BeforeSummary string    `json:"before_summary,omitempty"`
	AfterSummary  string    `json:"after_summary,omitempty"`
	Metadata      string    `json:"metadata,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type operationEventStats struct {
	lastUpdatedAt *time.Time
	count         int
}

func (s *AdminService) GetOrderAnalysisProgressDetail(orderID uint) (*AdminOrderAnalysisProgressDetailDTO, error) {
	db := s.GetDB()
	if db == nil {
		return nil, errors.New("数据库未初始化")
	}

	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, gorm.ErrRecordNotFound
	}

	assignment := findLatestOrderAssignment(db, order.ID)
	analysis := findVideoAnalysisByOrder(db, order.ID)
	report := findReportByOrder(db, order.ID)
	histories, _ := s.GetOrderStatusHistory(order.ID)
	events := findAnalysisOperationEvents(db, order.ID, 200)
	highlights := findAnalysisHighlights(db, analysis)

	eventStats := buildOperationEventStats(events)
	scoreOverview, scoreGroups := buildAdminScoreProgress(analysis, eventStats)
	textSections := buildAdminTextSections(analysis, eventStats)
	highlightOverview := buildAdminHighlightOverview(highlights)
	highlightItems := buildAdminHighlightItems(highlights)
	reportOverview := buildAdminReportOverview(analysis, report, events)
	completion := AdminOrderAnalysisCompletionDTO{
		ScoreOverview:     scoreOverview,
		ScoreGroups:       scoreGroups,
		TextSections:      textSections,
		HighlightOverview: highlightOverview,
		HighlightItems:    highlightItems,
		ReportOverview:    reportOverview,
	}

	progress := buildAdminOrderProgress(order, assignment, analysis, report, completion)
	exceptions := buildAdminOrderProgressExceptions(order, assignment, analysis, report, progress, events)
	progress.Exception = firstAdminOrderProgressException(exceptions)
	timeline := buildAdminOrderProgressTimeline(order, assignment, analysis, report, histories, events)

	return &AdminOrderAnalysisProgressDetailDTO{
		Order:            order,
		Analyst:          order.Analyst,
		AnalysisProgress: progress,
		Timeline:         timeline,
		Completion:       completion,
		OperationEvents:  toAdminAnalysisOperationEventDTOs(events),
		Exceptions:       exceptions,
	}, nil
}

func (s *AdminService) GetAllOrdersWithProgress(page, pageSize int, status string) ([]AdminOrderWithProgressDTO, int64, error) {
	orders, total, err := s.GetAllOrders(page, pageSize, status)
	if err != nil {
		return nil, 0, err
	}
	db := s.GetDB()
	if db == nil {
		return nil, 0, errors.New("数据库未初始化")
	}

	result := make([]AdminOrderWithProgressDTO, 0, len(orders))
	for i := range orders {
		order := orders[i]
		assignment := findLatestOrderAssignment(db, order.ID)
		analysis := findVideoAnalysisByOrder(db, order.ID)
		report := findReportByOrder(db, order.ID)
		events := findAnalysisOperationEvents(db, order.ID, 100)
		highlights := findAnalysisHighlights(db, analysis)
		eventStats := buildOperationEventStats(events)
		scoreOverview, scoreGroups := buildAdminScoreProgress(analysis, eventStats)
		textSections := buildAdminTextSections(analysis, eventStats)
		highlightOverview := buildAdminHighlightOverview(highlights)
		reportOverview := buildAdminReportOverview(analysis, report, events)
		completion := AdminOrderAnalysisCompletionDTO{
			ScoreOverview:     scoreOverview,
			ScoreGroups:       scoreGroups,
			TextSections:      textSections,
			HighlightOverview: highlightOverview,
			ReportOverview:    reportOverview,
		}
		progress := buildAdminOrderProgress(&order, assignment, analysis, report, completion)
		exceptions := buildAdminOrderProgressExceptions(&order, assignment, analysis, report, progress, events)
		progress.Exception = firstAdminOrderProgressException(exceptions)
		result = append(result, AdminOrderWithProgressDTO{
			Order:            order,
			AnalysisProgress: &progress,
		})
	}
	return result, total, nil
}

func (s *AdminService) SendOrderProgressReminder(orderID, adminID uint, message string) error {
	db := s.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return gorm.ErrRecordNotFound
	}
	if order.AnalystID == nil || *order.AnalystID == 0 {
		return errors.New("订单尚未分配分析师")
	}
	analyst, err := s.analystRepo.FindByID(*order.AnalystID)
	if err != nil {
		return err
	}
	if analyst == nil || analyst.UserID == 0 {
		return errors.New("分析师不存在或未绑定用户")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = "管理员提醒你尽快更新订单 " + order.OrderNo + " 的分析进度"
	}
	if s.notificationService != nil {
		if _, err := s.notificationService.CreateNotification(analyst.UserID, models.NotificationTypeOrder, "订单进度提醒", message, &models.NotificationData{
			TargetType: "order",
			TargetID:   order.ID,
			Link:       "/analyst/dashboard",
			Extra: map[string]interface{}{
				"order_no": order.OrderNo,
				"admin_id": adminID,
			},
		}); err != nil {
			return err
		}
	}
	analysis := findVideoAnalysisByOrder(db, order.ID)
	return createAdminOrderProgressControlEvent(db, order, analysis, *order.AnalystID, adminID, "admin_reminder_sent", "reminder", "管理员催办", message)
}

func (s *AdminService) MarkOrderProgressException(orderID, adminID uint, code, message string) error {
	db := s.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return gorm.ErrRecordNotFound
	}
	code = strings.TrimSpace(code)
	if code == "" {
		code = "admin_marked_exception"
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return errors.New("异常说明不能为空")
	}
	analysis := findVideoAnalysisByOrder(db, order.ID)
	analystID := uint(0)
	if order.AnalystID != nil {
		analystID = *order.AnalystID
	}
	if err := createAdminOrderProgressControlEvent(db, order, analysis, analystID, adminID, "admin_exception_marked", code, "管理员标记异常", message); err != nil {
		return err
	}
	if s.notificationService != nil && analystID != 0 {
		if analyst, err := s.analystRepo.FindByID(analystID); err == nil && analyst != nil && analyst.UserID != 0 {
			_, _ = s.notificationService.CreateNotification(analyst.UserID, models.NotificationTypeOrder, "订单异常提醒", message, &models.NotificationData{
				TargetType: "order",
				TargetID:   order.ID,
				Link:       "/analyst/dashboard",
				Extra: map[string]interface{}{
					"order_no":       order.OrderNo,
					"exception_code": code,
					"admin_id":       adminID,
				},
			})
		}
	}
	return nil
}

func (s *AdminService) ResolveOrderProgressException(orderID, adminID uint, code, message string) error {
	db := s.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return gorm.ErrRecordNotFound
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("异常编码不能为空")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = "管理员已解除异常标记"
	}
	analysis := findVideoAnalysisByOrder(db, order.ID)
	analystID := uint(0)
	if order.AnalystID != nil {
		analystID = *order.AnalystID
	}
	return createAdminOrderProgressControlEvent(db, order, analysis, analystID, adminID, "admin_exception_resolved", code, "管理员解除异常", message)
}

func createAdminOrderProgressControlEvent(db *gorm.DB, order *models.Order, analysis *models.VideoAnalysis, analystID uint, adminID uint, eventType, fieldKey, fieldLabel, summary string) error {
	analysisID := uint(0)
	if analysis != nil {
		analysisID = analysis.ID
		if analystID == 0 {
			analystID = analysis.AnalystID
		}
	}
	return models.NewAnalysisOperationEventRepository(db).Create(&models.AnalysisOperationEvent{
		OrderID:      order.ID,
		AnalysisID:   analysisID,
		AnalystID:    analystID,
		EventType:    eventType,
		Section:      "admin_control",
		FieldKey:     fieldKey,
		FieldLabel:   fieldLabel,
		AfterSummary: summary,
		Metadata: operationMetadata(map[string]interface{}{
			"admin_id": adminID,
			"order_no": order.OrderNo,
		}),
		CreatedAt: time.Now(),
	})
}

func findLatestOrderAssignment(db *gorm.DB, orderID uint) *models.OrderAssignment {
	var assignment models.OrderAssignment
	if err := db.Where("order_id = ?", orderID).Order("assigned_at DESC, id DESC").First(&assignment).Error; err != nil {
		return nil
	}
	return &assignment
}

func findVideoAnalysisByOrder(db *gorm.DB, orderID uint) *models.VideoAnalysis {
	var analysis models.VideoAnalysis
	if err := db.Where("order_id = ?", orderID).First(&analysis).Error; err != nil {
		return nil
	}
	return &analysis
}

func findReportByOrder(db *gorm.DB, orderID uint) *models.Report {
	var report models.Report
	if err := db.Where("order_id = ?", orderID).Order("updated_at DESC, id DESC").First(&report).Error; err != nil {
		return nil
	}
	return &report
}

func findAnalysisOperationEvents(db *gorm.DB, orderID uint, limit int) []models.AnalysisOperationEvent {
	events, err := models.NewAnalysisOperationEventRepository(db).FindByOrderID(orderID, limit)
	if err != nil {
		return []models.AnalysisOperationEvent{}
	}
	return events
}

func findAnalysisHighlights(db *gorm.DB, analysis *models.VideoAnalysis) []models.AnalysisHighlight {
	if analysis == nil {
		return []models.AnalysisHighlight{}
	}
	var highlights []models.AnalysisHighlight
	_ = db.Where("analysis_id = ?", analysis.ID).Order("sort_order ASC, start_time_ms ASC, id ASC").Find(&highlights).Error
	return highlights
}

func buildOperationEventStats(events []models.AnalysisOperationEvent) map[string]operationEventStats {
	stats := map[string]operationEventStats{}
	for _, event := range events {
		if event.FieldKey == "" {
			continue
		}
		current := stats[event.FieldKey]
		current.count++
		createdAt := event.CreatedAt
		if current.lastUpdatedAt == nil || createdAt.After(*current.lastUpdatedAt) {
			current.lastUpdatedAt = &createdAt
		}
		stats[event.FieldKey] = current
	}
	return stats
}

func buildAdminScoreProgress(analysis *models.VideoAnalysis, eventStats map[string]operationEventStats) (AdminScoreOverviewDTO, []AdminScoreGroupDTO) {
	var scores *models.VideoAnalysisScores
	if analysis != nil {
		parsed, err := models.ParseScoresFromJSON(analysis.Scores)
		if err == nil {
			scores = parsed
		}
	}
	details := models.ScoreDimensionDetails(scores)
	groupIndex := map[string]int{}
	groups := []AdminScoreGroupDTO{}
	overview := AdminScoreOverviewDTO{ScoreTotal: len(details)}

	for _, detail := range details {
		commentWords := runeCount(detail.Comment)
		scoreChanged := math.Abs(detail.Score-7.0) > 0.001
		status := "not_started"
		switch {
		case scoreChanged && commentWords > 0:
			status = "completed"
			overview.CompletedCount++
		case scoreChanged:
			status = "score_only"
			overview.ScoreOnlyCount++
		case commentWords > 0:
			status = "comment_only"
			overview.CommentOnlyCount++
		default:
			overview.NotStartedCount++
		}
		overview.CommentTotalWords += commentWords

		stats := eventStats[detail.FieldKey]
		if stats.lastUpdatedAt != nil && (overview.LastSavedAt == nil || stats.lastUpdatedAt.After(*overview.LastSavedAt)) {
			overview.LastSavedAt = stats.lastUpdatedAt
		}

		item := AdminScoreDimensionDTO{
			FieldKey:      detail.FieldKey,
			FieldLabel:    detail.FieldLabel,
			Score:         detail.Score,
			CommentWords:  commentWords,
			Status:        status,
			LastUpdatedAt: stats.lastUpdatedAt,
			UpdateCount:   stats.count,
		}

		idx, exists := groupIndex[detail.GroupKey]
		if !exists {
			idx = len(groups)
			groupIndex[detail.GroupKey] = idx
			groups = append(groups, AdminScoreGroupDTO{
				GroupKey:   detail.GroupKey,
				GroupLabel: detail.GroupLabel,
				Items:      []AdminScoreDimensionDTO{},
			})
		}
		groups[idx].Items = append(groups[idx].Items, item)
		groups[idx].Total++
		if status == "completed" {
			groups[idx].CompletedCount++
		}
	}

	return overview, groups
}

func buildAdminTextSections(analysis *models.VideoAnalysis, eventStats map[string]operationEventStats) []AdminTextSectionDTO {
	values := map[string]string{}
	if analysis != nil {
		values = map[string]string{
			"summary":       analysis.Summary,
			"strengths":     analysis.Strengths,
			"weaknesses":    analysis.Weaknesses,
			"improvements":  analysis.Improvements,
			"analyst_notes": analysis.AnalystNotes,
		}
	}
	metas := []struct {
		key   string
		label string
	}{
		{key: "summary", label: "综合评价"},
		{key: "strengths", label: "核心优势"},
		{key: "weaknesses", label: "待提升点"},
		{key: "improvements", label: "训练建议"},
		{key: "analyst_notes", label: "分析师补充说明"},
	}
	sections := make([]AdminTextSectionDTO, 0, len(metas))
	for _, meta := range metas {
		value := strings.TrimSpace(values[meta.key])
		stats := eventStats[meta.key]
		sections = append(sections, AdminTextSectionDTO{
			FieldKey:      meta.key,
			FieldLabel:    meta.label,
			Filled:        value != "",
			WordCount:     runeCount(value),
			LastUpdatedAt: stats.lastUpdatedAt,
			Preview:       textPreview(value, 100),
		})
	}
	return sections
}

func buildAdminHighlightOverview(highlights []models.AnalysisHighlight) AdminHighlightOverviewDTO {
	overview := AdminHighlightOverviewDTO{HighlightCount: len(highlights)}
	for _, highlight := range highlights {
		if highlight.IncludeInReport {
			overview.IncludedHighlightCount++
		}
		if highlight.Mode == models.HighlightModeRange {
			overview.RangeCount++
		} else {
			overview.PointCount++
		}
		switch highlight.ClipStatus {
		case models.HighlightClipReady:
			overview.ClipReadyCount++
		case models.HighlightClipQueued, models.HighlightClipProcessing:
			overview.ClipProcessingCount++
		case models.HighlightClipFailed:
			overview.ClipFailedCount++
		}
	}
	return overview
}

func buildAdminHighlightItems(highlights []models.AnalysisHighlight) []AdminHighlightItemDTO {
	items := make([]AdminHighlightItemDTO, 0, len(highlights))
	for _, highlight := range highlights {
		items = append(items, AdminHighlightItemDTO{
			ID:              highlight.ID,
			Timestamp:       highlight.Timestamp,
			MarkerType:      string(highlight.MarkerType),
			Mode:            string(highlight.Mode),
			StartTimeMs:     highlight.StartTimeMs,
			EndTimeMs:       highlight.EndTimeMs,
			TagType:         string(highlight.TagType),
			Description:     highlight.Description,
			VideoClipURL:    highlight.VideoClipURL,
			ClipStatus:      string(highlight.ClipStatus),
			ClipError:       highlight.ClipError,
			ClipVersion:     highlight.ClipVersion,
			ClipGeneratedAt: highlight.ClipGeneratedAt,
			IncludeInReport: highlight.IncludeInReport,
			SortOrder:       highlight.SortOrder,
			UpdatedAt:       highlight.UpdatedAt,
		})
	}
	return items
}

func buildAdminReportOverview(analysis *models.VideoAnalysis, report *models.Report, events []models.AnalysisOperationEvent) AdminReportOverviewDTO {
	overview := AdminReportOverviewDTO{}
	if analysis != nil {
		overview.AIReportStatus = analysis.AIReportStatus
		overview.AIReportVersion = analysis.AIReportVersion
		overview.TemplateVersion = analysis.AIReportTemplateVersion
		overview.InputSnapshot = strings.TrimSpace(analysis.AIReportInputSnapshot) != ""
		overview.WordReportReady = strings.TrimSpace(analysis.RatingReportMD) != ""
	}
	if report != nil {
		overview.ReviewStatus = string(report.Status)
		overview.ReviewRemark = report.ReviewRemark
		overview.WordReportReady = overview.WordReportReady || strings.TrimSpace(report.AIReportURL) != ""
		overview.PDFReportReady = strings.TrimSpace(report.PdfURL) != ""
	}
	for _, event := range events {
		if event.EventType == "report_submitted" {
			submittedAt := event.CreatedAt
			if overview.SubmittedAt == nil || submittedAt.After(*overview.SubmittedAt) {
				overview.SubmittedAt = &submittedAt
			}
		}
	}
	return overview
}

func buildAdminOrderProgress(order *models.Order, assignment *models.OrderAssignment, analysis *models.VideoAnalysis, report *models.Report, completion AdminOrderAnalysisCompletionDTO) AdminOrderAnalysisProgressDTO {
	stage, label, percent := detectAdminOrderProgressStage(order, assignment, analysis, report)
	sla := buildAdminOrderProgressSLA(order)
	riskLevel, riskLabel := "normal", "正常"
	if sla.IsOverdue {
		riskLevel, riskLabel = "danger", "已逾期"
	} else if sla.IsNearDeadline {
		riskLevel, riskLabel = "warning", "临近截止"
	}
	if analysis != nil && analysis.AIReportStatus == "failed" {
		riskLevel, riskLabel = "exception", "AI生成失败"
	}
	if report != nil && report.Status == models.ReportStatusFailed {
		riskLevel, riskLabel = "exception", "报告被驳回"
	}

	progress := AdminOrderAnalysisProgressDTO{
		Stage:      stage,
		StageLabel: label,
		Percent:    percent,
		RiskLevel:  riskLevel,
		RiskLabel:  riskLabel,
		SLA:        sla,
		Summary: AdminOrderProgressSummaryDTO{
			ScoreCompleted:         completion.ScoreOverview.CompletedCount,
			ScoreTotal:             completion.ScoreOverview.ScoreTotal,
			ScoreCommentCompleted:  completion.ScoreOverview.CompletedCount + completion.ScoreOverview.CommentOnlyCount,
			ScoreValueChanged:      completion.ScoreOverview.CompletedCount + completion.ScoreOverview.ScoreOnlyCount,
			TextSectionsCompleted:  completedTextSectionCount(completion.TextSections),
			TextSectionsTotal:      len(completion.TextSections),
			HighlightCount:         completion.HighlightOverview.HighlightCount,
			IncludedHighlightCount: completion.HighlightOverview.IncludedHighlightCount,
			ClipReadyCount:         completion.HighlightOverview.ClipReadyCount,
			ClipFailedCount:        completion.HighlightOverview.ClipFailedCount,
			HasSummary:             hasTextSection(completion.TextSections, "summary"),
			HasReport:              report != nil,
		},
	}
	if assignment != nil {
		assignedAt := assignment.AssignedAt
		progress.Assignment = &AdminOrderProgressAssignmentDTO{
			Status:         string(assignment.Status),
			AssignedAt:     &assignedAt,
			RespondedAt:    assignment.RespondedAt,
			RejectedReason: assignment.RejectedReason,
		}
	} else if order.AssignedAt != nil {
		progress.Assignment = &AdminOrderProgressAssignmentDTO{
			Status:     "unknown",
			AssignedAt: order.AssignedAt,
		}
	}
	if analysis != nil {
		updatedAt := analysis.UpdatedAt
		progress.Analysis = &AdminOrderProgressAnalysisDTO{
			ID:              analysis.ID,
			Status:          string(analysis.Status),
			AIReportStatus:  analysis.AIReportStatus,
			AIReportVersion: analysis.AIReportVersion,
			UpdatedAt:       &updatedAt,
		}
	}
	return progress
}

func detectAdminOrderProgressStage(order *models.Order, assignment *models.OrderAssignment, analysis *models.VideoAnalysis, report *models.Report) (string, string, int) {
	if order == nil {
		return "not_started", "未进入分析", 0
	}
	switch order.Status {
	case models.OrderStatusCancelled:
		return "cancelled", "已取消", 0
	case models.OrderStatusRefunded:
		return "refunded", "已退款", 0
	case models.OrderStatusCompleted:
		return "completed", "已完成", 100
	}
	if report != nil && report.Status == models.ReportStatusFailed {
		return "revision_required", "返工中", 70
	}
	if analysis != nil {
		if analysis.AIReportStatus == "failed" {
			return "revision_required", "AI生成失败", 70
		}
		if analysis.AIReportStatus == "admin_rejected" {
			return "revision_required", "返工中", 70
		}
		if analysis.Status == models.AnalysisStatusSubmitted || (report != nil && report.Status == models.ReportStatusProcessing) {
			return "review_pending", "待管理员审核", 90
		}
		if analysis.AIReportStatus == "generating" || analysis.AIReportStatus == "regenerating" || analysis.Status == models.AnalysisStatusGenerating {
			return "ai_generating", "报告生成中", 70
		}
		if analysis.AIReportStatus == "draft" || analysis.AIReportStatus == "confirmed" {
			return "analyst_editing", "分析师编辑中", 80
		}
		if analysis.Status == models.AnalysisStatusDraft {
			return "drafting", "草稿编辑中", 50
		}
		if analysis.Status == models.AnalysisStatusScoring {
			return "scoring", "评分中", 35
		}
	}
	if order.Status == models.OrderStatusProcessing {
		return "accepted", "已接单", 20
	}
	if order.Status == models.OrderStatusAssigned || (assignment != nil && assignment.Status == models.OrderAssignmentStatusPending) {
		return "waiting_accept", "待接单", 10
	}
	if order.Status == models.OrderStatusUploaded {
		return "waiting_dispatch", "待派发", 5
	}
	return "not_started", "未进入分析", 0
}

func buildAdminOrderProgressSLA(order *models.Order) AdminOrderProgressSLADTO {
	if order == nil || order.Deadline == nil {
		return AdminOrderProgressSLADTO{Label: "无截止时间"}
	}
	if order.Status == models.OrderStatusCompleted || order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusRefunded {
		return AdminOrderProgressSLADTO{Deadline: order.Deadline, Label: "已终止"}
	}
	now := time.Now()
	remaining := int64(order.Deadline.Sub(now).Seconds())
	sla := AdminOrderProgressSLADTO{
		Deadline:         order.Deadline,
		RemainingSeconds: remaining,
		IsNearDeadline:   remaining >= 0 && remaining <= int64((12*time.Hour).Seconds()),
		IsOverdue:        remaining < 0,
	}
	if sla.IsOverdue {
		sla.Label = "已逾期"
	} else if sla.IsNearDeadline {
		sla.Label = "临近截止"
	} else {
		sla.Label = "正常"
	}
	return sla
}

func buildAdminOrderProgressExceptions(order *models.Order, assignment *models.OrderAssignment, analysis *models.VideoAnalysis, report *models.Report, progress AdminOrderAnalysisProgressDTO, events []models.AnalysisOperationEvent) []AdminOrderProgressExceptionDTO {
	exceptions := []AdminOrderProgressExceptionDTO{}
	if assignment != nil && assignment.Status == models.OrderAssignmentStatusRejected {
		exceptions = append(exceptions, AdminOrderProgressExceptionDTO{Code: "assignment_rejected", Message: firstNonEmptyProgress(assignment.RejectedReason, "分析师已拒单")})
	}
	if progress.SLA.IsOverdue {
		exceptions = append(exceptions, AdminOrderProgressExceptionDTO{Code: "overdue", Message: "订单分析已逾期"})
	}
	if order != nil && order.Deadline == nil && (order.Status == models.OrderStatusAssigned || order.Status == models.OrderStatusProcessing) {
		exceptions = append(exceptions, AdminOrderProgressExceptionDTO{Code: "missing_deadline", Message: "订单缺少截止时间"})
	}
	if analysis != nil && analysis.AIReportStatus == "failed" {
		exceptions = append(exceptions, AdminOrderProgressExceptionDTO{Code: "ai_report_failed", Message: "AI报告生成失败"})
	}
	if report != nil && report.Status == models.ReportStatusFailed {
		exceptions = append(exceptions, AdminOrderProgressExceptionDTO{Code: "report_rejected", Message: firstNonEmptyProgress(report.ReviewRemark, "报告被管理员驳回")})
	}
	exceptions = append(exceptions, adminMarkedProgressExceptions(events)...)
	return exceptions
}

func adminMarkedProgressExceptions(events []models.AnalysisOperationEvent) []AdminOrderProgressExceptionDTO {
	seen := map[string]bool{}
	exceptions := []AdminOrderProgressExceptionDTO{}
	for _, event := range events {
		if event.EventType != "admin_exception_marked" && event.EventType != "admin_exception_resolved" {
			continue
		}
		code := strings.TrimSpace(event.FieldKey)
		if code == "" {
			code = "admin_marked_exception"
		}
		if seen[code] {
			continue
		}
		seen[code] = true
		if event.EventType == "admin_exception_marked" {
			exceptions = append(exceptions, AdminOrderProgressExceptionDTO{
				Code:    code,
				Message: firstNonEmptyProgress(event.AfterSummary, "管理员标记异常"),
			})
		}
	}
	return exceptions
}

func firstAdminOrderProgressException(exceptions []AdminOrderProgressExceptionDTO) *AdminOrderProgressExceptionDTO {
	if len(exceptions) == 0 {
		return nil
	}
	return &exceptions[0]
}

func buildAdminOrderProgressTimeline(order *models.Order, assignment *models.OrderAssignment, analysis *models.VideoAnalysis, report *models.Report, histories []models.OrderStatusHistory, events []models.AnalysisOperationEvent) []AdminProgressTimelineItemDTO {
	items := []AdminProgressTimelineItemDTO{}
	if order != nil {
		if order.PaidAt != nil {
			items = append(items, AdminProgressTimelineItemDTO{Key: "paid", Label: "用户支付", Status: "done", Time: order.PaidAt, Actor: "用户"})
		}
		if strings.TrimSpace(order.VideoURL) != "" {
			updatedAt := order.UpdatedAt
			items = append(items, AdminProgressTimelineItemDTO{Key: "video_uploaded", Label: "用户上传视频", Status: "done", Time: &updatedAt, Actor: "用户"})
		}
	}
	for _, history := range histories {
		createdAt := history.CreatedAt
		items = append(items, AdminProgressTimelineItemDTO{
			Key:    "order_status_" + string(history.ToStatus),
			Label:  "订单状态变更",
			Status: "done",
			Time:   &createdAt,
			Actor:  history.ActorRole,
			Remark: history.Reason,
		})
	}
	if assignment != nil {
		assignedAt := assignment.AssignedAt
		items = append(items, AdminProgressTimelineItemDTO{Key: "assigned", Label: "管理员派发", Status: "done", Time: &assignedAt, Actor: "管理员"})
		if assignment.RespondedAt != nil {
			label := "分析师接单"
			if assignment.Status == models.OrderAssignmentStatusRejected {
				label = "分析师拒单"
			}
			items = append(items, AdminProgressTimelineItemDTO{Key: "assignment_response", Label: label, Status: "done", Time: assignment.RespondedAt, Actor: "分析师", Remark: assignment.RejectedReason})
		}
	}
	if analysis != nil {
		createdAt := analysis.CreatedAt
		items = append(items, AdminProgressTimelineItemDTO{Key: "analysis_created", Label: "创建视频分析", Status: "done", Time: &createdAt, Actor: "分析师"})
	}
	if report != nil {
		updatedAt := report.UpdatedAt
		items = append(items, AdminProgressTimelineItemDTO{Key: "report_status", Label: "报告审核状态", Status: string(report.Status), Time: &updatedAt, Actor: "管理员", Remark: report.ReviewRemark})
	}
	for _, event := range events {
		createdAt := event.CreatedAt
		items = append(items, AdminProgressTimelineItemDTO{
			Key:    event.EventType,
			Label:  operationEventLabel(event),
			Status: "done",
			Time:   &createdAt,
			Actor:  "分析师",
			Remark: firstNonEmptyProgress(event.AfterSummary, event.FieldLabel),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Time == nil {
			return false
		}
		if items[j].Time == nil {
			return true
		}
		return items[i].Time.Before(*items[j].Time)
	})
	return items
}

func toAdminAnalysisOperationEventDTOs(events []models.AnalysisOperationEvent) []AdminAnalysisOperationEventDTO {
	result := make([]AdminAnalysisOperationEventDTO, 0, len(events))
	for _, event := range events {
		result = append(result, AdminAnalysisOperationEventDTO{
			EventType:     event.EventType,
			Section:       event.Section,
			FieldKey:      event.FieldKey,
			FieldLabel:    event.FieldLabel,
			Summary:       firstNonEmptyProgress(event.AfterSummary, event.FieldLabel),
			BeforeSummary: event.BeforeSummary,
			AfterSummary:  event.AfterSummary,
			Metadata:      event.Metadata,
			CreatedAt:     event.CreatedAt,
		})
	}
	return result
}

func operationEventLabel(event models.AnalysisOperationEvent) string {
	switch event.EventType {
	case "analysis_created":
		return "创建分析记录"
	case "score_saved":
		return "保存评分草稿"
	case "score_dimension_updated":
		return "更新评分项"
	case "text_section_updated":
		return "更新文本评价"
	case "report_submitted":
		return "提交报告审核"
	case "revision_received":
		return "收到返工要求"
	case "highlight_created":
		return "新增高光标记"
	case "highlight_updated":
		return "编辑高光标记"
	case "highlight_deleted":
		return "删除高光标记"
	case "clip_generation_started":
		return "片段生成开始"
	case "clip_generation_completed":
		return "片段生成完成"
	case "clip_generation_failed":
		return "片段生成失败"
	case "clip_export_started":
		return "批量导出开始"
	case "clip_export_completed":
		return "批量导出完成"
	case "clip_export_failed":
		return "批量导出失败"
	case "ai_report_generation_started":
		return "AI报告生成开始"
	case "ai_report_generation_completed":
		return "AI报告生成完成"
	case "ai_report_generation_failed":
		return "AI报告生成失败"
	case "ai_report_updated":
		return "AI报告人工编辑"
	case "admin_reminder_sent":
		return "管理员催办"
	case "admin_exception_marked":
		return "管理员标记异常"
	case "admin_exception_resolved":
		return "管理员解除异常"
	default:
		return event.EventType
	}
}

func completedTextSectionCount(sections []AdminTextSectionDTO) int {
	count := 0
	for _, section := range sections {
		if section.Filled {
			count++
		}
	}
	return count
}

func hasTextSection(sections []AdminTextSectionDTO, key string) bool {
	for _, section := range sections {
		if section.FieldKey == key {
			return section.Filled
		}
	}
	return false
}

func runeCount(value string) int {
	return utf8.RuneCountInString(strings.TrimSpace(value))
}

func textPreview(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || runeCount(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}

func firstNonEmptyProgress(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func operationMetadata(payload map[string]interface{}) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}
