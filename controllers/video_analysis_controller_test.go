package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVideoAnalysisControllerTest(t *testing.T) (*gorm.DB, *VideoAnalysisController, models.Analyst, models.Analyst) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Order{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
		&models.Report{},
		&models.ReportVersion{},
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
		&models.VideoClipExportJob{},
		&models.AnalysisOperationEvent{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	ownerUser := models.User{Phone: "13920000001", Password: "test", Role: models.RoleAnalyst, Status: models.StatusActive}
	otherUser := models.User{Phone: "13920000002", Password: "test", Role: models.RoleAnalyst, Status: models.StatusActive}
	if err := db.Create(&ownerUser).Error; err != nil {
		t.Fatalf("create owner user: %v", err)
	}
	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}

	owner := models.Analyst{UserID: ownerUser.ID, Name: "Owner Analyst", Status: models.AnalystStatusActive}
	other := models.Analyst{UserID: otherUser.ID, Name: "Other Analyst", Status: models.AnalystStatusActive}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner analyst: %v", err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other analyst: %v", err)
	}

	return db, NewVideoAnalysisController(db, nil), owner, other
}

func performVideoAnalysisRequest(t *testing.T, analystID uint, method, path string, body any, params gin.Params, handler func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, reader)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set("analystId", analystID)

	handler(c)
	return w
}

func requireAnalysisOperationEvent(t *testing.T, db *gorm.DB, analysisID uint, eventType string) models.AnalysisOperationEvent {
	t.Helper()
	var event models.AnalysisOperationEvent
	if err := db.Where("analysis_id = ? AND event_type = ?", analysisID, eventType).Order("id DESC").First(&event).Error; err != nil {
		t.Fatalf("expected analysis operation event %s for analysis %d: %v", eventType, analysisID, err)
	}
	return event
}

