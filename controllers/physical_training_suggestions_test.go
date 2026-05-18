package controllers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCollectTrainingSuggestionsFromPhysicalReports(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.PhysicalTestActivity{},
		&models.PhysicalTestRecord{},
		&models.PhysicalTestReport{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	testActivity := models.PhysicalTestActivity{
		ClubID:    1,
		Name:      "春季体测",
		StartDate: time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Template:  models.PTTemplateBasic,
		Status:    models.PTStatusReported,
	}
	if err := db.Create(&testActivity).Error; err != nil {
		t.Fatalf("create activity: %v", err)
	}

	reportData, _ := json.Marshal(models.PhysicalTestReportData{
		TrainingSuggestions: []string{
			"建议加强速度训练，每周3次冲刺练习",
			"建议每周至少2-3次专项体能训练",
			"建议加强速度训练，每周3次冲刺练习",
		},
	})
	report := models.PhysicalTestReport{
		RecordID:   1,
		PlayerID:   101,
		ClubID:     testActivity.ClubID,
		ActivityID: testActivity.ID,
		ReportData: string(reportData),
		ShareToken: "physical-suggestion-test",
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	ctrl := NewPhysicalTestController(services.NewPhysicalTestService(db))
	suggestions, playersCovered, reportsUsed := ctrl.collectTrainingSuggestions(&testActivity)

	if reportsUsed != 1 {
		t.Fatalf("reportsUsed = %d, want 1", reportsUsed)
	}
	if playersCovered != 1 {
		t.Fatalf("playersCovered = %d, want 1", playersCovered)
	}
	if len(suggestions) != 2 {
		t.Fatalf("suggestions = %#v, want two deduped suggestions", suggestions)
	}
	if suggestions[0] != "建议加强速度训练，每周3次冲刺练习" {
		t.Fatalf("first suggestion = %q", suggestions[0])
	}
}
