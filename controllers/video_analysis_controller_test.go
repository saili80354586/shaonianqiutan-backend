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
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
		&models.VideoClipExportJob{},
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

	reportContent := "# Flow Player AI Report\n\n整体表现稳定。"
	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	updateRes := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/"+params[0].Value+"/ai-report", map[string]string{
		"report": reportContent,
	}, params, ctrl.UpdateAIReport)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update ai report status = %d, want %d, body=%s", updateRes.Code, http.StatusOK, updateRes.Body.String())
	}
	var editedAnalysis models.VideoAnalysis
	if err := db.First(&editedAnalysis, analysis.ID).Error; err != nil {
		t.Fatalf("reload edited analysis: %v", err)
	}
	if editedAnalysis.AIReportVersion != 1 {
		t.Fatalf("ai report version after edit = %d, want 1", editedAnalysis.AIReportVersion)
	}

	confirmRes := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/confirm-ai-report", nil, params, ctrl.ConfirmAIReport)
	if confirmRes.Code != http.StatusOK {
		t.Fatalf("confirm ai report status = %d, want %d, body=%s", confirmRes.Code, http.StatusOK, confirmRes.Body.String())
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
	if report.UserID != player.ID || report.AnalystID != owner.ID || report.Content != reportContent {
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
