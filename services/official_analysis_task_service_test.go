package services

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOfficialAnalysisTaskTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.AnalystLevel{},
		&models.AnalystLevelApplication{},
		&models.AnalystGrowthSnapshot{},
		&models.AnalystLevelHistory{},
		&models.OfficialAnalysisTask{},
		&models.OfficialAnalysisTaskAcceptance{},
		&models.OfficialAnalysisSubmission{},
		&models.OfficialContentAdoption{},
		&models.OfficialContentPublishRecord{},
		&models.OfficialEventTopicConfig{},
		&models.AnalystRewardSettlementBatch{},
		&models.AnalystRewardRecord{},
		&models.Order{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func createOfficialTaskTestAnalyst(t *testing.T, db *gorm.DB, name, level string) models.Analyst {
	t.Helper()

	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	user := models.User{
		Phone:    fmt.Sprintf("139%08d", userCount+1),
		Password: "test",
		Name:     name,
		Nickname: name,
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	analyst := models.Analyst{
		UserID:    user.ID,
		Name:      name,
		LevelCode: level,
		Status:    models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}
	return analyst
}

func createOfficialTaskForTest(t *testing.T, db *gorm.DB, maxAcceptCount int, visibleLevel string) models.OfficialAnalysisTask {
	t.Helper()

	var taskCount int64
	if err := db.Model(&models.OfficialAnalysisTask{}).Count(&taskCount).Error; err != nil {
		t.Fatalf("count official tasks: %v", err)
	}
	deadline := time.Now().Add(24 * time.Hour)
	task := models.OfficialAnalysisTask{
		TaskNo:              fmt.Sprintf("OFF-%06d", taskCount+1),
		Title:               "2034杯 U12 官方选题",
		MatchName:           "2034杯",
		AgeGroup:            "U12",
		AuthorizationStatus: "authorized",
		TaskType:            "composite",
		MaxAcceptCount:      maxAcceptCount,
		VisibleLevelMin:     visibleLevel,
		Deadline:            &deadline,
		Status:              models.OfficialAnalysisTaskPublished,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create official task: %v", err)
	}
	return task
}

func TestSeedDefaultAnalystLevelsIsIdempotent(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)

	if err := models.SeedDefaultAnalystLevels(db); err != nil {
		t.Fatalf("seed default analyst levels: %v", err)
	}
	if err := models.SeedDefaultAnalystLevels(db); err != nil {
		t.Fatalf("seed default analyst levels second run: %v", err)
	}

	var count int64
	if err := db.Model(&models.AnalystLevel{}).Count(&count).Error; err != nil {
		t.Fatalf("count levels: %v", err)
	}
	if count != int64(len(models.DefaultAnalystLevels())) {
		t.Fatalf("level count = %d, want %d", count, len(models.DefaultAnalystLevels()))
	}
}

func TestAnalystLevelDefaultsToL1(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)

	user := models.User{Phone: "13900000001", Password: "test", Name: "默认等级分析师", Role: models.RoleAnalyst, Status: models.StatusActive}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analyst := models.Analyst{UserID: user.ID, Name: "默认等级分析师", Status: models.AnalystStatusActive}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	var found models.Analyst
	if err := db.First(&found, analyst.ID).Error; err != nil {
		t.Fatalf("find analyst: %v", err)
	}
	if found.LevelCode != models.DefaultAnalystLevelCode {
		t.Fatalf("level code = %q, want %q", found.LevelCode, models.DefaultAnalystLevelCode)
	}
}

