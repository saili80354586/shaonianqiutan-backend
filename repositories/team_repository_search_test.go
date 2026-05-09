package repositories

import (
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSearchUsersFindsCoachFromUserRoles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserRoleRecord{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	user := models.User{
		Phone:    "13932009001",
		Password: "test",
		Name:     "兼任教练管理员",
		Role:     models.RoleClub,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.UserRoleRecord{
		UserID:        user.ID,
		Role:          models.RoleCoach,
		Status:        "active",
		Source:        "test",
		PublicVisible: true,
	}).Error; err != nil {
		t.Fatalf("create role record: %v", err)
	}

	repo := NewTeamRepository(db)
	users, err := repo.SearchUsers("13932009001", "coach")
	if err != nil {
		t.Fatalf("search users: %v", err)
	}
	if len(users) != 1 || users[0].ID != user.ID {
		t.Fatalf("users = %#v, want club user with active coach role", users)
	}
}