func TestShouldAutoStartAIReportGeneration(t *testing.T) {
	tests := []struct {
		name     string
		analysis *models.VideoAnalysis
		want     bool
	}{
		{
			name:     "nil analysis",
			analysis: nil,
			want:     false,
		},
		{
			name: "fresh analysis",
			analysis: &models.VideoAnalysis{
				AIReportStatus: "",
				AIReport:       "",
			},
			want: true,
		},
		{
			name: "running generation",
			analysis: &models.VideoAnalysis{
				AIReportStatus: "generating",
				AIReport:       "",
			},
			want: false,
		},
		{
			name: "existing draft content",
			analysis: &models.VideoAnalysis{
				AIReportStatus: "draft",
				AIReport:       "已有 AI 报告正文",
			},
			want: false,
		},
		{
			name: "failed without content",
			analysis: &models.VideoAnalysis{
				AIReportStatus: "failed",
				AIReport:       "",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldAutoStartAIReportGeneration(tt.analysis); got != tt.want {
				t.Fatalf("shouldAutoStartAIReportGeneration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVideoAnalysisReadRequiresAssignedAnalyst(t *testing.T) {
	db, ctrl, owner, other := setupVideoAnalysisControllerTest(t)

	analysis := models.VideoAnalysis{
		OrderID:        1,
		AnalystID:      owner.ID,
		UserID:         100,
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	forbidden := performVideoAnalysisRequest(t, other.ID, http.MethodGet, "/video-analysis/"+params[0].Value, nil, params, ctrl.GetAnalysis)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other analyst status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	allowed := performVideoAnalysisRequest(t, owner.ID, http.MethodGet, "/video-analysis/"+params[0].Value, nil, params, ctrl.GetAnalysis)
	if allowed.Code != http.StatusOK {
		t.Fatalf("owner analyst status = %d, want %d", allowed.Code, http.StatusOK)
	}
}

func TestCreateVideoAnalysisFromOrderRequiresAssignedAnalyst(t *testing.T) {
	db, ctrl, owner, other := setupVideoAnalysisControllerTest(t)

	order := models.Order{
		UserID:         100,
		AnalystID:      &owner.ID,
		OrderNo:        "VA-ORDER-001",
		Amount:         99,
		Status:         models.OrderStatusProcessing,
		VideoURL:       "/uploads/videos/source-from-order.mp4",
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	req := CreateAnalysisFromOrderRequest{OrderID: order.ID}
	forbidden := performVideoAnalysisRequest(t, other.ID, http.MethodPost, "/video-analysis/create-from-order", req, nil, ctrl.CreateFromOrder)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other analyst status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	allowed := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/create-from-order", req, nil, ctrl.CreateFromOrder)
	if allowed.Code != http.StatusOK {
		t.Fatalf("owner analyst status = %d, want %d", allowed.Code, http.StatusOK)
	}

	var analysis models.VideoAnalysis
	if err := db.Where("order_id = ?", order.ID).First(&analysis).Error; err != nil {
		t.Fatalf("find created analysis: %v", err)
	}
	if analysis.AnalystID != owner.ID {
		t.Fatalf("created analysis analyst_id = %d, want %d", analysis.AnalystID, owner.ID)
	}
	if analysis.VideoURL != order.VideoURL {
		t.Fatalf("created analysis video_url = %q, want %q", analysis.VideoURL, order.VideoURL)
	}
}

func TestCreateVideoAnalysisFromOrderReturnsExistingForSameAnalyst(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)

	order := models.Order{
		UserID:         100,
		AnalystID:      &owner.ID,
		OrderNo:        "VA-ORDER-IDEMPOTENT",
		Amount:         99,
		Status:         models.OrderStatusProcessing,
		VideoURL:       "/uploads/videos/source-idempotent.mp4",
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	req := CreateAnalysisFromOrderRequest{OrderID: order.ID}
	first := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/create-from-order", req, nil, ctrl.CreateFromOrder)
	if first.Code != http.StatusOK {
		t.Fatalf("first create status = %d, want %d, body=%s", first.Code, http.StatusOK, first.Body.String())
	}

	second := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/create-from-order", req, nil, ctrl.CreateFromOrder)
	if second.Code != http.StatusOK {
		t.Fatalf("second create status = %d, want %d, body=%s", second.Code, http.StatusOK, second.Body.String())
	}

	var count int64
	if err := db.Model(&models.VideoAnalysis{}).Where("order_id = ?", order.ID).Count(&count).Error; err != nil {
		t.Fatalf("count analyses: %v", err)
	}
	if count != 1 {
		t.Fatalf("analysis count = %d, want 1", count)
	}

	var body struct {
		Success bool                 `json:"success"`
		Data    models.VideoAnalysis `json:"data"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if !body.Success || body.Data.OrderID != order.ID || body.Data.AnalystID != owner.ID {
		t.Fatalf("second response = %#v, want existing owned analysis", body)
	}
}

func TestGetAnalysisByOrderRequiresAnalysisOwner(t *testing.T) {
	db, ctrl, owner, other := setupVideoAnalysisControllerTest(t)

	order := models.Order{
		UserID:    100,
		AnalystID: &owner.ID,
		OrderNo:   "VA-ORDER-MISMATCH",
		Amount:    99,
		Status:    models.OrderStatusProcessing,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	analysis := models.VideoAnalysis{
		OrderID:   order.ID,
		AnalystID: other.ID,
		UserID:    order.UserID,
		Status:    models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create mismatched analysis: %v", err)
	}

	res := performVideoAnalysisRequest(t, owner.ID, http.MethodGet, "/video-analysis/by-order?order_id="+strconv.Itoa(int(order.ID)), nil, nil, ctrl.GetAnalysisByOrder)
	if res.Code != http.StatusForbidden {
		t.Fatalf("mismatched owner status = %d, want %d, body=%s", res.Code, http.StatusForbidden, res.Body.String())
	}
}

func TestGetAnalysisByOrderBackfillsMissingVideoURL(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)

	order := models.Order{
		UserID:         100,
		AnalystID:      &owner.ID,
		OrderNo:        "VA-ORDER-VIDEO",
		Amount:         99,
		Status:         models.OrderStatusProcessing,
		VideoURL:       "/uploads/videos/existing-order-source.mp4",
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	analysis := models.VideoAnalysis{
		OrderID:        order.ID,
		AnalystID:      owner.ID,
		UserID:         order.UserID,
		PlayerName:     order.PlayerName,
		PlayerPosition: order.PlayerPosition,
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create existing analysis: %v", err)
	}

	res := performVideoAnalysisRequest(t, owner.ID, http.MethodGet, "/video-analysis/by-order?order_id="+strconv.Itoa(int(order.ID)), nil, nil, ctrl.GetAnalysisByOrder)
	if res.Code != http.StatusOK {
		t.Fatalf("get analysis by order status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var reloaded models.VideoAnalysis
	if err := db.First(&reloaded, analysis.ID).Error; err != nil {
		t.Fatalf("reload analysis: %v", err)
	}
	if reloaded.VideoURL != order.VideoURL {
		t.Fatalf("backfilled analysis video_url = %q, want %q", reloaded.VideoURL, order.VideoURL)
	}
}

func TestUpdateScoresKeepsSubmittedStatus(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)

	analysis := models.VideoAnalysis{
		OrderID:        1,
		AnalystID:      owner.ID,
		UserID:         100,
		PlayerName:     "Submitted Player",
		PlayerPosition: "winger",
		Status:         models.AnalysisStatusSubmitted,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	scores := models.NewDefaultScores()
	scores.BallControl.Score = 9
	req := UpdateScoresRequest{
		Scores:       scores,
		Summary:      "提交后自动保存测试",
		Strengths:    "边路拿球后推进积极，连续动作稳定。",
		Weaknesses:   "对抗后出球节奏仍需提升。",
		Improvements: "保持已提交状态，不回退到评分中。",
		AnalystNotes: "late autosave",
	}
	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/"+params[0].Value+"/scores", req, params, ctrl.UpdateScores)
	if res.Code != http.StatusOK {
		t.Fatalf("update scores status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var reloaded models.VideoAnalysis
	if err := db.First(&reloaded, analysis.ID).Error; err != nil {
		t.Fatalf("reload analysis: %v", err)
	}
	if reloaded.Status != models.AnalysisStatusSubmitted {
		t.Fatalf("analysis status = %s, want %s", reloaded.Status, models.AnalysisStatusSubmitted)
	}
	if reloaded.Summary != req.Summary {
		t.Fatalf("summary = %q, want %q", reloaded.Summary, req.Summary)
	}
	if reloaded.Strengths != req.Strengths {
		t.Fatalf("strengths = %q, want %q", reloaded.Strengths, req.Strengths)
	}
	if reloaded.Weaknesses != req.Weaknesses {
		t.Fatalf("weaknesses = %q, want %q", reloaded.Weaknesses, req.Weaknesses)
	}
}

func TestVideoAnalysisConfirmReportApprovalAndPlayerVisibility(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	reportsDir := t.TempDir()
	ctrl.reportGen = services.NewReportGenerator(reportsDir)

	player := models.User{
		Phone:    "13920000011",
		Password: "test",
		Name:     "Flow Player",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
	}
	admin := models.User{
		Phone:    "13920000012",
		Password: "test",
		Name:     "Flow Admin",
		Role:     models.RoleAdmin,
		Status:   models.StatusActive,
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player user: %v", err)
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	order := models.Order{
		UserID:         player.ID,
		OrderNo:        "VA-FLOW-001",
		Amount:         99,
		Status:         models.OrderStatusUploaded,
		OrderType:      "basic",
		PlayerName:     "Flow Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	userRepo := models.NewUserRepository(db)
	reportRepo := models.NewReportRepository(db)
	orderRepo := models.NewOrderRepository(db)
	analystRepo := models.NewAnalystRepository(db)
	assignmentRepo := models.NewOrderAssignmentRepository(db)
	statusHistoryRepo := models.NewOrderStatusHistoryRepository(db)
	analystService := services.NewAnalystService(analystRepo, orderRepo, userRepo, assignmentRepo, statusHistoryRepo)
	adminService := services.NewAdminService(
		userRepo,
		reportRepo,
		orderRepo,
		analystRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		models.NewVideoAnalysisRepository(db),
		assignmentRepo,
		statusHistoryRepo,
	)
	reportService := services.NewReportService(reportRepo, userRepo)

	if _, err := adminService.AssignOrder(order.ID, owner.ID, admin.ID); err != nil {
		t.Fatalf("assign order: %v", err)
	}
	if err := analystService.AcceptOrder(owner.ID, order.ID); err != nil {
		t.Fatalf("accept order: %v", err)
	}

	createReq := CreateAnalysisFromOrderRequest{OrderID: order.ID}
	createRes := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/create-from-order", createReq, nil, ctrl.CreateFromOrder)
	if createRes.Code != http.StatusOK {
		t.Fatalf("create analysis status = %d, want %d, body=%s", createRes.Code, http.StatusOK, createRes.Body.String())
	}

	var analysis models.VideoAnalysis
	if err := db.Where("order_id = ?", order.ID).First(&analysis).Error; err != nil {
		t.Fatalf("find created analysis: %v", err)
	}

	scores := models.NewDefaultScores()
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("marshal scores: %v", err)
	}
	overallScore := scores.CalculateOverallScore()
	if err := db.Model(&models.VideoAnalysis{}).Where("id = ?", analysis.ID).Updates(map[string]interface{}{
		"scores":          scoresJSON,
		"overall_score":   overallScore,
		"potential_level": models.GetPotentialLevel(overallScore),
		"summary":         "整体表现稳定，边路参与积极。",
		"strengths":       "边路推进积极\n连续动作稳定",
		"weaknesses":      "对抗后出球选择需要提升",
		"improvements":    "继续加强对抗后的传球选择。",
		"analyst_notes":   "临时库闭环验证数据。",
	}).Error; err != nil {
		t.Fatalf("seed analysis scoring fields: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	confirmRes := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/submit-report", nil, params, ctrl.ConfirmReport)
	if confirmRes.Code != http.StatusOK {
		t.Fatalf("submit report status = %d, want %d, body=%s", confirmRes.Code, http.StatusOK, confirmRes.Body.String())
	}

	var submittedAnalysis models.VideoAnalysis
	if err := db.First(&submittedAnalysis, analysis.ID).Error; err != nil {
		t.Fatalf("reload submitted analysis: %v", err)
	}
	if submittedAnalysis.Status != models.AnalysisStatusSubmitted {
		t.Fatalf("analysis status after confirm = %s, want %s", submittedAnalysis.Status, models.AnalysisStatusSubmitted)
	}
	if submittedAnalysis.AIReportStatus != "confirmed" {
		t.Fatalf("analysis ai_report_status after confirm = %q, want confirmed", submittedAnalysis.AIReportStatus)
	}
	if submittedAnalysis.AIReportTemplateVersion != services.VideoAnalysisReportTemplateVersion {
		t.Fatalf("analysis template version = %q, want %q", submittedAnalysis.AIReportTemplateVersion, services.VideoAnalysisReportTemplateVersion)
	}
	if submittedAnalysis.AIReportInputSnapshot == "" {
		t.Fatalf("expected ai report input snapshot after confirm")
	}
	if strings.Contains(submittedAnalysis.AIReportInputSnapshot, player.Phone) {
		t.Fatalf("ai report input snapshot leaked player phone")
	}
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(submittedAnalysis.AIReportInputSnapshot), &snapshot); err != nil {
		t.Fatalf("snapshot should be valid json: %v", err)
	}
	if snapshot["template_version"] != services.VideoAnalysisReportTemplateVersion {
		t.Fatalf("snapshot template_version = %#v", snapshot["template_version"])
	}

	report, err := reportRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find bridged report: %v", err)
	}
	if report == nil {
		t.Fatalf("expected report created from confirmed analysis")
	}
	if report.Status != models.ReportStatusProcessing {
		t.Fatalf("report status after confirm = %s, want %s", report.Status, models.ReportStatusProcessing)
	}
	if report.UserID != player.ID || report.AnalystID != owner.ID || !strings.Contains(report.Content, "整体表现稳定，边路参与积极。") {
		t.Fatalf("bridged report fields mismatch: user=%d analyst=%d content=%q", report.UserID, report.AnalystID, report.Content)
	}
	if report.Strengths != `["边路推进积极","连续动作稳定"]` {
		t.Fatalf("report strengths = %q", report.Strengths)
	}
	if report.Weaknesses != `["对抗后出球选择需要提升"]` {
		t.Fatalf("report weaknesses = %q", report.Weaknesses)
	}
	if report.AIReportURL == "" {
		t.Fatalf("expected ai report url after confirm")
	}
	if !strings.HasPrefix(report.AIReportURL, "/uploads/reports/少年球探_视频分析报告_Flow Player_订单") {
		t.Fatalf("report ai_report_url = %q, want formal docx path", report.AIReportURL)
	}
	if !strings.HasSuffix(report.AIReportURL, ".docx") {
		t.Fatalf("report ai_report_url = %q, want docx suffix", report.AIReportURL)
	}
	if report.PdfURL == "" {
		t.Fatalf("expected pdf url after confirm")
	}
	if !strings.HasPrefix(report.PdfURL, "/uploads/reports/少年球探_视频分析报告_Flow Player_订单") {
		t.Fatalf("report pdf_url = %q, want formal pdf path", report.PdfURL)
	}
	if !strings.HasSuffix(report.PdfURL, ".pdf") {
		t.Fatalf("report pdf_url = %q, want pdf suffix", report.PdfURL)
	}
	if _, err := os.Stat(filepath.Join(reportsDir, filepath.Base(report.AIReportURL))); err != nil {
		t.Fatalf("expected generated docx on disk: %v", err)
	}
	if _, err := os.Stat(filepath.Join(reportsDir, filepath.Base(report.PdfURL))); err != nil {
		t.Fatalf("expected generated pdf on disk: %v", err)
	}
	versions, err := models.FindReportVersionsByReportID(db, report.ID)
	if err != nil {
		t.Fatalf("find report versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("report version count after submit = %d, want 1", len(versions))
	}
	if versions[0].SourceType != models.ReportVersionSourceSystem || versions[0].Status != models.ReportVersionStatusAnalystSubmitted {
		t.Fatalf("unexpected submitted version: %#v", versions[0])
	}
	if !strings.Contains(versions[0].InputSnapshot, "同龄") {
		t.Fatalf("expected peer benchmark guideline in version input snapshot")
	}

	var linkedOrder models.Order
	if err := db.First(&linkedOrder, order.ID).Error; err != nil {
		t.Fatalf("reload linked order: %v", err)
	}
	if linkedOrder.ReportID == nil || *linkedOrder.ReportID != report.ID {
		t.Fatalf("order report_id = %#v, want %d", linkedOrder.ReportID, report.ID)
	}
	if linkedOrder.Status != models.OrderStatusProcessing {
		t.Fatalf("order status before admin review = %s, want %s", linkedOrder.Status, models.OrderStatusProcessing)
	}

	playerReportsBefore, err := reportService.GetUserReports(player.ID, 1, 10)
	if err != nil {
		t.Fatalf("get player reports before approval: %v", err)
	}
	if playerReportsBefore.Total != 0 {
		t.Fatalf("player report list before approval total = %d, want 0", playerReportsBefore.Total)
	}
	if _, allowed, err := reportService.GetReportDetail(report.ID, player.ID, models.RoleUser); err != nil {
		t.Fatalf("get report detail before approval: %v", err)
	} else if allowed {
		t.Fatalf("player should not access processing report before admin approval")
	}

	if err := adminService.ReviewReport(report.ID, models.ReportStatusCompleted, "", admin.ID); err != nil {
		t.Fatalf("approve report: %v", err)
	}

	var approvedReport models.Report
	if err := db.First(&approvedReport, report.ID).Error; err != nil {
		t.Fatalf("reload approved report: %v", err)
	}
	if approvedReport.Status != models.ReportStatusCompleted {
		t.Fatalf("report status after approval = %s, want %s", approvedReport.Status, models.ReportStatusCompleted)
	}
	versions, err = models.FindReportVersionsByReportID(db, report.ID)
	if err != nil {
		t.Fatalf("find approved report versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("report version count after approval = %d, want 2", len(versions))
	}
	if versions[0].Status != models.ReportVersionStatusApproved || versions[0].SourceType != models.ReportVersionSourceAdminReview {
		t.Fatalf("unexpected approved version: %#v", versions[0])
	}

	playerListRecorder := httptest.NewRecorder()
	playerListContext, _ := gin.CreateTestContext(playerListRecorder)
	playerListContext.Request = httptest.NewRequest(http.MethodGet, "/video-analysis/my", nil)
	playerListContext.Set("userId", player.ID)
	ctrl.GetMyAnalyses(playerListContext)
	if playerListRecorder.Code != http.StatusOK {
		t.Fatalf("get my analyses status = %d, want %d, body=%s", playerListRecorder.Code, http.StatusOK, playerListRecorder.Body.String())
	}
	var playerListResponse struct {
		Success bool `json:"success"`
		Data    struct {
			List []struct {
				ReportID          uint                `json:"report_id"`
				ReportStatus      models.ReportStatus `json:"report_status"`
				ReportPDFURL      string              `json:"report_pdf_url"`
				ReportAIReportURL string              `json:"report_ai_report_url"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(playerListRecorder.Body.Bytes(), &playerListResponse); err != nil {
		t.Fatalf("unmarshal player list response: %v", err)
	}
	if !playerListResponse.Success || len(playerListResponse.Data.List) != 1 {
		t.Fatalf("unexpected player list response: %#v", playerListResponse)
	}
	if playerListResponse.Data.List[0].ReportID != report.ID {
		t.Fatalf("player list report_id = %d, want %d", playerListResponse.Data.List[0].ReportID, report.ID)
	}
	if playerListResponse.Data.List[0].ReportStatus != models.ReportStatusCompleted {
		t.Fatalf("player list report_status = %s, want %s", playerListResponse.Data.List[0].ReportStatus, models.ReportStatusCompleted)
	}
	if playerListResponse.Data.List[0].ReportAIReportURL == "" {
		t.Fatalf("expected report_ai_report_url in player list response")
	}
	if playerListResponse.Data.List[0].ReportPDFURL == "" {
		t.Fatalf("expected report_pdf_url in player list response")
	}

	var completedOrder models.Order
	if err := db.First(&completedOrder, order.ID).Error; err != nil {
		t.Fatalf("reload completed order: %v", err)
	}
	if completedOrder.Status != models.OrderStatusCompleted {
		t.Fatalf("order status after approval = %s, want %s", completedOrder.Status, models.OrderStatusCompleted)
	}
	if completedOrder.CompletedAt == nil {
		t.Fatalf("expected completed_at after report approval")
	}

	var completedAnalysis models.VideoAnalysis
	if err := db.First(&completedAnalysis, analysis.ID).Error; err != nil {
		t.Fatalf("reload completed analysis: %v", err)
	}
	if completedAnalysis.Status != models.AnalysisStatusCompleted {
		t.Fatalf("analysis status after approval = %s, want %s", completedAnalysis.Status, models.AnalysisStatusCompleted)
	}
	if completedAnalysis.AIReportStatus != "confirmed" {
		t.Fatalf("analysis ai_report_status after approval = %q, want confirmed", completedAnalysis.AIReportStatus)
	}

	playerReportsAfter, err := reportService.GetUserReports(player.ID, 1, 10)
	if err != nil {
		t.Fatalf("get player reports after approval: %v", err)
	}
	if playerReportsAfter.Total != 1 || len(playerReportsAfter.List) != 1 || playerReportsAfter.List[0].ID != report.ID {
		t.Fatalf("player reports after approval = total %d list %#v, want report %d", playerReportsAfter.Total, playerReportsAfter.List, report.ID)
	}
	if _, allowed, err := reportService.GetReportDetail(report.ID, player.ID, models.RoleUser); err != nil {
		t.Fatalf("get report detail after approval: %v", err)
	} else if !allowed {
		t.Fatalf("player should access completed report after admin approval")
	}

	histories, err := statusHistoryRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find status histories: %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("status history count = %d, want 3", len(histories))
	}
	if histories[0].ToStatus != models.OrderStatusAssigned ||
		histories[1].ToStatus != models.OrderStatusProcessing ||
		histories[2].ToStatus != models.OrderStatusCompleted {
		t.Fatalf("unexpected status history chain: %#v", histories)
	}
}

func TestConfirmReportDoesNotRestartExistingAIReportGeneration(t *testing.T) {
	db, _, owner, _ := setupVideoAnalysisControllerTest(t)
	ctrl := NewVideoAnalysisController(db, services.NewAIService(services.AIConfig{
		APIKey:  "test-key",
		BaseURL: "https://example.invalid/v1",
		Model:   "gpt-5.5",
	}))
	ctrl.reportGen = nil

	player := models.User{Phone: "13920000101", Password: "test", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	order := models.Order{
		UserID:         player.ID,
		AnalystID:      &owner.ID,
		OrderNo:        "TEST-CONFIRM-AI-RUNNING",
		Status:         models.OrderStatusProcessing,
		OrderType:      "pro",
		PlayerName:     "Flow Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	scores := models.NewDefaultScores()
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("marshal scores: %v", err)
	}
	overallScore := scores.CalculateOverallScore()

	analysis := models.VideoAnalysis{
		OrderID:         order.ID,
		AnalystID:       owner.ID,
		UserID:          player.ID,
		PlayerName:      "Flow Player",
		PlayerPosition:  "winger",
		Scores:          scoresJSON,
		OverallScore:    overallScore,
		PotentialLevel:  models.GetPotentialLevel(overallScore),
		Summary:         "整体表现稳定，边路参与积极。",
		Strengths:       "边路推进积极\n连续动作稳定",
		Weaknesses:      "对抗后出球选择需要提升",
		Improvements:    "继续加强对抗后的传球选择。",
		AnalystNotes:    "已有 AI 任务在跑。",
		AIReportStatus:  "generating",
		AIReportVersion: 1,
		Status:          models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/submit-report", nil, params, ctrl.ConfirmReport)
	if res.Code != http.StatusOK {
		t.Fatalf("submit report status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var submitted models.VideoAnalysis
	if err := db.First(&submitted, analysis.ID).Error; err != nil {
		t.Fatalf("reload analysis: %v", err)
	}
	if submitted.Status != models.AnalysisStatusSubmitted {
		t.Fatalf("analysis status after confirm = %s, want %s", submitted.Status, models.AnalysisStatusSubmitted)
	}
	if submitted.AIReportStatus != "generating" {
		t.Fatalf("analysis ai_report_status after confirm = %q, want generating", submitted.AIReportStatus)
	}
	if submitted.AIReportVersion != 2 {
		t.Fatalf("analysis ai_report_version after confirm = %d, want 2", submitted.AIReportVersion)
	}

	var generationEventCount int64
	if err := db.Model(&models.AnalysisOperationEvent{}).
		Where("analysis_id = ? AND event_type = ?", analysis.ID, "ai_report_generation_started").
		Count(&generationEventCount).Error; err != nil {
		t.Fatalf("count ai_report_generation_started events: %v", err)
	}
	if generationEventCount != 0 {
		t.Fatalf("ai_report_generation_started count = %d, want 0", generationEventCount)
	}

	requireAnalysisOperationEvent(t, db, analysis.ID, "report_submitted")
}

func TestUpdateAIReportRecordsManualEditEvent(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:                 701,
		AnalystID:               owner.ID,
		UserID:                  1701,
		PlayerName:              "AI Edit Player",
		PlayerPosition:          "midfielder",
		Status:                  models.AnalysisStatusSubmitted,
		AIReport:                "原始AI报告内容",
		AIReportStatus:          "draft",
		AIReportVersion:         2,
		AIReportTemplateVersion: services.VideoAnalysisReportTemplateVersion,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	report := models.Report{
		OrderID:        analysis.OrderID,
		UserID:         analysis.UserID,
		AnalystID:      analysis.AnalystID,
		PlayerName:     analysis.PlayerName,
		PlayerPosition: analysis.PlayerPosition,
		Status:         models.ReportStatusProcessing,
		Content:        "待审核报告",
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/"+params[0].Value+"/ai-report", map[string]string{
		"report": "人工优化后的AI报告内容，补充了关键片段和训练建议。",
	}, params, ctrl.UpdateAIReport)
	if res.Code != http.StatusOK {
		t.Fatalf("update ai report status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var updated models.VideoAnalysis
	if err := db.First(&updated, analysis.ID).Error; err != nil {
		t.Fatalf("reload analysis: %v", err)
	}
	if updated.AIReportVersion != 3 || updated.AIReportStatus != "draft" {
		t.Fatalf("updated ai report version/status = %d/%s, want 3/draft", updated.AIReportVersion, updated.AIReportStatus)
	}
	var updatedReport models.Report
	if err := db.First(&updatedReport, report.ID).Error; err != nil {
		t.Fatalf("reload report: %v", err)
	}
	if updatedReport.Content != "人工优化后的AI报告内容，补充了关键片段和训练建议。" {
		t.Fatalf("report content = %q, want edited ai report", updatedReport.Content)
	}
	event := requireAnalysisOperationEvent(t, db, analysis.ID, "ai_report_updated")
	if event.Section != "ai_report" || !strings.Contains(event.AfterSummary, "人工编辑AI报告") {
		t.Fatalf("manual edit event = %#v", event)
	}
	versions, err := models.FindReportVersionsByReportID(db, report.ID)
	if err != nil {
		t.Fatalf("find report versions: %v", err)
	}
	if len(versions) != 1 || versions[0].SourceType != models.ReportVersionSourceOnlineEdit || versions[0].VersionNo != 3 {
		t.Fatalf("manual edit versions = %#v, want one online edit v3", versions)
	}
}

func TestUpdateAIReportRegeneratesDownloadableWordAndPDF(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	reportsDir := t.TempDir()
	ctrl.reportGen = services.NewReportGenerator(reportsDir)

	user := models.User{Phone: "13920008888", Password: "test", Role: models.RoleUser, Status: models.StatusActive, Name: "Doc Player"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analysis := models.VideoAnalysis{
		OrderID:                 702,
		AnalystID:               owner.ID,
		UserID:                  user.ID,
		PlayerName:              "Doc Player",
		PlayerPosition:          "midfielder",
		Status:                  models.AnalysisStatusSubmitted,
		Summary:                 "原始综合评价",
		AIReport:                "原始AI报告内容",
		AIReportStatus:          "draft",
		AIReportVersion:         2,
		AIReportTemplateVersion: services.VideoAnalysisReportTemplateVersion,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	report := models.Report{
		OrderID:        analysis.OrderID,
		UserID:         analysis.UserID,
		AnalystID:      analysis.AnalystID,
		PlayerName:     analysis.PlayerName,
		PlayerPosition: analysis.PlayerPosition,
		Status:         models.ReportStatusProcessing,
		Content:        "旧报告内容",
		AIReportURL:    "/uploads/reports/old.docx",
		PdfURL:         "/uploads/reports/old.pdf",
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/"+params[0].Value+"/ai-report", map[string]string{
		"report": "人工优化后的AI报告内容，补充了关键片段和训练建议。",
	}, params, ctrl.UpdateAIReport)
	if res.Code != http.StatusOK {
		t.Fatalf("update ai report status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var updatedReport models.Report
	if err := db.First(&updatedReport, report.ID).Error; err != nil {
		t.Fatalf("reload report: %v", err)
	}
	if !strings.Contains(updatedReport.AIReportURL, "_v3.docx") {
		t.Fatalf("ai_report_url = %q, want regenerated v3 docx", updatedReport.AIReportURL)
	}
	if !strings.Contains(updatedReport.PdfURL, "_v3.pdf") {
		t.Fatalf("pdf_url = %q, want regenerated v3 pdf", updatedReport.PdfURL)
	}
	if _, err := os.Stat(filepath.Join(reportsDir, filepath.Base(updatedReport.AIReportURL))); err != nil {
		t.Fatalf("expected regenerated word file: %v", err)
	}

	versions, err := models.FindReportVersionsByReportID(db, report.ID)
	if err != nil {
		t.Fatalf("find report versions: %v", err)
	}
	if len(versions) != 1 || versions[0].WordURL != updatedReport.AIReportURL || versions[0].PDFURL != updatedReport.PdfURL {
		t.Fatalf("manual edit version files = %#v, want current report files", versions)
	}
}

func TestGenerateAIReportRegeneratesCurrentReportFiles(t *testing.T) {
	db, _, owner, _ := setupVideoAnalysisControllerTest(t)
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected ai path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"重新生成后的AI报告正文，包含新的训练建议。"}}]}`))
	}))
	defer aiServer.Close()

	ctrl := NewVideoAnalysisController(db, services.NewAIService(services.AIConfig{
		APIKey:  "test-key",
		BaseURL: aiServer.URL,
		Model:   "gpt-5.5",
	}))
	reportsDir := t.TempDir()
	ctrl.reportGen = services.NewReportGenerator(reportsDir)

	player := models.User{Phone: "13920009999", Password: "test", Name: "Regen Player", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	order := models.Order{
		UserID:         player.ID,
		AnalystID:      &owner.ID,
		OrderNo:        "TEST-REGEN-AI",
		Status:         models.OrderStatusProcessing,
		OrderType:      "pro",
		PlayerName:     "Regen Player",
		PlayerPosition: "winger",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	scores := models.NewDefaultScores()
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("marshal scores: %v", err)
	}
	overallScore := scores.CalculateOverallScore()
	analysis := models.VideoAnalysis{
		OrderID:                 order.ID,
		AnalystID:               owner.ID,
		UserID:                  player.ID,
		PlayerName:              "Regen Player",
		PlayerPosition:          "winger",
		Scores:                  scoresJSON,
		OverallScore:            overallScore,
		PotentialLevel:          models.GetPotentialLevel(overallScore),
		Summary:                 "整体表现稳定，边路参与积极。",
		Strengths:               "边路推进积极\n连续动作稳定",
		Weaknesses:              "对抗后出球选择需要提升",
		Improvements:            "继续加强对抗后的传球选择。",
		AnalystNotes:            "重新生成闭环验证。",
		AIReport:                "旧AI报告正文",
		AIReportStatus:          "draft",
		AIReportVersion:         2,
		AIReportTemplateVersion: services.VideoAnalysisReportTemplateVersion,
		Status:                  models.AnalysisStatusSubmitted,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	report := models.Report{
		OrderID:        order.ID,
		UserID:         analysis.UserID,
		AnalystID:      analysis.AnalystID,
		PlayerName:     analysis.PlayerName,
		PlayerPosition: analysis.PlayerPosition,
		Status:         models.ReportStatusProcessing,
		Content:        "旧报告内容",
		AIReportURL:    "/uploads/reports/old.docx",
		PdfURL:         "/uploads/reports/old.pdf",
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/generate-ai-report", GenerateAIReportRequest{AnalysisID: analysis.ID}, nil, ctrl.GenerateAIReport)
	if res.Code != http.StatusOK {
		t.Fatalf("generate ai report status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var updatedReport models.Report
	var updatedAnalysis models.VideoAnalysis
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_ = db.First(&updatedReport, report.ID).Error
		_ = db.First(&updatedAnalysis, analysis.ID).Error
		if strings.Contains(updatedReport.AIReportURL, "_v3.docx") && updatedAnalysis.AIReportStatus == "draft" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if updatedAnalysis.AIReportVersion != 3 || updatedAnalysis.AIReport != "重新生成后的AI报告正文，包含新的训练建议。" {
		t.Fatalf("updated analysis version/report = %d/%q", updatedAnalysis.AIReportVersion, updatedAnalysis.AIReport)
	}
	if updatedReport.Content != "重新生成后的AI报告正文，包含新的训练建议。" {
		t.Fatalf("report content = %q, want regenerated ai report", updatedReport.Content)
	}
	if !strings.Contains(updatedReport.AIReportURL, "_v3.docx") || !strings.Contains(updatedReport.PdfURL, "_v3.pdf") {
		t.Fatalf("regenerated report urls = %q / %q", updatedReport.AIReportURL, updatedReport.PdfURL)
	}
	if _, err := os.Stat(filepath.Join(reportsDir, filepath.Base(updatedReport.AIReportURL))); err != nil {
		t.Fatalf("expected regenerated word file: %v", err)
	}
	versions, err := models.FindReportVersionsByReportID(db, report.ID)
	if err != nil {
		t.Fatalf("find report versions: %v", err)
	}
	if len(versions) != 1 || versions[0].SourceType != models.ReportVersionSourceAI || versions[0].VersionNo != 3 {
		t.Fatalf("regenerated versions = %#v, want one ai v3", versions)
	}
}