func TestCreateAndPublishTaskRequiresAuthorizationStatus(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()
	player := models.User{Phone: "13918881111", Password: "test", Name: "发布测试球员", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	task, err := service.CreateTask(OfficialAnalysisTaskRequest{
		Title:           "2034杯 官方选题",
		MaxAcceptCount:  2,
		VisibleLevelMin: "L1",
	}, 7, now)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != models.OfficialAnalysisTaskDraft {
		t.Fatalf("status = %s, want draft", task.Status)
	}

	if _, err := service.PublishTask(task.ID, 7, now); !errors.Is(err, ErrOfficialTaskInvalid) {
		t.Fatalf("publish without authorization err = %v, want %v", err, ErrOfficialTaskInvalid)
	}

	updated, err := service.UpdateTask(task.ID, OfficialAnalysisTaskRequest{
		Title:                "2034杯 官方选题",
		MaxAcceptCount:       2,
		VisibleLevelMin:      "L1",
		AuthorizationStatus:  "authorized",
		TargetPlayerUserID:   player.ID,
		VideoFirstHalfURL:    "https://example.com/first.mp4",
		TargetPlayerName:     "发布测试球员",
		TargetPlayerTeam:     "测试队",
		TargetJerseyColor:    "红色",
		TargetJerseyNumber:   "9",
		TargetPlayerPosition: "前锋",
	}, now)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if updated.AuthorizationStatus != "authorized" {
		t.Fatalf("authorization_status = %s, want authorized", updated.AuthorizationStatus)
	}

	published, err := service.PublishTask(task.ID, 7, now)
	if err != nil {
		t.Fatalf("publish with authorization: %v", err)
	}
	if published.Status != models.OfficialAnalysisTaskPublished {
		t.Fatalf("status = %s, want published", published.Status)
	}
}

func TestListAvailableTasksFiltersByAnalystLevelAndExistingAcceptance(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)

	l2Task := createOfficialTaskForTest(t, db, 3, "L2")
	createOfficialTaskForTest(t, db, 3, "L4")
	analyst := createOfficialTaskTestAnalyst(t, db, "L3分析师", "L3")

	tasks, total, err := service.ListAvailableTasksForAnalyst(analyst.ID, 1, 20, time.Now())
	if err != nil {
		t.Fatalf("list available tasks: %v", err)
	}
	if total != 1 || len(tasks) != 1 || tasks[0].ID != l2Task.ID {
		t.Fatalf("available tasks = %#v total=%d, want only L2 task", tasks, total)
	}

	if _, err := service.AcceptTask(analyst.ID, l2Task.ID, time.Now()); err != nil {
		t.Fatalf("accept L2 task: %v", err)
	}

	tasks, total, err = service.ListAvailableTasksForAnalyst(analyst.ID, 1, 20, time.Now())
	if err != nil {
		t.Fatalf("list available after accept: %v", err)
	}
	if total != 0 || len(tasks) != 0 {
		t.Fatalf("available tasks after accept = %#v total=%d, want empty", tasks, total)
	}
}

func TestListAvailableTasksHonorsPriorityWindow(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 3, "L1")
	priorityUntil := now.Add(time.Hour)
	if err := db.Model(&models.OfficialAnalysisTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"priority_level_min": "L3",
		"priority_until":     &priorityUntil,
	}).Error; err != nil {
		t.Fatalf("update priority window: %v", err)
	}
	l2 := createOfficialTaskTestAnalyst(t, db, "优先窗口外分析师", "L2")
	l3 := createOfficialTaskTestAnalyst(t, db, "优先窗口内分析师", "L3")

	tasks, total, err := service.ListAvailableTasksForAnalyst(l2.ID, 1, 20, now)
	if err != nil {
		t.Fatalf("list L2 available tasks: %v", err)
	}
	if total != 0 || len(tasks) != 0 {
		t.Fatalf("L2 available tasks = %#v total=%d, want empty during priority window", tasks, total)
	}
	if _, err := service.AcceptTask(l2.ID, task.ID, now); !errors.Is(err, ErrOfficialTaskLevelDenied) {
		t.Fatalf("L2 accept during priority window err = %v, want %v", err, ErrOfficialTaskLevelDenied)
	}

	tasks, total, err = service.ListAvailableTasksForAnalyst(l3.ID, 1, 20, now)
	if err != nil {
		t.Fatalf("list L3 available tasks: %v", err)
	}
	if total != 1 || len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("L3 available tasks = %#v total=%d, want priority task", tasks, total)
	}

	tasks, total, err = service.ListAvailableTasksForAnalyst(l2.ID, 1, 20, priorityUntil.Add(time.Minute))
	if err != nil {
		t.Fatalf("list L2 after priority window: %v", err)
	}
	if total != 1 || len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("L2 available after priority window = %#v total=%d, want task", tasks, total)
	}
}

