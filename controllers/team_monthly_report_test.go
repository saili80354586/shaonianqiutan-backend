package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMonthlyReportRate(t *testing.T) {
	if got := monthlyReportRate(2, 3); got != 66.7 {
		t.Fatalf("expected 66.7, got %.1f", got)
	}
	if got := monthlyReportRate(0, 0); got != 0 {
		t.Fatalf("expected zero rate for empty denominator, got %.1f", got)
	}
}

func TestMonthlyPlanExpectedPlayers(t *testing.T) {
	teamPlan := models.TrainingPlan{}
	if got := monthlyPlanExpectedPlayers(teamPlan, 18); got != 18 {
		t.Fatalf("expected full team player count, got %d", got)
	}

	targetedPlan := models.TrainingPlan{PlayerIDs: "[101,102,103]"}
	if got := monthlyPlanExpectedPlayers(targetedPlan, 18); got != 3 {
		t.Fatalf("expected targeted player count, got %d", got)
	}
}

func TestBuildMonthlyTrainingRecommendations(t *testing.T) {
	recommendations := buildMonthlyTrainingRecommendations(monthlyRecommendationInput{
		TrainingCount:            8,
		CompletedTrainingCount:   5,
		ReviewCount:              2,
		ActivePlayerCount:        18,
		ExpectedAttendance:       144,
		PresentAttendance:        95,
		LateAttendance:           3,
		UnmarkedAttendance:       12,
		MatchCount:               3,
		PendingMatchSummaryCount: 1,
		Wins:                     0,
		Draws:                    1,
		Losses:                   2,
		PhysicalTestCount:        0,
		WeeklyTotalPlayers:       72,
		WeeklySubmittedCount:     50,
		WeeklyReviewedCount:      30,
		CompletionStatusCount: map[string]int{
			"poor":   1,
			"normal": 2,
		},
	})

	if len(recommendations) < 5 {
		t.Fatalf("expected multiple recommendations, got %d", len(recommendations))
	}

	hasHighAttendance := false
	hasPhysical := false
	for _, recommendation := range recommendations {
		if recommendation["category"] == "attendance" && recommendation["priority"] == "high" {
			hasHighAttendance = true
		}
		if recommendation["category"] == "physical" {
			hasPhysical = true
		}
	}
	if !hasHighAttendance {
		t.Fatalf("expected high attendance recommendation")
	}
	if !hasPhysical {
		t.Fatalf("expected physical baseline recommendation")
	}
}

func TestBuildMonthlyTrainingRecommendationsStableFallback(t *testing.T) {
	recommendations := buildMonthlyTrainingRecommendations(monthlyRecommendationInput{
		TrainingCount:          8,
		CompletedTrainingCount: 8,
		ReviewCount:            8,
		ActivePlayerCount:      18,
		ExpectedAttendance:     144,
		PresentAttendance:      140,
		LateAttendance:         4,
		PhysicalTestCount:      1,
		PhysicalRecordCount:    18,
		WeeklyTotalPlayers:     72,
		WeeklySubmittedCount:   72,
		WeeklyReviewedCount:    72,
		CompletionStatusCount:  map[string]int{},
	})

	if len(recommendations) != 1 {
		t.Fatalf("expected stable fallback recommendation, got %d", len(recommendations))
	}
	if recommendations[0]["priority"] != "low" {
		t.Fatalf("expected low priority fallback, got %v", recommendations[0]["priority"])
	}
}

func TestBuildMonthlyAITrainingInsights(t *testing.T) {
	insights := buildMonthlyAITrainingInsights(monthlyRecommendationInput{
		TrainingCount:            8,
		CompletedTrainingCount:   5,
		ReviewCount:              2,
		ExpectedAttendance:       144,
		PresentAttendance:        128,
		LateAttendance:           4,
		PendingMatchSummaryCount: 1,
		WeeklyTotalPlayers:       72,
		WeeklySubmittedCount:     58,
		CompletionStatusCount:    map[string]int{},
	})

	if len(insights) == 0 {
		t.Fatalf("expected ai insights")
	}
	if insights[0]["confidence"] == nil || insights[0]["action"] == "" {
		t.Fatalf("expected confidence and action, got %#v", insights[0])
	}
}

