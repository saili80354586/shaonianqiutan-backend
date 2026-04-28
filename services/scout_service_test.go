package services

import (
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupScoutServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Scout{}, &models.ScoutReport{}, &models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestPublishScoutReportNotifiesPlayer(t *testing.T) {
	db := setupScoutServiceTestDB(t)
	now := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	scoutUser := models.User{ID: 8101, Phone: "13900008101", Password: "hashed", Role: models.RoleScout, CurrentRole: models.RoleScout, Status: models.StatusActive, Name: "测试球探", CreatedAt: now, UpdatedAt: now}
	player := models.User{ID: 8102, Phone: "13900008102", Password: "hashed", Role: models.RoleUser, CurrentRole: models.RoleUser, Status: models.StatusActive, Name: "测试球员", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&scoutUser).Error; err != nil {
		t.Fatalf("create scout user: %v", err)
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	scout := models.Scout{ID: 8201, UserID: scoutUser.ID, CurrentOrganization: "测试机构", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&scout).Error; err != nil {
		t.Fatalf("create scout: %v", err)
	}
	report := models.ScoutReport{ID: 8301, ScoutID: scout.ID, PlayerID: player.ID, Status: "draft", Summary: "边路潜力突出", Strengths: `["速度"]`, Weaknesses: `[]`, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	service := NewScoutService(db)
	published, err := service.PublishScoutReport(scoutUser.ID, report.ID)
	if err != nil {
		t.Fatalf("publish report: %v", err)
	}
	if published.Status != "published" || published.PublishedAt == nil {
		t.Fatalf("unexpected published report: %+v", published)
	}

	var notification models.Notification
	if err := db.Where("user_id = ? AND type = ?", player.ID, models.NotificationTypeScoutReport).First(&notification).Error; err != nil {
		t.Fatalf("find notification: %v", err)
	}
	data := notification.GetData()
	if data == nil || data.TargetType != "scout_report" || data.ReportID != report.ID {
		t.Fatalf("unexpected notification data: %+v", data)
	}
}
