package services

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
)

func TestAnalystLevelSubmitApplicationRejectsDuplicatePending(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	analyst := createOfficialTaskTestAnalyst(t, db, "等级申请分析师", "L1")

	req := AnalystLevelApplicationRequest{
		RequestedLevelCode: "L2",
		ApplicationReason:  "已完成多次样例分析，希望申请认证分析师",
	}
	if _, err := service.SubmitApplication(analyst.UserID, req, time.Now()); err != nil {
		t.Fatalf("submit application: %v", err)
	}
	if _, err := service.SubmitApplication(analyst.UserID, req, time.Now()); !errors.Is(err, ErrAnalystLevelApplicationConflict) {
		t.Fatalf("duplicate submit err = %v, want %v", err, ErrAnalystLevelApplicationConflict)
	}
}

func TestAnalystLevelReviewApprovedUpdatesAnalystLevel(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	analyst := createOfficialTaskTestAnalyst(t, db, "等级审核分析师", "L1")

	app, err := service.SubmitApplication(analyst.UserID, AnalystLevelApplicationRequest{
		RequestedLevelCode: "L3",
		ApplicationReason:  "有稳定交付经验",
	}, time.Now())
	if err != nil {
		t.Fatalf("submit application: %v", err)
	}

	reviewed, err := service.ReviewApplication(app.ID, 99, AnalystLevelReviewRequest{
		Status:     models.AnalystLevelApplicationApproved,
		ReviewNote: "同意升级",
	}, time.Now())
	if err != nil {
		t.Fatalf("review application: %v", err)
	}
	if reviewed.Status != models.AnalystLevelApplicationApproved || reviewed.ReviewedLevelCode != "L3" {
		t.Fatalf("reviewed status/level = %s/%s, want approved/L3", reviewed.Status, reviewed.ReviewedLevelCode)
	}

	var updated models.Analyst
	if err := db.First(&updated, analyst.ID).Error; err != nil {
		t.Fatalf("find analyst: %v", err)
	}
	if updated.LevelCode != "L3" || updated.LevelUpdatedBy != 99 || updated.LevelUpdatedAt == nil {
		t.Fatalf("analyst level/by/at = %s/%d/%v, want L3/99/non-nil", updated.LevelCode, updated.LevelUpdatedBy, updated.LevelUpdatedAt)
	}
}

func TestAnalystLevelReviewAdjustedAndRejected(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	adjustedAnalyst := createOfficialTaskTestAnalyst(t, db, "调整通过分析师", "L1")
	rejectedAnalyst := createOfficialTaskTestAnalyst(t, db, "驳回分析师", "L1")

	adjustedApp, err := service.SubmitApplication(adjustedAnalyst.UserID, AnalystLevelApplicationRequest{
		RequestedLevelCode: "L4",
		ApplicationReason:  "申请专家分析师",
	}, time.Now())
	if err != nil {
		t.Fatalf("submit adjusted application: %v", err)
	}
	if _, err := service.ReviewApplication(adjustedApp.ID, 88, AnalystLevelReviewRequest{
		Status:            models.AnalystLevelApplicationAdjusted,
		ReviewedLevelCode: "L2",
		ReviewNote:        "先调整为 L2",
	}, time.Now()); err != nil {
		t.Fatalf("review adjusted application: %v", err)
	}
	var adjusted models.Analyst
	if err := db.First(&adjusted, adjustedAnalyst.ID).Error; err != nil {
		t.Fatalf("find adjusted analyst: %v", err)
	}
	if adjusted.LevelCode != "L2" {
		t.Fatalf("adjusted level = %s, want L2", adjusted.LevelCode)
	}

	rejectedApp, err := service.SubmitApplication(rejectedAnalyst.UserID, AnalystLevelApplicationRequest{
		RequestedLevelCode: "L3",
		ApplicationReason:  "申请优选分析师",
	}, time.Now())
	if err != nil {
		t.Fatalf("submit rejected application: %v", err)
	}
	if _, err := service.ReviewApplication(rejectedApp.ID, 88, AnalystLevelReviewRequest{
		Status:     models.AnalystLevelApplicationRejected,
		ReviewNote: "资料不足",
	}, time.Now()); err != nil {
		t.Fatalf("review rejected application: %v", err)
	}
	var rejected models.Analyst
	if err := db.First(&rejected, rejectedAnalyst.ID).Error; err != nil {
		t.Fatalf("find rejected analyst: %v", err)
	}
	if rejected.LevelCode != "L1" {
		t.Fatalf("rejected analyst level = %s, want unchanged L1", rejected.LevelCode)
	}
}

