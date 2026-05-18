package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminOrderProgressTestService(t *testing.T) (*gorm.DB, *AdminService, models.User, models.Analyst) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-order-progress.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Order{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
		&models.VideoAnalysis{},
		&models.AnalysisHighlight{},
		&models.Report{},
		&models.ReportVersion{},
		&models.AnalysisOperationEvent{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	player := models.User{Phone: "13920000001", Password: "test", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	analystUser := models.User{Phone: "13920000002", Password: "test", Role: models.RoleAnalyst, Status: models.StatusActive}
	if err := db.Create(&analystUser).Error; err != nil {
		t.Fatalf("create analyst user: %v", err)
	}
	analyst := models.Analyst{UserID: analystUser.ID, Name: "测试分析师", Status: models.AnalystStatusActive}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	service := NewAdminService(
		models.NewUserRepository(db),
		models.NewReportRepository(db),
		models.NewOrderRepository(db),
		models.NewAnalystRepository(db),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		models.NewVideoAnalysisRepository(db),
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
	)
	return db, service, player, analyst
}

func TestGetOrderAnalysisProgressDetailWaitingDispatch(t *testing.T) {
	db, service, player, _ := setupAdminOrderProgressTestService(t)
	order := models.Order{
		UserID:    player.ID,
		OrderNo:   "ORD-PROGRESS-001",
		Amount:    299,
		Status:    models.OrderStatusUploaded,
		OrderType: "basic",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	detail, err := service.GetOrderAnalysisProgressDetail(order.ID)
	if err != nil {
		t.Fatalf("GetOrderAnalysisProgressDetail: %v", err)
	}
	if detail.AnalysisProgress.Stage != "waiting_dispatch" {
		t.Fatalf("stage = %s, want waiting_dispatch", detail.AnalysisProgress.Stage)
	}
	if detail.Completion.ScoreOverview.CompletedCount != 0 {
		t.Fatalf("completed scores = %d, want 0", detail.Completion.ScoreOverview.CompletedCount)
	}
}

func TestGetOrderAnalysisProgressDetailScoreDimensionsUseEvents(t *testing.T) {
	db, service, player, analyst := setupAdminOrderProgressTestService(t)
	now := time.Now().Add(-2 * time.Hour)
	deadline := time.Now().Add(24 * time.Hour)
	order := models.Order{
		UserID:     player.ID,
		AnalystID:  &analyst.ID,
		OrderNo:    "ORD-PROGRESS-002",
		Amount:     799,
		Status:     models.OrderStatusProcessing,
		OrderType:  "pro",
		AssignedAt: &now,
		AcceptedAt: &now,
		Deadline:   &deadline,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	scores := models.NewDefaultScores()
	scores.BallControl.Score = 8.2
	scores.BallControl.Comment = "控球稳定，面对逼抢时能保持节奏。"
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("scores json: %v", err)
	}
	analysis := models.VideoAnalysis{
		OrderID:      order.ID,
		AnalystID:    analyst.ID,
		UserID:       player.ID,
		PlayerName:   "测试球员",
		OverallScore: scores.CalculateOverallScore(),
		Scores:       scoresJSON,
		Summary:      "综合评价已填写，具备较好的边路推进能力。",
		Status:       models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	eventTime := time.Now().Add(-30 * time.Minute)
	if err := db.Create(&models.AnalysisOperationEvent{
		OrderID:      order.ID,
		AnalysisID:   analysis.ID,
		AnalystID:    analyst.ID,
		EventType:    "score_dimension_updated",
		Section:      "score",
		FieldKey:     "ball_control",
		FieldLabel:   "控球能力",
		AfterSummary: "分数 8.2，评语 17 字",
		CreatedAt:    eventTime,
	}).Error; err != nil {
		t.Fatalf("create event: %v", err)
	}

	detail, err := service.GetOrderAnalysisProgressDetail(order.ID)
	if err != nil {
		t.Fatalf("GetOrderAnalysisProgressDetail: %v", err)
	}
	if detail.AnalysisProgress.Stage != "scoring" {
		t.Fatalf("stage = %s, want scoring", detail.AnalysisProgress.Stage)
	}
	if detail.Completion.ScoreOverview.CompletedCount != 1 {
		t.Fatalf("completed scores = %d, want 1", detail.Completion.ScoreOverview.CompletedCount)
	}
	var ballControl *AdminScoreDimensionDTO
	for _, group := range detail.Completion.ScoreGroups {
		for i := range group.Items {
			if group.Items[i].FieldKey == "ball_control" {
				ballControl = &group.Items[i]
			}
		}
	}
	if ballControl == nil {
		t.Fatalf("ball_control dimension not found")
	}
	if ballControl.Status != "completed" {
		t.Fatalf("ball_control status = %s, want completed", ballControl.Status)
	}
	if ballControl.LastUpdatedAt == nil {
		t.Fatalf("ball_control last_updated_at is nil")
	}
	if ballControl.UpdateCount != 1 {
		t.Fatalf("ball_control update_count = %d, want 1", ballControl.UpdateCount)
	}
}

func TestBuildAdminScoreProgressNormalizesHundredPointScores(t *testing.T) {
	scores := models.NewDefaultScores()
	scores.BallControl.Score = 82
	scores.BallControl.Comment = "控球稳定，处理球更主动。"
	scores.OffBallMovement.Score = 70
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("scores json: %v", err)
	}

	overview, groups := buildAdminScoreProgress(&models.VideoAnalysis{Scores: scoresJSON}, map[string]operationEventStats{})
	if overview.CompletedCount != 1 {
		t.Fatalf("completed scores = %d, want 1", overview.CompletedCount)
	}
	if overview.NotStartedCount != 19 {
		t.Fatalf("not started scores = %d, want 19", overview.NotStartedCount)
	}

	dimensions := map[string]AdminScoreDimensionDTO{}
	for _, group := range groups {
		for _, item := range group.Items {
			dimensions[item.FieldKey] = item
		}
	}
	if dimensions["ball_control"].Score != 8.2 {
		t.Fatalf("ball_control score = %.1f, want 8.2", dimensions["ball_control"].Score)
	}
	if dimensions["off_ball_movement"].Score != 7.0 {
		t.Fatalf("off_ball_movement score = %.1f, want 7.0", dimensions["off_ball_movement"].Score)
	}
	if dimensions["off_ball_movement"].Status != "not_started" {
		t.Fatalf("off_ball_movement status = %s, want not_started", dimensions["off_ball_movement"].Status)
	}
}

func TestBuildAdminScoreProgressTreatsExplicitSevenWithCommentAsCompleted(t *testing.T) {
	scores := models.NewDefaultScores()
	scores.PressingAwareness.Score = 7
	scores.PressingAwareness.Comment = "防守积极，多通过预判抢夺球权。"
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		t.Fatalf("scores json: %v", err)
	}

	eventStats := map[string]operationEventStats{
		"pressing_awareness": {count: 1},
	}

	overview, groups := buildAdminScoreProgress(&models.VideoAnalysis{Scores: scoresJSON}, eventStats)
	if overview.CompletedCount != 1 {
		t.Fatalf("completed scores = %d, want 1", overview.CompletedCount)
	}
	if overview.CommentOnlyCount != 0 {
		t.Fatalf("comment only scores = %d, want 0", overview.CommentOnlyCount)
	}

	dimensions := map[string]AdminScoreDimensionDTO{}
	for _, group := range groups {
		for _, item := range group.Items {
			dimensions[item.FieldKey] = item
		}
	}
	if dimensions["pressing_awareness"].Status != "completed" {
		t.Fatalf("pressing_awareness status = %s, want completed", dimensions["pressing_awareness"].Status)
	}
}