func TestSubmitTaskRequiresAcceptanceAndAuthorization(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)

	task := createOfficialTaskForTest(t, db, 1, "L1")
	analyst := createOfficialTaskTestAnalyst(t, db, "提交测试分析师", "L1")

	if _, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		Summary: "版本摘要",
	}, time.Now()); !errors.Is(err, ErrOfficialSubmissionInvalid) {
		t.Fatalf("submit without authorization err = %v, want %v", err, ErrOfficialSubmissionInvalid)
	}

	if _, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "版本摘要",
	}, time.Now()); !errors.Is(err, ErrOfficialTaskUnavailable) {
		t.Fatalf("submit without acceptance err = %v, want %v", err, ErrOfficialTaskUnavailable)
	}
}

func TestSubmitTaskCreatesSubmissionAndBaseRewardOnce(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 1, "L1")
	if err := db.Model(&models.OfficialAnalysisTask{}).Where("id = ?", task.ID).Update("base_reward_amount", 100).Error; err != nil {
		t.Fatalf("update base reward: %v", err)
	}
	analyst := createOfficialTaskTestAnalyst(t, db, "基础奖励分析师", "L1")
	if _, err := service.AcceptTask(analyst.ID, task.ID, now); err != nil {
		t.Fatalf("accept task: %v", err)
	}

	submission, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "官方任务提交版本",
	}, now)
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}
	if submission.Status != models.OfficialAnalysisSubmissionSubmitted {
		t.Fatalf("submission status = %s, want submitted", submission.Status)
	}

	var acceptance models.OfficialAnalysisTaskAcceptance
	if err := db.Where("task_id = ? AND analyst_id = ?", task.ID, analyst.ID).First(&acceptance).Error; err != nil {
		t.Fatalf("find acceptance: %v", err)
	}
	if acceptance.Status != models.OfficialAnalysisAcceptanceSubmitted || acceptance.SubmittedAt == nil {
		t.Fatalf("acceptance status/submitted_at = %s/%v, want submitted/non-nil", acceptance.Status, acceptance.SubmittedAt)
	}

	var rewards []models.AnalystRewardRecord
	if err := db.Where("analyst_id = ?", analyst.ID).Find(&rewards).Error; err != nil {
		t.Fatalf("find rewards: %v", err)
	}
	if len(rewards) != 1 || rewards[0].RewardType != "base" || rewards[0].Amount != 100 {
		t.Fatalf("rewards = %#v, want one base reward 100", rewards)
	}

	if _, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "重复提交版本",
	}, now); !errors.Is(err, ErrOfficialSubmissionInvalid) {
		t.Fatalf("duplicate submit err = %v, want %v", err, ErrOfficialSubmissionInvalid)
	}
}

func TestRevisionRequiredSubmissionCanBeResubmittedAndListed(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 2, "L1")
	analyst := createOfficialTaskTestAnalyst(t, db, "重提分析师", "L1")
	if _, err := service.AcceptTask(analyst.ID, task.ID, now); err != nil {
		t.Fatalf("accept task: %v", err)
	}
	first, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "第一版",
	}, now)
	if err != nil {
		t.Fatalf("submit first version: %v", err)
	}
	if _, err := service.ReviewSubmission(first.ID, 8, OfficialSubmissionReviewRequest{
		Status:     models.OfficialAnalysisSubmissionRevisionRequired,
		ReviewNote: "补充关键片段",
	}, now); err != nil {
		t.Fatalf("review revision required: %v", err)
	}
	second, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "第二版",
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("resubmit after revision required: %v", err)
	}
	if second.ID == first.ID {
		t.Fatalf("second submission id = first id %d, want a new version", second.ID)
	}

	acceptances, total, err := service.ListMyAcceptances(analyst.ID, 1, 20)
	if err != nil {
		t.Fatalf("list my acceptances: %v", err)
	}
	if total != 1 || len(acceptances) != 1 {
		t.Fatalf("acceptances = %#v total=%d, want one", acceptances, total)
	}
	if len(acceptances[0].Submissions) != 2 || acceptances[0].Submissions[0].ID != second.ID {
		t.Fatalf("submissions = %#v, want latest second version first", acceptances[0].Submissions)
	}
}

