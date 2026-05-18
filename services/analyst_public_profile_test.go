package services

import (
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalystPublicProfileTestService(t *testing.T) (*gorm.DB, *AnalystService, models.Analyst) {
	t.Helper()

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
		&models.OfficialAnalysisSubmission{},
		&models.OfficialContentAdoption{},
		&models.OfficialContentPublishRecord{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := models.User{
		Phone:    "13920000001",
		Password: "test",
		Name:     "公开主页分析师",
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analyst := models.Analyst{
		UserID:    user.ID,
		Name:      "公开主页分析师",
		LevelCode: "L3",
		Status:    models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	service := NewAnalystService(
		models.NewAnalystRepository(db),
		models.NewOrderRepository(db),
		models.NewUserRepository(db),
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
	)
	return db, service, analyst
}

func TestAnalystPublicProfileIncludesOnlyPublicOfficialWorks(t *testing.T) {
	db, service, analyst := setupAnalystPublicProfileTestService(t)

	task := models.OfficialAnalysisTask{
		TaskNo:              "PUBLIC-WORK-001",
		Title:               "2034杯 U12 官方选题",
		MatchName:           "2034杯小组赛",
		AgeGroup:            "U12",
		AuthorizationStatus: "authorized",
		MaxAcceptCount:      2,
		VisibleLevelMin:     "L1",
		Status:              models.OfficialAnalysisTaskPublished,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	visible := models.OfficialContentAdoption{
		TaskID:         task.ID,
		SubmissionID:   1,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionOfficialPublished,
		Channel:        "douyin",
		WorkTitle:      "2034杯精选分析",
		WorkSummary:    "官方采用作品摘要",
		AdoptionNote:   "节奏判断清晰",
		IsPublic:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	hidden := models.OfficialContentAdoption{
		TaskID:         task.ID,
		SubmissionID:   2,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionMaterial,
		WorkTitle:      "内部素材",
		IsPublic:       false,
		CreatedAt:      time.Now().Add(time.Second),
		UpdatedAt:      time.Now().Add(time.Second),
	}
	if err := db.Create(&visible).Error; err != nil {
		t.Fatalf("create visible adoption: %v", err)
	}
	if err := db.Create(&hidden).Error; err != nil {
		t.Fatalf("create hidden adoption: %v", err)
	}

	profile, err := service.GetAnalystPublicProfile(analyst.ID)
	if err != nil {
		t.Fatalf("get public profile: %v", err)
	}
	if len(profile.OfficialWorks) != 1 {
		t.Fatalf("official works len = %d, want 1", len(profile.OfficialWorks))
	}
	work := profile.OfficialWorks[0]
	if work.WorkTitle != "2034杯精选分析" || work.MatchName != "2034杯小组赛" || work.AdoptionStatus != string(models.OfficialContentAdoptionOfficialPublished) {
		t.Fatalf("official work = %#v, want public official published work with match info", work)
	}

	works, err := service.GetAnalystOfficialWorks(analyst.ID, 1, 20)
	if err != nil {
		t.Fatalf("get official works: %v", err)
	}
	if len(works) != 1 || works[0].ID != visible.ID {
		t.Fatalf("official works = %#v, want only visible adoption", works)
	}
}