func TestAnalystLevelManualSetUpdatesAnalyst(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	analyst := createOfficialTaskTestAnalyst(t, db, "手动定级分析师", "L1")

	if _, err := service.SetAnalystLevel(analyst.ID, 77, AnalystLevelSetRequest{LevelCode: "bad"}, time.Now()); !errors.Is(err, ErrAnalystLevelInvalid) {
		t.Fatalf("invalid level err = %v, want %v", err, ErrAnalystLevelInvalid)
	}

	updated, err := service.SetAnalystLevel(analyst.ID, 77, AnalystLevelSetRequest{
		LevelCode: "L4",
		Note:      "官方采用表现优秀",
	}, time.Now())
	if err != nil {
		t.Fatalf("set analyst level: %v", err)
	}
	if updated.LevelCode != "L4" || updated.LevelUpdatedBy != 77 || updated.LevelNote != "官方采用表现优秀" {
		t.Fatalf("updated analyst level/by/note = %s/%d/%s, want L4/77/note", updated.LevelCode, updated.LevelUpdatedBy, updated.LevelNote)
	}

	histories, err := service.ListLevelHistories(analyst.ID, 1, 20)
	if err != nil {
		t.Fatalf("list histories: %v", err)
	}
	if len(histories) == 0 || histories[0].FromLevelCode != "L1" || histories[0].ToLevelCode != "L4" {
		t.Fatalf("histories = %#v, want manual L1 -> L4", histories)
	}
}

func TestAnalystOfficialPartnershipCanBeSetAndCleared(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	analyst := createOfficialTaskTestAnalyst(t, db, "官方合作分析师", "L5")
	now := time.Now()

	partner, err := service.SetOfficialPartnership(analyst.ID, 77, AnalystOfficialPartnershipRequest{
		IsOfficialPartner:   true,
		PartnershipNote:     "连续官方采用，进入长期合作池",
		PartnershipBenefits: "重点赛事定向邀约；官方主页展示；周期结算",
	}, now)
	if err != nil {
		t.Fatalf("set official partnership: %v", err)
	}
	if !partner.IsOfficialPartner || partner.PartnershipStartedAt == nil || partner.PartnershipUpdatedBy != 77 {
		t.Fatalf("partner = %#v, want active partnership with admin", partner)
	}
	if partner.PartnershipBenefits != "重点赛事定向邀约；官方主页展示；周期结算" {
		t.Fatalf("benefits = %q, want saved benefits", partner.PartnershipBenefits)
	}

	cleared, err := service.SetOfficialPartnership(analyst.ID, 78, AnalystOfficialPartnershipRequest{
		IsOfficialPartner: false,
		PartnershipNote:   "暂停合作",
	}, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("clear official partnership: %v", err)
	}
	if cleared.IsOfficialPartner || cleared.PartnershipStartedAt != nil || cleared.PartnershipUpdatedBy != 78 {
		t.Fatalf("cleared = %#v, want inactive partnership", cleared)
	}
}

