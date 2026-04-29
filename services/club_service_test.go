package services

import (
	"reflect"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newClubServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.PhysicalTestRecord{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestGetAbilityRadarUsesPhysicalRecords(t *testing.T) {
	db := newClubServiceTestDB(t)
	service := NewClubService(db)
	clubID := uint(10)
	otherClubID := uint(20)
	now := time.Now()

	sprint := 4.5
	jump := 200.0
	push := 30
	if err := db.Create(&models.PhysicalTestRecord{
		ActivityID:       1,
		PlayerID:         101,
		ClubID:           clubID,
		TestDate:         now,
		Sprint30m:        &sprint,
		StandingLongJump: &jump,
		PushUp:           &push,
	}).Error; err != nil {
		t.Fatalf("create club record: %v", err)
	}

	otherSprint := 5.6
	otherJump := 160.0
	otherPush := 12
	if err := db.Create(&models.PhysicalTestRecord{
		ActivityID:       2,
		PlayerID:         201,
		ClubID:           otherClubID,
		TestDate:         now,
		Sprint30m:        &otherSprint,
		StandingLongJump: &otherJump,
		PushUp:           &otherPush,
	}).Error; err != nil {
		t.Fatalf("create platform record: %v", err)
	}

	radar, err := service.GetAbilityRadar(clubID)
	if err != nil {
		t.Fatalf("GetAbilityRadar returned error: %v", err)
	}

	labels, _ := radar["labels"].([]string)
	teamAvg, _ := radar["teamAvg"].([]int)
	platformAvg, _ := radar["platformAvg"].([]int)

	if !reflect.DeepEqual(labels, []string{"速度", "力量", "爆发"}) {
		t.Fatalf("labels = %#v, want speed/strength/explosive labels", labels)
	}
	if reflect.DeepEqual(teamAvg, []int{75, 68, 72, 70, 65, 78}) {
		t.Fatalf("teamAvg still matches the old mock payload: %#v", teamAvg)
	}
	if len(teamAvg) != len(labels) || len(platformAvg) != len(labels) {
		t.Fatalf("radar lengths mismatch: labels=%d team=%d platform=%d", len(labels), len(teamAvg), len(platformAvg))
	}
	if teamAvg[0] <= platformAvg[0] {
		t.Fatalf("club speed score should beat platform average, got team=%d platform=%d", teamAvg[0], platformAvg[0])
	}
}

func TestGetAbilityRadarReturnsEmptyWithoutPhysicalRecords(t *testing.T) {
	db := newClubServiceTestDB(t)
	service := NewClubService(db)

	radar, err := service.GetAbilityRadar(999)
	if err != nil {
		t.Fatalf("GetAbilityRadar returned error: %v", err)
	}

	labels, _ := radar["labels"].([]string)
	teamAvg, _ := radar["teamAvg"].([]int)
	platformAvg, _ := radar["platformAvg"].([]int)

	if len(labels) != 0 || len(teamAvg) != 0 || len(platformAvg) != 0 {
		t.Fatalf("expected empty radar for no records, got labels=%#v team=%#v platform=%#v", labels, teamAvg, platformAvg)
	}
}