func TestReviewAndAdoptSubmissionCreatesRewardAndUpdatesAnalystStats(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 1, "L1")
	analyst := createOfficialTaskTestAnalyst(t, db, "采用测试分析师", "L1")
	if _, err := service.AcceptTask(analyst.ID, task.ID, now); err != nil {
		t.Fatalf("accept task: %v", err)
	}
	submission, err := service.SubmitTask(analyst.ID, task.ID, OfficialTaskSubmitRequest{
		VideoAuthorizationStatus: "authorized",
		Summary:                  "可采用版本",
	}, now)
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}

	reviewed, err := service.ReviewSubmission(submission.ID, 8, OfficialSubmissionReviewRequest{
		Status:     models.OfficialAnalysisSubmissionApproved,
		ReviewNote: "通过",
	}, now)
	if err != nil {
		t.Fatalf("review submission: %v", err)
	}
	if reviewed.Status != models.OfficialAnalysisSubmissionApproved {
		t.Fatalf("reviewed status = %s, want approved", reviewed.Status)
	}

	adoption, err := service.AdoptSubmission(submission.ID, 8, OfficialSubmissionAdoptionRequest{
		AdoptionStatus: models.OfficialContentAdoptionOfficialPublished,
		Channel:        "douyin",
		WorkTitle:      "2034杯精选分析",
		WorkSummary:    "官方采用版本",
		RewardAmount:   300,
		IsPublic:       true,
	}, now)
	if err != nil {
		t.Fatalf("adopt submission: %v", err)
	}
	if adoption.RewardAmount != 300 || !adoption.IsPublic {
		t.Fatalf("adoption = %#v, want reward 300 and public", adoption)
	}

	var updatedAnalyst models.Analyst
	if err := db.First(&updatedAnalyst, analyst.ID).Error; err != nil {
		t.Fatalf("find analyst: %v", err)
	}
	if updatedAnalyst.OfficialAdoptionCount != 1 || updatedAnalyst.OfficialPublishCount != 1 {
		t.Fatalf("analyst stats adoption/publish = %d/%d, want 1/1", updatedAnalyst.OfficialAdoptionCount, updatedAnalyst.OfficialPublishCount)
	}

	var reward models.AnalystRewardRecord
	if err := db.Where("source_type = ? AND source_id = ? AND reward_type = ?", "official_adoption", adoption.ID, "adoption").First(&reward).Error; err != nil {
		t.Fatalf("find adoption reward: %v", err)
	}
	if reward.Amount != 300 || reward.Status != models.AnalystRewardPending {
		t.Fatalf("reward amount/status = %.2f/%s, want 300/pending", reward.Amount, reward.Status)
	}

	rewards, total, err := service.ListAnalystRewards(analyst.ID, 1, 20)
	if err != nil {
		t.Fatalf("list analyst rewards: %v", err)
	}
	if total != 1 || len(rewards) != 1 || rewards[0].ID != reward.ID {
		t.Fatalf("analyst rewards = %#v total=%d, want adoption reward", rewards, total)
	}

	materials, materialTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		MatchName:      "2034",
		AgeGroup:       "U12",
		AdoptionStatus: string(models.OfficialContentAdoptionOfficialPublished),
		AnalystID:      analyst.ID,
	})
	if err != nil {
		t.Fatalf("list official materials: %v", err)
	}
	if materialTotal != 1 || len(materials) != 1 || materials[0].ID != adoption.ID || materials[0].Task == nil || materials[0].Submission == nil {
		t.Fatalf("materials = %#v total=%d, want adoption with task and submission", materials, materialTotal)
	}

	publishRecord, err := service.CreatePublishRecord(adoption.ID, 8, OfficialPublishRecordRequest{
		Channel:       "douyin",
		AccountName:   "少年球探官方抖音号",
		PublishURL:    "https://example.com/published",
		PublishTitle:  "2034杯精选分析发布版",
		ReusePurpose:  "官方短视频发布",
		PlayCount:     12000,
		LikeCount:     860,
		CommentCount:  45,
		ShareCount:    21,
		FavoriteCount: 98,
		MetricsAt:     &now,
		Note:          "首发记录",
	}, now)
	if err != nil {
		t.Fatalf("create publish record: %v", err)
	}
	if publishRecord.Channel != "douyin" || publishRecord.CreatedBy != 8 || publishRecord.PublishedAt == nil || publishRecord.PlayCount != 12000 {
		t.Fatalf("publish record = %#v, want douyin record with published time", publishRecord)
	}
	updatedPublishRecord, err := service.UpdatePublishRecord(adoption.ID, publishRecord.ID, OfficialPublishRecordRequest{
		Channel:       "douyin",
		AccountName:   "少年球探官方抖音号",
		PublishURL:    "https://example.com/published-updated",
		PublishTitle:  "2034杯精选分析发布版更新",
		ReusePurpose:  "官方短视频复盘",
		PublishedAt:   publishRecord.PublishedAt,
		PlayCount:     24000,
		LikeCount:     1600,
		CommentCount:  88,
		ShareCount:    45,
		FavoriteCount: 166,
		MetricsAt:     &now,
		Note:          "复盘更新",
	}, now)
	if err != nil {
		t.Fatalf("update publish record: %v", err)
	}
	if updatedPublishRecord.PublishTitle != "2034杯精选分析发布版更新" || updatedPublishRecord.PlayCount != 24000 || updatedPublishRecord.LikeCount != 1600 {
		t.Fatalf("updated publish record = %#v, want updated metrics", updatedPublishRecord)
	}
	tempPublishRecord, err := service.CreatePublishRecord(adoption.ID, 8, OfficialPublishRecordRequest{
		Channel:      "other",
		PublishTitle: "临时发布记录",
	}, now)
	if err != nil {
		t.Fatalf("create temp publish record: %v", err)
	}
	if err := service.DeletePublishRecord(adoption.ID, tempPublishRecord.ID); err != nil {
		t.Fatalf("delete temp publish record: %v", err)
	}
	if err := service.DeletePublishRecord(adoption.ID, tempPublishRecord.ID); !errors.Is(err, ErrOfficialPublishRecordNotFound) {
		t.Fatalf("delete temp publish record again err = %v, want not found", err)
	}
	publishedMaterials, publishedTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:     analyst.ID,
		PublishStatus: "published",
	})
	if err != nil {
		t.Fatalf("list published materials: %v", err)
	}
	if publishedTotal != 1 || len(publishedMaterials) != 1 || publishedMaterials[0].ID != adoption.ID || len(publishedMaterials[0].PublishRecords) != 1 {
		t.Fatalf("published materials = %#v total=%d, want adoption with publish record", publishedMaterials, publishedTotal)
	}
	metricsMaterials, metricsTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:     analyst.ID,
		MetricsStatus: "has_metrics",
		SortBy:        "total_play_desc",
	})
	if err != nil {
		t.Fatalf("list metrics materials: %v", err)
	}
	if metricsTotal != 1 || len(metricsMaterials) != 1 || metricsMaterials[0].ID != adoption.ID {
		t.Fatalf("metrics materials = %#v total=%d, want adoption with metrics", metricsMaterials, metricsTotal)
	}
	missingMetricsMaterials, missingMetricsTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:     analyst.ID,
		MetricsStatus: "missing_metrics",
		PublishStatus: "published",
	})
	if err != nil {
		t.Fatalf("list missing metrics materials: %v", err)
	}
	if missingMetricsTotal != 0 || len(missingMetricsMaterials) != 0 {
		t.Fatalf("missing metrics materials = %#v total=%d, want none", missingMetricsMaterials, missingMetricsTotal)
	}
	withoutBonusMaterials, withoutBonusTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:   analyst.ID,
		BonusStatus: "without_playback_bonus",
	})
	if err != nil {
		t.Fatalf("list without bonus materials: %v", err)
	}
	if withoutBonusTotal != 1 || len(withoutBonusMaterials) != 1 || withoutBonusMaterials[0].ID != adoption.ID {
		t.Fatalf("without bonus materials = %#v total=%d, want adoption before playback bonus", withoutBonusMaterials, withoutBonusTotal)
	}
	unpublishedMaterials, unpublishedTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:     analyst.ID,
		PublishStatus: "unpublished",
	})
	if err != nil {
		t.Fatalf("list unpublished materials: %v", err)
	}
	if unpublishedTotal != 0 || len(unpublishedMaterials) != 0 {
		t.Fatalf("unpublished materials = %#v total=%d, want none after publish record", unpublishedMaterials, unpublishedTotal)
	}

	hidden, err := service.UpdateAdoptionPublic(adoption.ID, OfficialAdoptionPublicRequest{IsPublic: false}, now)
	if err != nil {
		t.Fatalf("update adoption public: %v", err)
	}
	if hidden.IsPublic {
		t.Fatalf("hidden adoption is public, want false")
	}

	playbackBonus, err := service.CreatePlaybackBonus(adoption.ID, 8, OfficialPlaybackBonusRequest{
		Amount:     88,
		BonusBasis: "抖音播放表现优秀",
		Note:       "播放表现奖金",
	}, now)
	if err != nil {
		t.Fatalf("create playback bonus: %v", err)
	}
	if playbackBonus.RewardType != "playback_bonus" || playbackBonus.Amount != 88 || playbackBonus.Status != models.AnalystRewardPending {
		t.Fatalf("playback bonus = %#v, want pending playback bonus 88", playbackBonus)
	}
	withBonusMaterials, withBonusTotal, err := service.ListOfficialMaterials(1, 20, OfficialMaterialListFilters{
		AnalystID:   analyst.ID,
		BonusStatus: "with_playback_bonus",
	})
	if err != nil {
		t.Fatalf("list with bonus materials: %v", err)
	}
	if withBonusTotal != 1 || len(withBonusMaterials) != 1 || withBonusMaterials[0].ID != adoption.ID {
		t.Fatalf("with bonus materials = %#v total=%d, want adoption after playback bonus", withBonusMaterials, withBonusTotal)
	}

	if _, err := service.UpdateAdoptionPublic(adoption.ID, OfficialAdoptionPublicRequest{IsPublic: true}, now); err != nil {
		t.Fatalf("restore adoption public: %v", err)
	}
	topics, topicTotal, err := service.ListPublicEventTopics(1, 20, OfficialEventTopicFilters{Keyword: "2034"})
	if err != nil {
		t.Fatalf("list public event topics: %v", err)
	}
	if topicTotal != 1 || len(topics) != 1 || topics[0].MatchName != "2034杯" || topics[0].WorkCount != 1 || topics[0].FeaturedWork == nil {
		t.Fatalf("topics = %#v total=%d, want one public 2034 topic", topics, topicTotal)
	}
	detail, detailTotal, err := service.GetPublicEventTopic("2034杯", 1, 20, "")
	if err != nil {
		t.Fatalf("get public event topic: %v", err)
	}
	if detailTotal != 1 || detail.MatchName != "2034杯" || len(detail.Works) != 1 || detail.Works[0].WorkTitle != "2034杯精选分析" {
		t.Fatalf("topic detail = %#v total=%d, want public work", detail, detailTotal)
	}

	settled, err := service.SettleReward(reward.ID, 8, OfficialRewardActionRequest{Note: "本月结算"}, now)
	if err != nil {
		t.Fatalf("settle reward: %v", err)
	}
	if settled.Status != models.AnalystRewardSettled || settled.SettledAt == nil || settled.SettledBy != 8 {
		t.Fatalf("settled reward = %#v, want settled with admin", settled)
	}
	if settled.SettlementBatchID == 0 {
		t.Fatalf("settled reward settlement batch id = 0, want batch id")
	}
	batch, err := service.GetRewardSettlementBatch(settled.SettlementBatchID)
	if err != nil {
		t.Fatalf("get reward settlement batch: %v", err)
	}
	if batch.BatchNo == "" || batch.RewardCount != 1 || batch.TotalAmount != reward.Amount || len(batch.Rewards) != 1 {
		t.Fatalf("settlement batch = %#v, want one reward batch", batch)
	}
	batchedRewards, batchTotal, err := service.ListAdminRewards(1, 20, OfficialRewardListFilters{BatchID: batch.ID})
	if err != nil {
		t.Fatalf("list rewards by batch: %v", err)
	}
	if batchTotal != 1 || len(batchedRewards) != 1 || batchedRewards[0].ID != reward.ID {
		t.Fatalf("batched rewards = %#v total=%d, want settled reward", batchedRewards, batchTotal)
	}
	if _, err := service.SettleReward(reward.ID, 8, OfficialRewardActionRequest{}, now); !errors.Is(err, ErrAnalystRewardInvalid) {
		t.Fatalf("settle already settled err = %v, want %v", err, ErrAnalystRewardInvalid)
	}
	reversed, err := service.ReverseReward(reward.ID, 8, OfficialRewardActionRequest{Note: "采用取消"}, now)
	if err != nil {
		t.Fatalf("reverse reward: %v", err)
	}
	if reversed.Status != models.AnalystRewardReversed {
		t.Fatalf("reversed status = %s, want reversed", reversed.Status)
	}
}

