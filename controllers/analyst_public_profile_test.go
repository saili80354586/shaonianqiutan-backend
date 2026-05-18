package controllers

import (
	"net/http"
	"net/http/httptest"
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

func setupAnalystPublicProfileControllerTest(t *testing.T) (*gorm.DB, *AnalystController, models.Analyst) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Order{},
		&models.Report{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
		&models.OfficialAnalysisTask{},
		&models.OfficialContentAdoption{},
		&models.OfficialContentPublishRecord{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := models.User{
		Phone:    "13920000002",
		Password: "test",
		Name:     "公开接口分析师",
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analyst := models.Analyst{
		UserID: user.ID,
		Name:   "公开接口分析师",
		Status: models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	service := services.NewAnalystService(
		models.NewAnalystRepository(db),
		models.NewOrderRepository(db),
		models.NewUserRepository(db),
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
	)
	return db, NewAnalystController(service, db), analyst
}

func TestGetAnalystOfficialWorksPublicEndpoint(t *testing.T) {
	db, ctrl, analyst := setupAnalystPublicProfileControllerTest(t)
	router := gin.New()
	router.GET("/analysts/:id/official-works", ctrl.GetAnalystOfficialWorks)

	task := models.OfficialAnalysisTask{
		TaskNo:              "PUBLIC-WORK-API-001",
		Title:               "官方选题",
		MatchName:           "U 系列国家队赛事",
		AuthorizationStatus: "authorized",
		MaxAcceptCount:      1,
		VisibleLevelMin:     "L1",
		Status:              models.OfficialAnalysisTaskPublished,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	adoption := models.OfficialContentAdoption{
		TaskID:         task.ID,
		SubmissionID:   1,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionKeySpread,
		Channel:        "douyin",
		WorkTitle:      "国家队赛事精选",
		IsPublic:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.Create(&adoption).Error; err != nil {
		t.Fatalf("create adoption: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analysts/"+strconv.FormatUint(uint64(analyst.ID), 10)+"/official-works", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !containsAll(body, []string{"国家队赛事精选", "U 系列国家队赛事", "key_spread"}) {
		t.Fatalf("response body missing official work fields: %s", body)
	}
}

func containsAll(text string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			return false
		}
	}
	return true
}
