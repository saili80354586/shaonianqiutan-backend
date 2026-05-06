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

func setupWeeklyReportAccessTest(t *testing.T) (*WeeklyReportService, *models.WeeklyReport) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.Team{},
		&models.TeamPlayer{},
		&models.TeamCoach{},
		&models.WeeklyReport{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	users := []models.User{
		{ID: 10, Phone: "player", Password: "x", Name: "球员", Role: models.RoleUser, Status: models.StatusActive},
		{ID: 12, Phone: "shared-player", Password: "x", Name: "共享球员", Role: models.RoleUser, Status: models.StatusActive},
		{ID: 20, Phone: "coach-a", Password: "x", Name: "A 教练", Role: models.RoleCoach, Status: models.StatusActive},
		{ID: 21, Phone: "coach-b", Password: "x", Name: "B 教练", Role: models.RoleCoach, Status: models.StatusActive},
		{ID: 30, Phone: "club", Password: "x", Name: "俱乐部", Role: models.RoleClub, Status: models.StatusActive},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	club := models.Club{ID: 1, UserID: 30, Name: "测试俱乐部"}
	teams := []models.Team{
		{ID: 100, ClubID: 1, Name: "U12", AgeGroup: "U12", Status: models.TeamStatusActive},
		{ID: 101, ClubID: 1, Name: "U14", AgeGroup: "U14", Status: models.TeamStatusActive},
	}
	if err := db.Create(&club).Error; err != nil {
		t.Fatalf("seed club: %v", err)
	}
	if err := db.Create(&teams).Error; err != nil {
		t.Fatalf("seed teams: %v", err)
	}
	if err := db.Create(&models.TeamPlayer{TeamID: 100, UserID: 10, Status: "active"}).Error; err != nil {
		t.Fatalf("seed team player: %v", err)
	}
	sharedPlayers := []models.TeamPlayer{
		{TeamID: 100, UserID: 12, Status: "active"},
		{TeamID: 101, UserID: 12, Status: "active"},
	}
	if err := db.Create(&sharedPlayers).Error; err != nil {
		t.Fatalf("seed shared team players: %v", err)
	}
	coaches := []models.TeamCoach{
		{TeamID: 100, UserID: 20, Role: models.CoachRoleHead, Status: "active"},
		{TeamID: 101, UserID: 21, Role: models.CoachRoleHead, Status: "active"},
	}
	if err := db.Create(&coaches).Error; err != nil {
		t.Fatalf("seed team coaches: %v", err)
	}

	now := time.Now()
	report := &models.WeeklyReport{
		ID:           900,
		TeamID:       100,
		PlayerID:     10,
		CoachID:      20,
		WeekStart:    now,
		WeekEnd:      now.AddDate(0, 0, 6),
		ReviewStatus: "pending",
		SubmitStatus: "submitted",
	}
	if err := db.Create(report).Error; err != nil {
		t.Fatalf("seed report: %v", err)
	}
	sharedReports := []models.WeeklyReport{
		{
			ID:           901,
			TeamID:       100,
			PlayerID:     12,
			CoachID:      20,
			WeekStart:    now,
			WeekEnd:      now.AddDate(0, 0, 6),
			ReviewStatus: "pending",
			SubmitStatus: "submitted",
		},
		{
			ID:           902,
			TeamID:       101,
			PlayerID:     12,
			CoachID:      21,
			WeekStart:    now,
			WeekEnd:      now.AddDate(0, 0, 6),
			ReviewStatus: "pending",
			SubmitStatus: "submitted",
		},
	}
	if err := db.Create(&sharedReports).Error; err != nil {
		t.Fatalf("seed shared reports: %v", err)
	}

	service := NewWeeklyReportService(
		db,
		repositories.NewWeeklyReportRepository(db),
		repositories.NewTeamRepository(db),
		models.NewUserRepository(db),
	)
	return service, report
}

func TestWeeklyReportDetailAccessByObjectOwner(t *testing.T) {
	service, report := setupWeeklyReportAccessTest(t)

	allowedUsers := []uint{10, 20, 30}
	for _, userID := range allowedUsers {
		got, err := service.GetByIDForUser(report.ID, userID)
		if err != nil {
			t.Fatalf("expected user %d to access report: %v", userID, err)
		}
		if got.ID != report.ID {
			t.Fatalf("expected report %d, got %d", report.ID, got.ID)
		}
	}

	if _, err := service.GetByIDForUser(report.ID, 21); !errors.Is(err, ErrWeeklyReportAccessDenied) {
		t.Fatalf("expected other-team coach to be denied, got %v", err)
	}
}

func TestListByPlayerForUserAccess(t *testing.T) {
	service, _ := setupWeeklyReportAccessTest(t)

	if reports, total, err := service.ListByPlayerForUser(10, 20, 1, 10); err != nil || total != 1 || len(reports) != 1 {
		t.Fatalf("expected team coach to list player reports, len=%d total=%d err=%v", len(reports), total, err)
	}
	if _, _, err := service.ListByPlayerForUser(10, 21, 1, 10); !errors.Is(err, ErrWeeklyReportAccessDenied) {
		t.Fatalf("expected other-team coach to be denied, got %v", err)
	}
}

func TestListByPlayerForUserFiltersSharedPlayerReportsByTeamAccess(t *testing.T) {
	service, _ := setupWeeklyReportAccessTest(t)

	reports, total, err := service.ListByPlayerForUser(12, 20, 1, 10)
	if err != nil {
		t.Fatalf("expected team coach to list shared player reports: %v", err)
	}
	if total != 1 || len(reports) != 1 {
		t.Fatalf("expected only one accessible report, len=%d total=%d", len(reports), total)
	}
	if reports[0].TeamID != 100 {
		t.Fatalf("expected only team 100 report, got team %d", reports[0].TeamID)
	}
}
