package controllers

import (
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

func setupAdminReportDownloadTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-report-download.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Report{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	service := services.NewAdminService(
		models.NewUserRepository(db),
		models.NewReportRepository(db),
		models.NewOrderRepository(db),
		models.NewAnalystRepository(db),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		models.NewVideoAnalysisRepository(db),
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
	)
	controller := NewAdminController(service, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin/reports/:id/download", controller.DownloadReportDoc)
	return router, db
}

func TestAdminDownloadReportDocSupportsWordAndPDF(t *testing.T) {
	router, db := setupAdminReportDownloadTest(t)
	dir := t.TempDir()
	wordPath := filepath.Join(dir, "正式报告.docx")
	pdfPath := filepath.Join(dir, "正式报告.pdf")
	if err := os.WriteFile(wordPath, []byte("word"), 0644); err != nil {
		t.Fatalf("write word: %v", err)
	}
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	report := models.Report{
		OrderID:     1001,
		UserID:      1,
		AnalystID:   2,
		PlayerName:  "程奕",
		Content:     "content",
		AIReportURL: wordPath,
		PdfURL:      pdfPath,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	cases := []struct {
		docType     string
		contentType string
		fileName    string
	}{
		{docType: "report", contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", fileName: "正式报告.docx"},
		{docType: "pdf", contentType: "application/pdf", fileName: "正式报告.pdf"},
	}
	for _, tt := range cases {
		req := httptest.NewRequest(http.MethodGet, "/admin/reports/"+strconvUint(report.ID)+"/download?type="+tt.docType, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body=%s", tt.docType, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); !strings.Contains(got, tt.contentType) {
			t.Fatalf("%s content-type = %q, want %q", tt.docType, got, tt.contentType)
		}
		if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, tt.fileName) && !strings.Contains(got, "%E6%AD%A3%E5%BC%8F%E6%8A%A5%E5%91%8A") {
			t.Fatalf("%s content-disposition = %q, want filename %q", tt.docType, got, tt.fileName)
		}
	}
}

func TestAdminReportFileRefResolvesUploadURLsToWorkingDirectory(t *testing.T) {
	if got := adminReportFileRef("/uploads/reports/report.pdf"); got != "uploads/reports/report.pdf" {
		t.Fatalf("upload file ref = %q, want uploads/reports/report.pdf", got)
	}
	if got := adminReportFileRef("uploads/reports/report.pdf"); got != "uploads/reports/report.pdf" {
		t.Fatalf("relative file ref = %q, want uploads/reports/report.pdf", got)
	}
	if got := adminReportFileRef("/var/reports/report.pdf"); got != "/var/reports/report.pdf" {
		t.Fatalf("absolute file ref = %q, want /var/reports/report.pdf", got)
	}
	if got := adminReportFileRef("https://example.com/report.pdf"); got != "https://example.com/report.pdf" {
		t.Fatalf("remote file ref = %q, want https://example.com/report.pdf", got)
	}
}

func strconvUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