func TestAcceptTaskRespectsMaxAcceptCount(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 2, "L2")
	first := createOfficialTaskTestAnalyst(t, db, "分析师一", "L3")
	second := createOfficialTaskTestAnalyst(t, db, "分析师二", "L3")
	third := createOfficialTaskTestAnalyst(t, db, "分析师三", "L3")

	if _, err := service.AcceptTask(first.ID, task.ID, now); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	if _, err := service.AcceptTask(second.ID, task.ID, now); err != nil {
		t.Fatalf("second accept: %v", err)
	}
	if _, err := service.AcceptTask(third.ID, task.ID, now); !errors.Is(err, ErrOfficialTaskFull) {
		t.Fatalf("third accept err = %v, want %v", err, ErrOfficialTaskFull)
	}

	var acceptedCount int64
	if err := db.Model(&models.OfficialAnalysisTaskAcceptance{}).Where("task_id = ?", task.ID).Count(&acceptedCount).Error; err != nil {
		t.Fatalf("count acceptances: %v", err)
	}
	if acceptedCount != 2 {
		t.Fatalf("accepted count = %d, want 2", acceptedCount)
	}

	var updated models.OfficialAnalysisTask
	if err := db.First(&updated, task.ID).Error; err != nil {
		t.Fatalf("find updated task: %v", err)
	}
	if updated.CurrentAcceptCount != 2 {
		t.Fatalf("current_accept_count = %d, want 2", updated.CurrentAcceptCount)
	}
	if updated.Status != models.OfficialAnalysisTaskFull {
		t.Fatalf("task status = %s, want %s", updated.Status, models.OfficialAnalysisTaskFull)
	}
}

