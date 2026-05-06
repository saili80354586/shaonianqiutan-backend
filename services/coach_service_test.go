package services

import (
	"errors"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCoachPlayerProgressTest(t *testing.T) *CoachService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Coach{},
		&models.Club{},
		&models.Team{},
		&models.TeamPlayer{},
		&models.TeamCoach{},
		&models.TrainingNote{},
		&models.Report{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	users := []models.User{
		{ID: 10, Phone: "player-a", Password: "x", Name: "球员 A", Role: models.RoleUser, Status: models.StatusActive},
		{ID: 11, Phone: "player-b", Password: "x", Name: "球员 B", Role: models.RoleUser, Status: models.StatusActive},
		{ID: 20, Phone: "coach-a", Password: "x", Name: "A 教练", Role: models.RoleCoach, Status: models.StatusActive},
		{ID: 21, Phone: "coach-b", Password: "x", Name: "B 教练", Role: models.RoleCoach, Status: models.StatusActive},
		{ID: 30, Phone: "club", Password: "x", Name: "俱乐部", Role: models.RoleClub, Status: models.StatusActive},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	if err := db.Create(&[]models.Coach{
		{ID: 200, UserID: 20},
		{ID: 201, UserID: 21},
	}).Error; err != nil {
		t.Fatalf("seed coaches: %v", err)
	}

	if err := db.Create(&models.Club{ID: 1, UserID: 30, Name: "测试俱乐部"}).Error; err != nil {
		t.Fatalf("seed club: %v", err)
	}
	if err := db.Create(&[]models.Team{
		{ID: 100, ClubID: 1, Name: "U12", AgeGroup: "U12", Status: models.TeamStatusActive},
		{ID: 101, ClubID: 1, Name: "U14", AgeGroup: "U14", Status: models.TeamStatusActive},
	}).Error; err != nil {
		t.Fatalf("seed teams: %v", err)
	}
	if err := db.Create(&[]models.TeamPlayer{
		{TeamID: 100, UserID: 10, Status: "active"},
		{TeamID: 101, UserID: 11, Status: "active"},
	}).Error; err != nil {
		t.Fatalf("seed team players: %v", err)
	}
	if err := db.Create(&[]models.TeamCoach{
		{TeamID: 100, UserID: 20, Role: models.CoachRoleHead, Status: "active"},
		{TeamID: 101, UserID: 21, Role: models.CoachRoleHead, Status: "active"},
	}).Error; err != nil {
		t.Fatalf("seed team coaches: %v", err)
	}

	if err := db.Create(&models.TrainingNote{
		CoachID:  200,
		PlayerID: 10,
		Title:    "训练记录",
		Content:  "完成训练",
		Category: "technical",
	}).Error; err != nil {
		t.Fatalf("seed training note: %v", err)
	}
	if err := db.Create(&models.Report{
		OrderID:    1,
		UserID:     10,
		AnalystID:  40,
		PlayerName: "球员 A",
		Content:    "报告内容",
		Status:     models.ReportStatusCompleted,
		CreatedAt:  time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed report: %v", err)
	}

	return NewCoachService(
		db,
		repositories.NewTeamRepository(db),
		repositories.NewWeeklyReportRepository(db),
		repositories.NewMatchSummaryRepository(db),
	)
}

func TestCoachPlayerProgressRequiresTeamAccess(t *testing.T) {
	service := setupCoachPlayerProgressTest(t)

	progress, err := service.GetPlayerProgress(200, 10)
	if err != nil {
		t.Fatalf("expected assigned coach to access player progress: %v", err)
	}
	if len(progress) != 2 {
		t.Fatalf("expected note and report progress entries, got %d", len(progress))
	}

	if _, err := service.GetPlayerProgress(201, 10); !errors.Is(err, ErrCoachPlayerAccessDenied) {
		t.Fatalf("expected other-team coach to be denied, got %v", err)
	}
}
