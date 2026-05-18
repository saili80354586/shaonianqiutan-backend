package models

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserAdminListTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &LoginLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestUserRepositoryFindAdminUsersAppliesFiltersAndLastLogin(t *testing.T) {
	db := setupUserAdminListTestDB(t)
	repo := NewUserRepository(db)

	primary := User{
		Phone:        "13800001001",
		Password:     "x",
		Nickname:     "",
		Name:         "李小明",
		Role:         RoleAnalyst,
		Status:       StatusActive,
		City:         "上海市",
		Age:          12,
		CurrentTeam:  "上海少年队",
		JerseyColor:  "蓝色",
		JerseyNumber: 10,
		Position:     "前锋",
	}
	secondary := User{
		Phone:    "13800001002",
		Password: "x",
		Nickname: "王球探",
		Name:     "王小军",
		Role:     RoleScout,
		Status:   StatusInactive,
		City:     "北京市",
		Age:      15,
	}
	if err := db.Create(&primary).Error; err != nil {
		t.Fatalf("create primary user: %v", err)
	}
	if err := db.Create(&secondary).Error; err != nil {
		t.Fatalf("create secondary user: %v", err)
	}

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	if err := db.Create(&LoginLog{
		UserID:    primary.ID,
		Phone:     primary.Phone,
		Nickname:  primary.Name,
		Role:      string(primary.Role),
		Status:    "success",
		CreatedAt: now.Add(-2 * time.Hour),
	}).Error; err != nil {
		t.Fatalf("create first login log: %v", err)
	}
	if err := db.Create(&LoginLog{
		UserID:    primary.ID,
		Phone:     primary.Phone,
		Nickname:  primary.Name,
		Role:      string(primary.Role),
		Status:    "success",
		CreatedAt: now.Add(-30 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create second login log: %v", err)
	}
	if err := db.Create(&LoginLog{
		UserID:    secondary.ID,
		Phone:     secondary.Phone,
		Nickname:  secondary.Name,
		Role:      string(secondary.Role),
		Status:    "success",
		CreatedAt: now.Add(-90 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create secondary login log: %v", err)
	}

	ageMin := 12
	ageMax := 13
	list, total, err := repo.FindAdminUsers(1, 10, AdminUserListFilters{
		Keyword: "李",
		Role:    string(RoleAnalyst),
		Status:  string(StatusActive),
		City:    "上海",
		AgeMin:  &ageMin,
		AgeMax:  &ageMax,
	})
	if err != nil {
		t.Fatalf("find admin users: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}

	item := list[0]
	if item.DisplayName != "李小明" {
		t.Fatalf("display name = %q, want 李小明", item.DisplayName)
	}
	if item.Nickname != "" {
		t.Fatalf("nickname = %q, want empty to keep original data", item.Nickname)
	}
	if item.LastLoginAt == nil {
		t.Fatalf("last login = nil, want latest login time")
	}
	if !strings.Contains(*item.LastLoginAt, "11:30") {
		t.Fatalf("last login = %q, want latest login time", *item.LastLoginAt)
	}
	if item.CurrentTeam != "上海少年队" || item.JerseyColor != "蓝色" || item.JerseyNumber != 10 || item.Position != "前锋" {
		t.Fatalf("player fields not returned: %#v", item)
	}

	byID, total, err := repo.FindAdminUsers(1, 10, AdminUserListFilters{
		Keyword: strconv.FormatUint(uint64(primary.ID), 10),
		Role:    string(RoleAnalyst),
	})
	if err != nil {
		t.Fatalf("find admin users by id: %v", err)
	}
	if total != 1 || len(byID) != 1 || byID[0].ID != primary.ID {
		t.Fatalf("find by id total=%d list=%#v, want primary", total, byID)
	}
}