func TestPublicEventTopicConfigMergesAliasesAndPinsWork(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	analyst := createOfficialTaskTestAnalyst(t, db, "专题分析师", "L3")

	taskPrimary := createOfficialTaskForTest(t, db, 2, "L1")
	taskAlias := createOfficialTaskForTest(t, db, 2, "L1")
	if err := db.Model(&taskAlias).Updates(map[string]interface{}{
		"task_no":    "OFF-ALIAS-001",
		"title":      "2034杯 小组赛精选",
		"match_name": "2034杯小组赛",
		"age_group":  "U13",
		"updated_at": time.Now(),
	}).Error; err != nil {
		t.Fatalf("update alias task: %v", err)
	}

	primary := models.OfficialContentAdoption{
		TaskID:         taskPrimary.ID,
		SubmissionID:   11,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionOfficialPublished,
		Channel:        "douyin",
		WorkTitle:      "2034杯官方精选",
		WorkSummary:    "主赛事作品",
		IsPublic:       true,
	}
	if err := db.Create(&primary).Error; err != nil {
		t.Fatalf("create primary adoption: %v", err)
	}

	alias := models.OfficialContentAdoption{
		TaskID:         taskAlias.ID,
		SubmissionID:   12,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionKeySpread,
		Channel:        "douyin",
		WorkTitle:      "2034杯小组赛爆款",
		WorkSummary:    "别名赛事作品",
		IsPublic:       true,
	}
	if err := db.Create(&alias).Error; err != nil {
		t.Fatalf("create alias adoption: %v", err)
	}

	if _, err := service.SaveEventTopicConfig(OfficialEventTopicConfigRequest{
		MatchName:        "2034杯",
		DisplayName:      "2034杯官方专题",
		Summary:          "官方采用作品合集",
		CoverURL:         "https://example.com/2034-cover.jpg",
		AliasNames:       []string{"2034杯小组赛"},
		PinnedAdoptionID: alias.ID,
		IsFeatured:       true,
		SortOrder:        1,
	}, 99, time.Now()); err != nil {
		t.Fatalf("save topic config: %v", err)
	}

	topics, total, err := service.ListPublicEventTopics(1, 20, OfficialEventTopicFilters{Keyword: "2034"})
	if err != nil {
		t.Fatalf("list public event topics with config: %v", err)
	}
	if total != 1 || len(topics) != 1 {
		t.Fatalf("topics total=%d len=%d, want one merged topic", total, len(topics))
	}
	topic := topics[0]
	if topic.MatchName != "2034杯" || topic.DisplayName != "2034杯官方专题" || topic.WorkCount != 2 || !topic.IsFeatured || topic.SortOrder != 1 {
		t.Fatalf("topic = %#v, want merged canonical topic", topic)
	}
	if topic.FeaturedWork == nil || topic.FeaturedWork.WorkTitle != "2034杯小组赛爆款" {
		t.Fatalf("featured work = %#v, want pinned alias work", topic.FeaturedWork)
	}

	detail, detailTotal, err := service.GetPublicEventTopic("2034杯小组赛", 1, 20, "")
	if err != nil {
		t.Fatalf("get public event topic by alias: %v", err)
	}
	if detailTotal != 2 || detail.DisplayName != "2034杯官方专题" || detail.FeaturedWork == nil || detail.FeaturedWork.WorkTitle != "2034杯小组赛爆款" {
		t.Fatalf("detail = %#v total=%d, want merged configured detail", detail, detailTotal)
	}
	filtered, filteredTotal, err := service.GetPublicEventTopic("2034杯", 1, 20, "U13")
	if err != nil {
		t.Fatalf("get public event topic by age group: %v", err)
	}
	if filteredTotal != 1 || filtered.AgeGroupFilter != "U13" || len(filtered.Works) != 1 || filtered.Works[0].WorkTitle != "2034杯小组赛爆款" {
		t.Fatalf("filtered detail = %#v total=%d, want only U13 alias work", filtered, filteredTotal)
	}
	featured, featuredTotal, err := service.ListPublicEventTopics(1, 20, OfficialEventTopicFilters{FeaturedOnly: true})
	if err != nil {
		t.Fatalf("list featured topics: %v", err)
	}
	if featuredTotal != 1 || len(featured) != 1 || featured[0].MatchName != "2034杯" {
		t.Fatalf("featured topics = %#v total=%d, want 2034 topic", featured, featuredTotal)
	}
	if _, err := service.SaveEventTopicConfig(OfficialEventTopicConfigRequest{
		MatchName:  "2034杯冲突专题",
		AliasNames: []string{"2034杯小组赛"},
	}, 99, time.Now()); !errors.Is(err, ErrOfficialTaskInvalid) {
		t.Fatalf("save duplicate alias err = %v, want %v", err, ErrOfficialTaskInvalid)
	}
}