func TestGetTeamMonthlyReportArchive(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.TeamMonthlyReportArchive{},
		&models.TeamMonthlyReportArchiveVersion{},
		&models.TeamMonthlyReportArchiveReviewEvent{},
		&models.TeamMonthlyReportArchiveAdjustmentItem{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&models.User{
		ID:          7,
		Phone:       "13800000007",
		Password:    "password",
		Nickname:    "周运营",
		Role:        models.RoleClub,
		CurrentRole: models.RoleClub,
		Status:      models.StatusActive,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.User{
		ID:          8,
		Phone:       "13800000008",
		Password:    "password",
		Nickname:    "李教练",
		Role:        models.RoleCoach,
		CurrentRole: models.RoleCoach,
		Status:      models.StatusActive,
	}).Error; err != nil {
		t.Fatalf("create reviewer: %v", err)
	}

	payload := gin.H{
		"teamId":   133,
		"teamName": "U12 精英队",
		"month":    "2026-05",
		"label":    "v2",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	firstPayload := gin.H{
		"teamId":   133,
		"teamName": "U12 精英队",
		"month":    "2026-05",
		"label":    "v1",
	}
	firstRaw, err := json.Marshal(firstPayload)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	archive := models.TeamMonthlyReportArchive{
		TeamID:     133,
		Month:      "2026-05",
		Version:    2,
		Snapshot:   string(raw),
		ArchivedBy: 7,
		ArchivedAt: time.Now(),
	}
	if err := db.Create(&archive).Error; err != nil {
		t.Fatalf("create archive: %v", err)
	}
	if err := db.Create(&models.TeamMonthlyReportArchiveVersion{
		ArchiveID:  archive.ID,
		TeamID:     archive.TeamID,
		Month:      archive.Month,
		Version:    1,
		Snapshot:   string(firstRaw),
		ArchivedBy: 6,
		ArchivedAt: archive.ArchivedAt.Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create first archive version: %v", err)
	}
	if err := db.Create(&models.TeamMonthlyReportArchiveVersion{
		ArchiveID:  archive.ID,
		TeamID:     archive.TeamID,
		Month:      archive.Month,
		Version:    2,
		Snapshot:   string(raw),
		ArchivedBy: 7,
		ArchivedAt: archive.ArchivedAt,
	}).Error; err != nil {
		t.Fatalf("create second archive version: %v", err)
	}

	ctrl := &TeamController{db: db}
	gotArchive, gotPayload, err := ctrl.getTeamMonthlyReportArchive(133, "2026-05")
	if err != nil {
		t.Fatalf("get archive: %v", err)
	}
	if gotArchive == nil {
		t.Fatalf("expected archive")
	}
	if gotPayload["month"] != "2026-05" {
		t.Fatalf("expected month payload, got %#v", gotPayload)
	}
	if gotArchive.Version != 2 {
		t.Fatalf("expected version 2, got %d", gotArchive.Version)
	}
	versions := ctrl.monthlyReportArchiveVersionSummaries(gotArchive)
	if len(versions) != 2 {
		t.Fatalf("expected 2 version summaries, got %#v", versions)
	}
	if versions[0]["version"] != 2 {
		t.Fatalf("expected newest version first, got %#v", versions)
	}
	userSummary, ok := versions[0]["archivedByUser"].(gin.H)
	if !ok {
		t.Fatalf("expected archived user summary, got %#v", versions[0])
	}
	if userSummary["displayName"] != "周运营" || userSummary["role"] != models.RoleClub {
		t.Fatalf("unexpected archived user summary: %#v", userSummary)
	}
	reviewSummary, ok := versions[0]["review"].(gin.H)
	if !ok || reviewSummary["status"] != "pending" {
		t.Fatalf("expected pending review summary, got %#v", versions[0])
	}
	versionArchive, versionPayload, err := ctrl.getTeamMonthlyReportArchiveVersion(archive.ID, 1)
	if err != nil {
		t.Fatalf("get archive version: %v", err)
	}
	if versionArchive == nil {
		t.Fatalf("expected archive version")
	}
	if versionPayload["label"] != "v1" {
		t.Fatalf("expected first version payload, got %#v", versionPayload)
	}

	missingArchive, missingPayload, err := ctrl.getTeamMonthlyReportArchive(133, "2026-04")
	if err != nil {
		t.Fatalf("missing archive should not error: %v", err)
	}
	if missingArchive != nil || missingPayload != nil {
		t.Fatalf("expected nil missing archive, got %#v %#v", missingArchive, missingPayload)
	}

	body := bytes.NewBufferString(`{"month":"2026-05","version":2,"status":"confirmed","note":"教练组已确认"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/teams/133/monthly-report/archive/review", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "teamId", Value: "133"}}
	ctx.Set("userId", uint(8))
	ctrl.ReviewTeamMonthlyReportArchive(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d %s", w.Code, w.Body.String())
	}

	var reviewed models.TeamMonthlyReportArchiveVersion
	if err := db.Where("archive_id = ? AND version = ?", archive.ID, 2).First(&reviewed).Error; err != nil {
		t.Fatalf("get reviewed version: %v", err)
	}
	if reviewed.ReviewStatus != "confirmed" || reviewed.ReviewNote != "教练组已确认" {
		t.Fatalf("unexpected review fields: %#v", reviewed)
	}
	if reviewed.ReviewedBy == nil || *reviewed.ReviewedBy != 8 || reviewed.ReviewedAt == nil {
		t.Fatalf("expected reviewer metadata, got %#v", reviewed)
	}
	var reviewEvents []models.TeamMonthlyReportArchiveReviewEvent
	if err := db.Where("version_id = ?", reviewed.ID).Order("created_at ASC").Find(&reviewEvents).Error; err != nil {
		t.Fatalf("get review events: %v", err)
	}
	if len(reviewEvents) != 1 || reviewEvents[0].Status != "confirmed" || reviewEvents[0].ActorID != 8 {
		t.Fatalf("unexpected review events after confirm: %#v", reviewEvents)
	}

	body = bytes.NewBufferString(`{"month":"2026-05","version":2,"status":"needs_revision","note":"需补充重点球员跟进","adjustmentItems":["需补充重点球员跟进","复核复盘标题：月报测试训练复盘"]}`)
	req = httptest.NewRequest(http.MethodPut, "/api/teams/133/monthly-report/archive/review", body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "teamId", Value: "133"}}
	ctx.Set("userId", uint(8))
	ctrl.ReviewTeamMonthlyReportArchive(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected needs revision success, got %d %s", w.Code, w.Body.String())
	}

	if err := db.Where("archive_id = ? AND version = ?", archive.ID, 2).First(&reviewed).Error; err != nil {
		t.Fatalf("get needs revision version: %v", err)
	}
	if reviewed.ReviewStatus != "needs_revision" || reviewed.ReviewNote != "需补充重点球员跟进" {
		t.Fatalf("unexpected needs revision fields: %#v", reviewed)
	}
	var adjustmentItems []models.TeamMonthlyReportArchiveAdjustmentItem
	if err := db.Where("version_id = ?", reviewed.ID).Order("created_at ASC").Find(&adjustmentItems).Error; err != nil {
		t.Fatalf("get adjustment items: %v", err)
	}
	if len(adjustmentItems) != 2 || adjustmentItems[0].Status != "open" {
		t.Fatalf("unexpected adjustment items: %#v", adjustmentItems)
	}

	body = bytes.NewBufferString(`{"status":"completed"}`)
	req = httptest.NewRequest(http.MethodPut, "/api/teams/133/monthly-report/archive/adjustments/1", body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "teamId", Value: "133"}, {Key: "itemId", Value: "1"}}
	ctx.Set("userId", uint(8))
	ctrl.UpdateTeamMonthlyReportArchiveAdjustment(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected adjustment update success, got %d %s", w.Code, w.Body.String())
	}
	var completedItem models.TeamMonthlyReportArchiveAdjustmentItem
	if err := db.First(&completedItem, adjustmentItems[0].ID).Error; err != nil {
		t.Fatalf("get completed item: %v", err)
	}
	if completedItem.Status != "completed" || completedItem.CompletedBy == nil || *completedItem.CompletedBy != 8 || completedItem.CompletedAt == nil {
		t.Fatalf("unexpected completed item: %#v", completedItem)
	}

	body = bytes.NewBufferString(`{"month":"2026-05","version":2,"status":"revision_submitted","note":"已补充重点球员跟进说明"}`)
	req = httptest.NewRequest(http.MethodPut, "/api/teams/133/monthly-report/archive/review", body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "teamId", Value: "133"}}
	ctx.Set("userId", uint(8))
	ctrl.ReviewTeamMonthlyReportArchive(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected revision submitted success, got %d %s", w.Code, w.Body.String())
	}

	if err := db.Where("archive_id = ? AND version = ?", archive.ID, 2).First(&reviewed).Error; err != nil {
		t.Fatalf("get revised version: %v", err)
	}
	if reviewed.ReviewStatus != "revision_submitted" || reviewed.ReviewNote != "已补充重点球员跟进说明" {
		t.Fatalf("unexpected revision fields: %#v", reviewed)
	}
	if err := db.Where("version_id = ?", reviewed.ID).Order("created_at ASC").Find(&reviewEvents).Error; err != nil {
		t.Fatalf("get revision events: %v", err)
	}
	if len(reviewEvents) != 3 || reviewEvents[2].Status != "revision_submitted" || reviewEvents[2].Note != "已补充重点球员跟进说明" {
		t.Fatalf("unexpected review events after revision: %#v", reviewEvents)
	}
	reviewSummary = ctrl.monthlyReportArchiveReviewSummary(&reviewed)
	history, ok := reviewSummary["history"].([]gin.H)
	if !ok || len(history) != 3 {
		t.Fatalf("expected review history summary, got %#v", reviewSummary)
	}
	adjustments, ok := reviewSummary["adjustments"].([]gin.H)
	if !ok || len(adjustments) != 2 {
		t.Fatalf("expected adjustment summary, got %#v", reviewSummary)
	}
}
