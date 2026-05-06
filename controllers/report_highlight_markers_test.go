package controllers

import (
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReportHighlightMarkerTest(t *testing.T) (*gorm.DB, *ReportController) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Report{},
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db, NewReportController(nil, nil, db)
}

func TestReportHighlightMarkersOnlyExposeIncludedReadyClips(t *testing.T) {
	db, ctrl := setupReportHighlightMarkerTest(t)

	report := models.Report{
		OrderID:        1,
		UserID:         10,
		AnalystID:      20,
		PlayerName:     "Marker Report Player",
		PlayerPosition: "winger",
		Content:        "report",
		Status:         models.ReportStatusCompleted,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
	analysis := models.VideoAnalysis{
		OrderID:        report.OrderID,
		AnalystID:      report.AnalystID,
		UserID:         report.UserID,
		PlayerName:     report.PlayerName,
		PlayerPosition: report.PlayerPosition,
		Status:         models.AnalysisStatusCompleted,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	readyEndTime := 12000
	queuedEndTime := 26000
	markers := []models.AnalysisHighlight{
		{
			AnalysisID:      analysis.ID,
			Timestamp:       "00:05-00:12",
			MarkerType:      models.HighlightMarkerHighlight,
			Mode:            models.HighlightModeRange,
			StartTimeMs:     5000,
			EndTimeMs:       &readyEndTime,
			TagType:         models.HighlightGoal,
			Description:     "可公开片段。",
			VideoClipURL:    "http://localhost:8080/uploads/video-clips/ready.mp4",
			ClipStatus:      models.HighlightClipReady,
			IncludeInReport: true,
			SortOrder:       1,
		},
		{
			AnalysisID:      analysis.ID,
			Timestamp:       "00:20-00:26",
			MarkerType:      models.HighlightMarkerIssue,
			Mode:            models.HighlightModeRange,
			StartTimeMs:     20000,
			EndTimeMs:       &queuedEndTime,
			TagType:         models.HighlightTurnover,
			Description:     "未完成片段。",
			VideoClipURL:    "http://localhost:8080/uploads/video-clips/queued.mp4",
			ClipStatus:      models.HighlightClipQueued,
			IncludeInReport: true,
			SortOrder:       2,
		},
		{
			AnalysisID:      analysis.ID,
			Timestamp:       "00:30",
			MarkerType:      models.HighlightMarkerObservation,
			Mode:            models.HighlightModePoint,
			StartTimeMs:     30000,
			TagType:         models.HighlightTacticalNote,
			Description:     "不进报告的观察。",
			IncludeInReport: false,
			SortOrder:       3,
		},
	}
	if err := db.Create(&markers).Error; err != nil {
		t.Fatalf("create markers: %v", err)
	}
	if err := db.Model(&models.AnalysisHighlight{}).Where("description = ?", "不进报告的观察。").Update("include_in_report", false).Error; err != nil {
		t.Fatalf("force excluded marker flag: %v", err)
	}

	got, analysisID := ctrl.getReportHighlightMarkers(&report, models.RoleUser)
	if analysisID != analysis.ID {
		t.Fatalf("analysis id = %d, want %d", analysisID, analysis.ID)
	}
	if len(got) != 2 {
		t.Fatalf("markers len = %d, want 2", len(got))
	}
	if got[0].VideoClipURL == "" {
		t.Fatalf("ready clip url should be exposed")
	}
	if got[1].VideoClipURL != "" {
		t.Fatalf("queued clip url = %q, want empty", got[1].VideoClipURL)
	}
}

func TestReportHighlightMarkersHidePendingReportFromNonAdmin(t *testing.T) {
	db, ctrl := setupReportHighlightMarkerTest(t)

	report := models.Report{
		OrderID:        2,
		UserID:         11,
		AnalystID:      21,
		PlayerName:     "Pending Report Player",
		PlayerPosition: "midfielder",
		Content:        "report",
		Status:         models.ReportStatusProcessing,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
	analysis := models.VideoAnalysis{
		OrderID:   report.OrderID,
		AnalystID: report.AnalystID,
		UserID:    report.UserID,
		Status:    models.AnalysisStatusSubmitted,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	if err := db.Create(&models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:08",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModePoint,
		StartTimeMs:     8000,
		TagType:         models.HighlightPass,
		Description:     "待审核报告标记。",
		IncludeInReport: true,
	}).Error; err != nil {
		t.Fatalf("create marker: %v", err)
	}

	playerMarkers, _ := ctrl.getReportHighlightMarkers(&report, models.RoleUser)
	if len(playerMarkers) != 0 {
		t.Fatalf("player markers len = %d, want 0 before approval", len(playerMarkers))
	}

	adminMarkers, _ := ctrl.getReportHighlightMarkers(&report, models.RoleAdmin)
	if len(adminMarkers) != 1 {
		t.Fatalf("admin markers len = %d, want 1 for review", len(adminMarkers))
	}

	analystMarkers, _ := ctrl.getReportHighlightMarkers(&report, models.RoleAnalyst)
	if len(analystMarkers) != 1 {
		t.Fatalf("analyst markers len = %d, want 1 for own preview", len(analystMarkers))
	}
}
