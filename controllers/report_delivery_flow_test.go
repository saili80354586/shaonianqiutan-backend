package controllers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReportDeliveryFlowTest(t *testing.T) (*gorm.DB, *ReportController, *AnalystController) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Scout{},
		&models.Club{},
		&models.ClubCoach{},
		&models.TeamCoach{},
		&models.Order{},
		&models.Report{},
		&models.ReportVersion{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	userRepo := models.NewUserRepository(db)
	orderRepo := models.NewOrderRepository(db)
	reportRepo := models.NewReportRepository(db)
	analystRepo := models.NewAnalystRepository(db)
	reportService := services.NewReportService(reportRepo, userRepo)
	authService := services.NewAuthService(userRepo, nil, nil, nil, nil, nil, db)
	analystService := services.NewAnalystService(
		analystRepo,
		orderRepo,
		userRepo,
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
	)

	return db, NewReportController(reportService, authService, db), NewAnalystController(analystService, db)
}

func TestDownloadReportFallsBackToMarkdownContent(t *testing.T) {
	db, reportCtrl, _ := setupReportDeliveryFlowTest(t)

	player := models.User{
		Phone:    "13800000001",
		Password: "test-password",
		Name:     "下载球员",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	report := models.Report{
		OrderID:        1,
		UserID:         player.ID,
		AnalystID:      99,
		PlayerName:     "下载球员",
		PlayerPosition: "中场",
		Content:        "在线球探报告正文",
		Status:         models.ReportStatusCompleted,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	router := gin.New()
	router.GET("/reports/:id/download", func(c *gin.Context) {
		c.Set("userId", player.ID)
		reportCtrl.DownloadReport(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reports/1/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("content type = %q, want markdown", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "在线球探报告正文") {
		t.Fatalf("body = %q, want report content", w.Body.String())
	}
}

func TestPlayerReportVersionsOnlyExposeApprovedVersions(t *testing.T) {
	db, reportCtrl, _ := setupReportDeliveryFlowTest(t)

	player := models.User{
		Phone:    "13800000004",
		Password: "test-password",
		Name:     "版本球员",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	report := models.Report{
		OrderID:        11,
		UserID:         player.ID,
		AnalystID:      88,
		PlayerName:     "版本球员",
		PlayerPosition: "边锋",
		Content:        "正式报告正文",
		Status:         models.ReportStatusCompleted,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	versions := []models.ReportVersion{
		{
			ReportID:      report.ID,
			OrderID:       report.OrderID,
			VersionNo:     1,
			SourceType:    models.ReportVersionSourceAI,
			Status:        models.ReportVersionStatusAIDraft,
			Content:       "AI 草稿",
			InputSnapshot: "内部输入快照",
		},
		{
			ReportID:      report.ID,
			OrderID:       report.OrderID,
			VersionNo:     2,
			SourceType:    models.ReportVersionSourceAdminReview,
			Status:        models.ReportVersionStatusApproved,
			Content:       "审核通过正文",
			InputSnapshot: "内部输入快照",
			ReviewRemark:  "内部审核备注",
			WordURL:       "/uploads/reports/final.docx",
		},
	}
	if err := db.Create(&versions).Error; err != nil {
		t.Fatalf("create versions: %v", err)
	}

	router := gin.New()
	router.GET("/reports/:id/versions", func(c *gin.Context) {
		c.Set("userId", player.ID)
		reportCtrl.GetReportVersions(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reports/1/versions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Versions []models.ReportVersion `json:"versions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data.Versions) != 1 {
		t.Fatalf("versions len = %d, want 1: %s", len(payload.Data.Versions), w.Body.String())
	}
	version := payload.Data.Versions[0]
	if version.Status != models.ReportVersionStatusApproved {
		t.Fatalf("version status = %s, want approved", version.Status)
	}
	if version.Content != "" || version.InputSnapshot != "" || version.ReviewRemark != "" {
		t.Fatalf("player version exposed internal fields: %#v", version)
	}
}

func TestAnalystUploadAIReportUpdatesOrderReportURL(t *testing.T) {
	db, _, analystCtrl := setupReportDeliveryFlowTest(t)

	player := models.User{
		Phone:    "13800000002",
		Password: "test-password",
		Name:     "上传球员",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
	}
	analystUser := models.User{
		Phone:    "13800000003",
		Password: "test-password",
		Name:     "上传分析师",
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := db.Create(&analystUser).Error; err != nil {
		t.Fatalf("create analyst user: %v", err)
	}

	analyst := models.Analyst{
		UserID: analystUser.ID,
		Name:   "上传分析师",
		Status: models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	order := models.Order{
		UserID:         player.ID,
		AnalystID:      &analyst.ID,
		OrderNo:        "UPLOAD-REPORT-001",
		Amount:         99,
		Status:         models.OrderStatusProcessing,
		OrderType:      "video",
		PlayerName:     "上传球员",
		PlayerPosition: "前锋",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	report := models.Report{
		OrderID:        order.ID,
		UserID:         player.ID,
		AnalystID:      analyst.ID,
		PlayerName:     "上传球员",
		PlayerPosition: "前锋",
		Content:        "待审核报告正文",
		Status:         models.ReportStatusProcessing,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
	if err := db.Model(&order).Update("report_id", report.ID).Error; err != nil {
		t.Fatalf("link report: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "final-report.docx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("docx placeholder")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	router := gin.New()
	router.POST("/analyst/orders/:id/ai-report/upload", func(c *gin.Context) {
		c.Set("analystId", analyst.ID)
		analystCtrl.UploadAIReport(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/analyst/orders/1/ai-report/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var updated models.Report
	if err := db.First(&updated, report.ID).Error; err != nil {
		t.Fatalf("load updated report: %v", err)
	}
	if updated.AIReportURL != "" {
		t.Cleanup(func() {
			_ = os.Remove("." + updated.AIReportURL)
		})
	}
	if !strings.HasPrefix(updated.AIReportURL, "/uploads/reports/analyst_ai_report_") {
		t.Fatalf("ai_report_url = %q, want analyst upload path", updated.AIReportURL)
	}
	if !strings.HasSuffix(updated.AIReportURL, ".docx") {
		t.Fatalf("ai_report_url = %q, want docx suffix", updated.AIReportURL)
	}
}
