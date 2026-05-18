package services

import (
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotifyTeamCalendarEventCreatesDedupedNotifications(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	service := NewNotificationService(db, repositories.NewNotificationRepository(db), models.NewUserRepository(db))
	err = service.NotifyTeamCalendarEvent(
		[]uint{101, 102, 101, 0},
		models.NotificationTypeTrainingPlanCreated,
		"新的训练计划",
		"团队战术训练 已加入球队日历，请按时参加",
		"training_plan",
		88,
		"/user-dashboard?tab=team_calendar",
		map[string]interface{}{"team_id": uint(12)},
	)
	if err != nil {
		t.Fatalf("notify calendar event: %v", err)
	}

	var notifications []models.Notification
	if err := db.Order("user_id ASC").Find(&notifications).Error; err != nil {
		t.Fatalf("find notifications: %v", err)
	}
	if len(notifications) != 2 {
		t.Fatalf("notification count = %d, want 2", len(notifications))
	}
	if notifications[0].UserID != 101 || notifications[1].UserID != 102 {
		t.Fatalf("unexpected notification users: %+v", []uint{notifications[0].UserID, notifications[1].UserID})
	}
	if notifications[0].Type != models.NotificationTypeTrainingPlanCreated {
		t.Fatalf("notification type = %s", notifications[0].Type)
	}

	data := notifications[0].GetData()
	if data == nil || data.TargetType != "training_plan" || data.TargetID != 88 || data.Link != "/user-dashboard?tab=team_calendar" {
		t.Fatalf("unexpected notification data: %+v", data)
	}
}