func TestAnalystGrowthSnapshotSuggestionAndActions(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewAnalystLevelService(db)
	now := time.Now()
	analyst := createOfficialTaskTestAnalyst(t, db, "成长分分析师", "L1")

	task := createOfficialTaskForTest(t, db, 5, "L1")
	var adoptedSubmissionID uint
	for i := 0; i < 3; i++ {
		submission := models.OfficialAnalysisSubmission{
			TaskID:                   task.ID,
			AnalystID:                analyst.ID,
			AcceptanceID:             uint(i + 1),
			VideoAuthorizationStatus: "authorized",
			Summary:                  "成长分采用版本",
			Status:                   models.OfficialAnalysisSubmissionAdopted,
			CreatedAt:                now,
			UpdatedAt:                now,
		}
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("create adopted submission: %v", err)
		}
		if adoptedSubmissionID == 0 {
			adoptedSubmissionID = submission.ID
		}
	}
	adoption := models.OfficialContentAdoption{
		TaskID:         task.ID,
		SubmissionID:   adoptedSubmissionID,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionOfficialPublished,
		Channel:        "douyin",
		WorkTitle:      "成长分官方发布",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.Create(&adoption).Error; err != nil {
		t.Fatalf("create adoption: %v", err)
	}
	for i, playCount := range []int64{12000, 36000} {
		record := models.OfficialContentPublishRecord{
			AdoptionID:    adoption.ID,
			Channel:       "douyin",
			PublishTitle:  fmt.Sprintf("官方发布记录 %d", i+1),
			PlayCount:     playCount,
			LikeCount:     playCount / 20,
			CommentCount:  playCount / 300,
			ShareCount:    playCount / 600,
			FavoriteCount: playCount / 400,
			PublishedAt:   &now,
			MetricsAt:     &now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("create publish record: %v", err)
		}
	}
	if err := db.Model(&models.Analyst{}).Where("id = ?", analyst.ID).Updates(map[string]interface{}{
		"official_adoption_count": 3,
		"official_publish_count":  3,
		"official_material_count": 1,
		"rating":                  4.8,
		"review_count":            8,
	}).Error; err != nil {
		t.Fatalf("update analyst stats: %v", err)
	}
	completedAt := now
	for i := 0; i < 2; i++ {
		order := models.Order{
			UserID:      analyst.UserID,
			AnalystID:   &analyst.ID,
			OrderNo:     "GROWTH-ORDER-" + string(rune('A'+i)),
			Amount:      100,
			Status:      models.OrderStatusCompleted,
			CompletedAt: &completedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Create(&order).Error; err != nil {
			t.Fatalf("create completed order: %v", err)
		}
	}

	growth, err := service.RefreshGrowthSnapshot(analyst.ID, now)
	if err != nil {
		t.Fatalf("refresh growth: %v", err)
	}
	if growth.GrowthScore <= 0 || models.AnalystLevelRank(growth.SuggestedLevelCode) <= models.AnalystLevelRank("L1") {
		t.Fatalf("growth = %#v, want positive score and suggested upgrade", growth)
	}
	if !strings.Contains(growth.SuggestionReason, "累计播放 48000") {
		t.Fatalf("suggestion reason = %q, want publish metrics reference", growth.SuggestionReason)
	}
	analysts, _, err := service.ListAnalysts(1, 20)
	if err != nil {
		t.Fatalf("list analysts: %v", err)
	}
	if len(analysts) == 0 || analysts[0].OfficialPublishMetrics == nil || analysts[0].OfficialPublishMetrics.TotalPlayCount != 48000 || analysts[0].OfficialPublishMetrics.MaxPlayCount != 36000 {
		t.Fatalf("analyst publish metrics = %#v, want aggregated publish metrics", analysts)
	}

	applied, err := service.ApplyLevelSuggestion(analyst.ID, 66, AnalystLevelSuggestionActionRequest{Note: "采纳系统建议"}, now)
	if err != nil {
		t.Fatalf("apply suggestion: %v", err)
	}
	if applied.LevelCode != growth.SuggestedLevelCode {
		t.Fatalf("applied level = %s, want %s", applied.LevelCode, growth.SuggestedLevelCode)
	}

	histories, err := service.ListLevelHistories(analyst.ID, 1, 20)
	if err != nil {
		t.Fatalf("list histories: %v", err)
	}
	if len(histories) == 0 || histories[0].Source != "system_suggestion" || histories[0].Action != "applied" {
		t.Fatalf("histories = %#v, want applied system suggestion", histories)
	}
}

func TestOfficialTaskDailyLimitByLevel(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()
	analyst := createOfficialTaskTestAnalyst(t, db, "每日额度分析师", "L1")
	firstTask := createOfficialTaskForTest(t, db, 5, "L1")
	secondTask := createOfficialTaskForTest(t, db, 5, "L1")

	if _, err := service.AcceptTask(analyst.ID, firstTask.ID, now); err != nil {
		t.Fatalf("accept first task: %v", err)
	}
	if _, err := service.AcceptTask(analyst.ID, secondTask.ID, now); !errors.Is(err, ErrOfficialTaskDailyLimit) {
		t.Fatalf("accept second task err = %v, want %v", err, ErrOfficialTaskDailyLimit)
	}
}