func TestAcceptTaskRejectsDuplicateAnalyst(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 2, "L1")
	analyst := createOfficialTaskTestAnalyst(t, db, "重复接单分析师", "L2")

	if _, err := service.AcceptTask(analyst.ID, task.ID, now); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	if _, err := service.AcceptTask(analyst.ID, task.ID, now); !errors.Is(err, ErrOfficialTaskDuplicate) {
		t.Fatalf("duplicate accept err = %v, want %v", err, ErrOfficialTaskDuplicate)
	}

	var updated models.OfficialAnalysisTask
	if err := db.First(&updated, task.ID).Error; err != nil {
		t.Fatalf("find updated task: %v", err)
	}
	if updated.CurrentAcceptCount != 1 {
		t.Fatalf("current_accept_count after duplicate = %d, want 1", updated.CurrentAcceptCount)
	}
}

func TestAcceptTaskRequiresVisibleLevel(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)

	task := createOfficialTaskForTest(t, db, 1, "L3")
	analyst := createOfficialTaskTestAnalyst(t, db, "等级不足分析师", "L1")

	if _, err := service.AcceptTask(analyst.ID, task.ID, time.Now()); !errors.Is(err, ErrOfficialTaskLevelDenied) {
		t.Fatalf("accept err = %v, want %v", err, ErrOfficialTaskLevelDenied)
	}
}

func TestAcceptTaskUsesExplicitVisibleLevelCodes(t *testing.T) {
	db := setupOfficialAnalysisTaskTestDB(t)
	service := NewOfficialAnalysisTaskService(db)
	now := time.Now()

	task := createOfficialTaskForTest(t, db, 2, "L1")
	if err := db.Model(task).Update("visible_level_codes", "L1,L3").Error; err != nil {
		t.Fatalf("update visible level codes: %v", err)
	}
	l2 := createOfficialTaskTestAnalyst(t, db, "多选等级不足分析师", "L2")
	l3 := createOfficialTaskTestAnalyst(t, db, "多选等级命中分析师", "L3")

	if _, err := service.AcceptTask(l2.ID, task.ID, now); !errors.Is(err, ErrOfficialTaskLevelDenied) {
		t.Fatalf("L2 accept err = %v, want %v", err, ErrOfficialTaskLevelDenied)
	}
	if _, err := service.AcceptTask(l3.ID, task.ID, now); err != nil {
		t.Fatalf("L3 accept err = %v", err)
	}
}
