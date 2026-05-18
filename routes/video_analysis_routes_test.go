package routes_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVideoAnalysisRouteTestRouter(t *testing.T, clipDir string) (*gin.Engine, *gorm.DB, models.Analyst) {
	t.Helper()

	t.Setenv("JWT_SECRET", "video-analysis-route-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", clipDir)
	t.Setenv("BASE_URL", "http://localhost:8080")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "video-analysis-route.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
		&models.VideoClipExportJob{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = oldDB
	})

	user := models.User{
		Phone:       "13930009001",
		Password:    "hashed-password",
		Role:        models.RoleAnalyst,
		CurrentRole: models.RoleAnalyst,
		Status:      models.StatusActive,
		Name:        "视频分析路由测试分析师",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create analyst user: %v", err)
	}
	analyst := models.Analyst{
		UserID: user.ID,
		Name:   "Route Analyst",
		Status: models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst profile: %v", err)
	}

	router := gin.New()
	api := router.Group("/api")
	routes.SetupVideoAnalysisRoutes(api, controllers.NewVideoAnalysisController(db, nil))
	return router, db, analyst
}

func TestVideoAnalysisClipExportRouteDownloadsZip(t *testing.T) {
	clipDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(clipDir, "route-ready.mp4"), []byte("route-ready-clip"), 0644); err != nil {
		t.Fatalf("write ready clip: %v", err)
	}

	router, db, analyst := setupVideoAnalysisRouteTestRouter(t, clipDir)
	analysis := models.VideoAnalysis{
		OrderID:        21,
		AnalystID:      analyst.ID,
		UserID:         301,
		PlayerName:     "Route Export Player",
		PlayerPosition: "winger",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTimeMs := 9000
	highlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:02-00:09",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     2000,
		EndTimeMs:       &endTimeMs,
		TagType:         models.HighlightGoal,
		Description:     "路由导出验证片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/route-ready.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&highlight).Error; err != nil {
		t.Fatalf("create highlight: %v", err)
	}

	token, err := middleware.GenerateToken(analyst.UserID, "13930009001")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/video-analysis/"+strconv.Itoa(int(analysis.ID))+"/clips/export", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/zip") {
		t.Fatalf("content-type = %q, want application/zip", contentType)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	foundManifest := false
	foundClip := false
	for _, file := range zipReader.File {
		switch {
		case file.Name == "markers_manifest.csv":
			foundManifest = true
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("open manifest: %v", err)
			}
			content, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if !strings.Contains(string(content), "路由导出验证片段") {
				t.Fatalf("manifest content = %q, want description", string(content))
			}
		case strings.HasPrefix(file.Name, "01_精彩表现_进球_0m02s-0m09s"):
			foundClip = true
		}
	}
	if !foundManifest || !foundClip {
		t.Fatalf("zip entries missing manifest=%v clip=%v", foundManifest, foundClip)
	}
}

func TestVideoAnalysisSingleClipDownloadRoute(t *testing.T) {
	clipDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(clipDir, "route-single-ready.mp4"), []byte("route-single-clip"), 0644); err != nil {
		t.Fatalf("write ready clip: %v", err)
	}

	router, db, analyst := setupVideoAnalysisRouteTestRouter(t, clipDir)
	analysis := models.VideoAnalysis{
		OrderID:        23,
		AnalystID:      analyst.ID,
		UserID:         303,
		PlayerName:     "Route Single Clip Player",
		PlayerPosition: "midfielder",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTimeMs := 12000
	highlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:04-00:12",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     4000,
		EndTimeMs:       &endTimeMs,
		TagType:         models.HighlightPass,
		Description:     "单片段下载验证。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/route-single-ready.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&highlight).Error; err != nil {
		t.Fatalf("create highlight: %v", err)
	}

	token, err := middleware.GenerateToken(analyst.UserID, "13930009001")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/video-analysis/markers/"+strconv.Itoa(int(highlight.ID))+"/clip/download", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "route-single-ready.mp4") {
		t.Fatalf("content-disposition = %q, want clip filename", rec.Header().Get("Content-Disposition"))
	}
	if rec.Body.String() != "route-single-clip" {
		t.Fatalf("body = %q, want route-single-clip", rec.Body.String())
	}
}

