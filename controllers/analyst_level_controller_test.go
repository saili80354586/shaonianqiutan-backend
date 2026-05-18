package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalystLevelControllerTest(t *testing.T) (*gorm.DB, *AnalystLevelController, models.User, models.Analyst) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.AnalystLevel{},
		&models.AnalystLevelApplication{},
		&models.AnalystGrowthSnapshot{},
		&models.AnalystLevelHistory{},
		&models.OfficialAnalysisSubmission{},
		&models.OfficialAnalysisTask{},
		&models.OfficialContentAdoption{},
		&models.OfficialContentPublishRecord{},
		&models.Order{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := models.SeedDefaultAnalystLevels(db); err != nil {
		t.Fatalf("seed levels: %v", err)
	}

	user := models.User{
		Phone:    "13919990001",
		Password: "test",
		Name:     "等级接口分析师",
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analyst := models.Analyst{
		UserID:    user.ID,
		Name:      "等级接口分析师",
		LevelCode: "L1",
		Status:    models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	return db, NewAnalystLevelController(services.NewAnalystLevelService(db)), user, analyst
}

func TestAnalystLevelAPIApplicationReviewAndManualSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, ctrl, user, analyst := setupAnalystLevelControllerTest(t)
	router := gin.New()
	router.GET("/analyst/level", func(c *gin.Context) {
		c.Set("userId", user.ID)
		ctrl.GetMyLevel(c)
	})
	router.POST("/analyst/level-applications", func(c *gin.Context) {
		c.Set("userId", user.ID)
		ctrl.SubmitMyApplication(c)
	})
	router.GET("/analyst/level-applications", func(c *gin.Context) {
		c.Set("userId", user.ID)
		ctrl.ListMyApplications(c)
	})
	router.GET("/admin/analyst-levels", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListLevels(c)
	})
	router.GET("/admin/analyst-level-applications", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListAdminApplications(c)
	})
	router.GET("/admin/analyst-level-applications/:id", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.GetApplication(c)
	})
	router.POST("/admin/analyst-level-applications/:id/review", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ReviewApplication(c)
	})
	router.PUT("/admin/analysts/:id/level", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.SetAnalystLevel(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analyst/level", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get my level status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	submitBody := mustAnalystLevelJSON(t, gin.H{
		"requested_level_code": "L3",
		"application_reason":   "希望申请优选分析师",
		"experience_summary":   "完成过多场比赛分析",
	})
	req = httptest.NewRequest(http.MethodPost, "/analyst/level-applications", bytes.NewReader(submitBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit application status = %d body=%s", rec.Code, rec.Body.String())
	}

	var app models.AnalystLevelApplication
	if err := db.Where("analyst_id = ?", analyst.ID).First(&app).Error; err != nil {
		t.Fatalf("find application: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/analyst/level-applications", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "希望申请优选分析师") {
		t.Fatalf("list my applications status = %d body=%s, want submitted application", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/analyst-level-applications?status=pending", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list applications status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/analyst-level-applications/"+analystLevelIDString(app.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "希望申请优选分析师") {
		t.Fatalf("get application status = %d body=%s, want application detail", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	reviewBody := mustAnalystLevelJSON(t, gin.H{
		"status":              "adjusted",
		"reviewed_level_code": "L2",
		"review_note":         "先调整为认证分析师",
	})
	req = httptest.NewRequest(http.MethodPost, "/admin/analyst-level-applications/"+analystLevelIDString(app.ID)+"/review", bytes.NewReader(reviewBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("review application status = %d body=%s", rec.Code, rec.Body.String())
	}

	var updated models.Analyst
	if err := db.First(&updated, analyst.ID).Error; err != nil {
		t.Fatalf("find analyst: %v", err)
	}
	if updated.LevelCode != "L2" {
		t.Fatalf("level after review = %s, want L2", updated.LevelCode)
	}

	rec = httptest.NewRecorder()
	setBody := mustAnalystLevelJSON(t, gin.H{
		"level_code": "L4",
		"note":       "管理员手动定级",
	})
	req = httptest.NewRequest(http.MethodPut, "/admin/analysts/"+analystLevelIDString(analyst.ID)+"/level", bytes.NewReader(setBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("manual set level status = %d body=%s", rec.Code, rec.Body.String())
	}
	if err := db.First(&updated, analyst.ID).Error; err != nil {
		t.Fatalf("find analyst after manual set: %v", err)
	}
	if updated.LevelCode != "L4" || updated.LevelNote != "管理员手动定级" {
		t.Fatalf("level/note after manual set = %s/%s, want L4/note", updated.LevelCode, updated.LevelNote)
	}
}

func mustAnalystLevelJSON(t *testing.T, payload interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}

func analystLevelIDString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