func TestVideoAnalysisClipExportJobRouteDownloadsZip(t *testing.T) {
	clipDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(clipDir, "route-job-ready.mp4"), []byte("route-job-ready-clip"), 0644); err != nil {
		t.Fatalf("write ready clip: %v", err)
	}

	router, db, analyst := setupVideoAnalysisRouteTestRouter(t, clipDir)
	analysis := models.VideoAnalysis{
		OrderID:        22,
		AnalystID:      analyst.ID,
		UserID:         302,
		PlayerName:     "Route Job Export Player",
		PlayerPosition: "winger",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTimeMs := 8000
	highlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:01-00:08",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     1000,
		EndTimeMs:       &endTimeMs,
		TagType:         models.HighlightGoal,
		Description:     "路由异步导出验证片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/route-job-ready.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&highlight).Error; err != nil {
		t.Fatalf("create highlight: %v", err)
	}

	token, err := middleware.GenerateToken(analyst.UserID, "13930009001")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/video-analysis/"+strconv.Itoa(int(analysis.ID))+"/clips/export/jobs", strings.NewReader("{}"))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create job status = %d, want %d, body=%s", createRec.Code, http.StatusOK, createRec.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create job: %v", err)
	}
	if created.Data.ID == "" {
		t.Fatalf("created job id is empty")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/video-analysis/"+strconv.Itoa(int(analysis.ID))+"/clips/export/jobs", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list jobs status = %d, want %d, body=%s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), created.Data.ID) {
		t.Fatalf("list jobs body = %s, want job id %s", listRec.Body.String(), created.Data.ID)
	}

	job := waitForRouteClipExportJobStatus(t, router, token, analysis.ID, created.Data.ID, "ready")
	if job.DownloadURL == "" || job.Progress != 100 {
		t.Fatalf("ready job = %#v, want download URL and progress 100", job)
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/video-analysis/"+strconv.Itoa(int(analysis.ID))+"/clips/export/jobs/"+created.Data.ID+"/download", nil)
	downloadReq.Header.Set("Authorization", "Bearer "+token)
	downloadRec := httptest.NewRecorder()
	router.ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusOK {
		t.Fatalf("download job status = %d, want %d, body=%s", downloadRec.Code, http.StatusOK, downloadRec.Body.String())
	}
	zipReader, err := zip.NewReader(bytes.NewReader(downloadRec.Body.Bytes()), int64(downloadRec.Body.Len()))
	if err != nil {
		t.Fatalf("open job zip: %v", err)
	}
	foundManifest := false
	foundClip := false
	for _, file := range zipReader.File {
		if file.Name == "markers_manifest.csv" {
			foundManifest = true
		}
		if strings.HasPrefix(file.Name, "01_精彩表现_进球_0m01s-0m08s") {
			foundClip = true
		}
	}
	if !foundManifest || !foundClip {
		t.Fatalf("job zip entries missing manifest=%v clip=%v", foundManifest, foundClip)
	}
}

type routeClipExportJob struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	DownloadURL string `json:"download_url"`
}

func waitForRouteClipExportJobStatus(t *testing.T, router *gin.Engine, token string, analysisID uint, jobID string, want string) routeClipExportJob {
	t.Helper()
	var last routeClipExportJob
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/video-analysis/"+strconv.Itoa(int(analysisID))+"/clips/export/jobs/"+jobID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("get job status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var body struct {
			Data routeClipExportJob `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode job: %v", err)
		}
		last = body.Data
		if body.Data.Status == want {
			return body.Data
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach %s, last=%#v", jobID, want, last)
	return last
}
