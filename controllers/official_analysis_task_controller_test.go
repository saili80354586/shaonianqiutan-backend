package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOfficialAnalysisTaskControllerTest(t *testing.T) (*gorm.DB, *OfficialAnalysisTaskController, models.User, models.Analyst) {
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
		&models.Order{},
		&models.OrderAssignment{},
		&models.OfficialAnalysisTask{},
		&models.OfficialAnalysisTaskAcceptance{},
		&models.OfficialAnalysisSubmission{},
		&models.OfficialContentAdoption{},
		&models.OfficialContentPublishRecord{},
		&models.OfficialEventTopicConfig{},
		&models.AnalystRewardSettlementBatch{},
		&models.AnalystRewardRecord{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := models.User{
		Phone:    "13918880001",
		Password: "test",
		Name:     "接口测试分析师",
		Role:     models.RoleAnalyst,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analyst := models.Analyst{
		UserID:    user.ID,
		Name:      "接口测试分析师",
		LevelCode: "L2",
		Status:    models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	return db, NewOfficialAnalysisTaskController(services.NewOfficialAnalysisTaskService(db)), user, analyst
}

func TestOfficialAnalysisTaskAPIAdminCreatePublishAndAnalystAccept(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, ctrl, analystUser, _ := setupOfficialAnalysisTaskControllerTest(t)
	player := models.User{
		Phone:        "13918880002",
		Password:     "test",
		Name:         "测试球员",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		CurrentTeam:  "河南少年队",
		JerseyColor:  "红色",
		JerseyNumber: 10,
		Position:     "前锋",
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	router := gin.New()
	router.POST("/admin/official-analysis-tasks", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.CreateAdminTask(c)
	})
	router.POST("/admin/official-analysis-tasks/batch", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.BatchCreateAdminTasks(c)
	})
	router.GET("/admin/official-analysis-tasks/:id", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.GetAdminTask(c)
	})
	router.GET("/admin/official-analysis-tasks/:id/acceptances", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListTaskAcceptances(c)
	})
	router.GET("/admin/official-analysis-tasks/:id/submissions", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListTaskSubmissions(c)
	})
	router.PUT("/admin/official-analysis-tasks/:id", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.UpdateAdminTask(c)
	})
	router.POST("/admin/official-analysis-tasks/:id/publish", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.PublishAdminTask(c)
	})
	router.GET("/analyst/official-analysis-tasks", func(c *gin.Context) {
		c.Set("userId", analystUser.ID)
		ctrl.ListAvailableTasks(c)
	})
	router.GET("/analyst/official-analysis-tasks/:id", func(c *gin.Context) {
		c.Set("userId", analystUser.ID)
		ctrl.GetAvailableTask(c)
	})
	router.GET("/analyst/official-analysis-tasks/mine", func(c *gin.Context) {
		c.Set("userId", analystUser.ID)
		ctrl.ListMyTasks(c)
	})
	router.POST("/analyst/official-analysis-tasks/:id/accept", func(c *gin.Context) {
		c.Set("userId", analystUser.ID)
		ctrl.AcceptTask(c)
	})
	router.POST("/analyst/official-analysis-tasks/:id/submit", func(c *gin.Context) {
		c.Set("userId", analystUser.ID)
		ctrl.SubmitTask(c)
	})
	router.GET("/admin/official-analysis-submissions", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListAdminSubmissions(c)
	})
	router.GET("/admin/official-analysis-submissions/:id", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.GetSubmission(c)
	})
	router.POST("/admin/official-analysis-submissions/:id/review", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ReviewSubmission(c)
	})
	router.POST("/admin/official-analysis-submissions/:id/adopt", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.AdoptSubmission(c)
	})
	router.GET("/admin/official-content-adoptions", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListOfficialMaterials(c)
	})
	router.POST("/admin/official-content-adoptions/:id/publish-records", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.CreatePublishRecord(c)
	})
	router.PUT("/admin/official-content-adoptions/:id/publish-records/:recordId", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.UpdatePublishRecord(c)
	})
	router.DELETE("/admin/official-content-adoptions/:id/publish-records/:recordId", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.DeletePublishRecord(c)
	})

	createBody := mustJSON(t, gin.H{
		"title":              "2034杯 U12 官方选题",
		"max_accept_count":   2,
		"visible_level_min":  "L2",
		"base_reward_amount": 0,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/official-analysis-tasks", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}

	var task models.OfficialAnalysisTask
	if err := db.First(&task).Error; err != nil {
		t.Fatalf("find created task: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/publish", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("publish without authorization status = %d, want 400", rec.Code)
	}

	updateBody := mustJSON(t, gin.H{
		"title":                  "2034杯 U12 官方选题",
		"max_accept_count":       2,
		"visible_level_min":      "L2",
		"base_reward_amount":     0,
		"authorization_status":   "authorized",
		"target_player_user_id":  player.ID,
		"video_first_half_url":   "https://example.com/videos/2034-u12-first.mp4",
		"target_player_name":     "测试球员",
		"target_player_team":     "河南少年队",
		"target_jersey_color":    "红色",
		"target_jersey_number":   "10",
		"target_player_position": "前锋",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/publish", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "2034杯 U12 官方选题") {
		t.Fatalf("admin task detail status = %d body=%s, want task detail", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/analyst/official-analysis-tasks", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/analyst/official-analysis-tasks/"+officialTaskIDString(task.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "2034杯 U12 官方选题") {
		t.Fatalf("analyst task detail status = %d body=%s, want visible task", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/analyst/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/accept", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("accept status = %d body=%s", rec.Code, rec.Body.String())
	}

	var acceptedCount int64
	if err := db.Model(&models.OfficialAnalysisTaskAcceptance{}).Where("task_id = ?", task.ID).Count(&acceptedCount).Error; err != nil {
		t.Fatalf("count acceptance: %v", err)
	}
	if acceptedCount != 1 {
		t.Fatalf("accepted count = %d, want 1", acceptedCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/acceptances", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "接口测试分析师") {
		t.Fatalf("task acceptances status = %d body=%s, want analyst acceptance", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/analyst/official-analysis-tasks/mine", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "2034杯 U12 官方选题") {
		t.Fatalf("my official tasks alias status = %d body=%s, want accepted task", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	submitMissingAuth := mustJSON(t, gin.H{
		"summary": "缺授权状态版本",
	})
	req = httptest.NewRequest(http.MethodPost, "/analyst/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/submit", bytes.NewReader(submitMissingAuth))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("submit missing auth status = %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	submitBody := mustJSON(t, gin.H{
		"video_authorization_status": "authorized",
		"summary":                    "官方任务提交摘要",
		"script_text":                "视频分析脚本",
	})
	req = httptest.NewRequest(http.MethodPost, "/analyst/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/submit", bytes.NewReader(submitBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit status = %d body=%s", rec.Code, rec.Body.String())
	}

	var submission models.OfficialAnalysisSubmission
	if err := db.Where("task_id = ?", task.ID).First(&submission).Error; err != nil {
		t.Fatalf("find submission: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-analysis-submissions?task_id="+officialTaskIDString(task.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list submissions status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-analysis-tasks/"+officialTaskIDString(task.ID)+"/submissions", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "官方任务提交摘要") {
		t.Fatalf("task submissions status = %d body=%s, want task submission", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-analysis-submissions/"+officialTaskIDString(submission.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "官方任务提交摘要") {
		t.Fatalf("submission detail status = %d body=%s, want submission detail", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	reviewBody := mustJSON(t, gin.H{
		"status":      "approved",
		"review_note": "通过",
	})
	req = httptest.NewRequest(http.MethodPost, "/admin/official-analysis-submissions/"+officialTaskIDString(submission.ID)+"/review", bytes.NewReader(reviewBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("review status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	adoptBody := mustJSON(t, gin.H{
		"adoption_status": "official_published",
		"channel":         "douyin",
		"work_title":      "2034杯 U12 精选分析",
		"work_summary":    "官方采用作品",
		"reward_amount":   200,
		"is_public":       true,
	})
	req = httptest.NewRequest(http.MethodPost, "/admin/official-analysis-submissions/"+officialTaskIDString(submission.ID)+"/adopt", bytes.NewReader(adoptBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("adopt status = %d body=%s", rec.Code, rec.Body.String())
	}
	var adoption models.OfficialContentAdoption
	if err := db.Where("work_title = ?", "2034杯 U12 精选分析").First(&adoption).Error; err != nil {
		t.Fatalf("find adoption: %v", err)
	}

	publishBody := mustJSON(t, gin.H{
		"channel":        "douyin",
		"account_name":   "少年球探官方抖音号",
		"publish_url":    "https://example.com/controller-publish",
		"publish_title":  "2034杯 U12 精选分析发布版",
		"reuse_purpose":  "官方短视频发布",
		"play_count":     12345,
		"like_count":     678,
		"comment_count":  45,
		"share_count":    12,
		"favorite_count": 34,
		"note":           "接口发布记录",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/official-content-adoptions/"+officialTaskIDString(adoption.ID)+"/publish-records", bytes.NewReader(publishBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish record status = %d body=%s", rec.Code, rec.Body.String())
	}
	var publishRecord models.OfficialContentPublishRecord
	if err := db.Where("publish_url = ?", "https://example.com/controller-publish").First(&publishRecord).Error; err != nil {
		t.Fatalf("find publish record: %v", err)
	}
	if publishRecord.PlayCount != 12345 || publishRecord.LikeCount != 678 {
		t.Fatalf("publish record metrics = %#v, want interaction data", publishRecord)
	}

	updatePublishBody := mustJSON(t, gin.H{
		"channel":        "douyin",
		"account_name":   "少年球探官方抖音号",
		"publish_url":    "https://example.com/controller-publish-updated",
		"publish_title":  "2034杯 U12 精选分析发布版更新",
		"reuse_purpose":  "官方短视频复盘",
		"play_count":     23456,
		"like_count":     1200,
		"comment_count":  88,
		"share_count":    35,
		"favorite_count": 90,
		"note":           "接口发布记录更新",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/admin/official-content-adoptions/"+officialTaskIDString(adoption.ID)+"/publish-records/"+officialTaskIDString(publishRecord.ID), bytes.NewReader(updatePublishBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "23456") {
		t.Fatalf("publish record update status = %d body=%s", rec.Code, rec.Body.String())
	}

	tempPublishBody := mustJSON(t, gin.H{
		"channel":       "other",
		"publish_title": "临时删除发布记录",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/official-content-adoptions/"+officialTaskIDString(adoption.ID)+"/publish-records", bytes.NewReader(tempPublishBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("temp publish record status = %d body=%s", rec.Code, rec.Body.String())
	}
	var tempPublishRecord models.OfficialContentPublishRecord
	if err := db.Where("publish_title = ?", "临时删除发布记录").First(&tempPublishRecord).Error; err != nil {
		t.Fatalf("find temp publish record: %v", err)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/admin/official-content-adoptions/"+officialTaskIDString(adoption.ID)+"/publish-records/"+officialTaskIDString(tempPublishRecord.ID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("temp publish record delete status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-content-adoptions?publish_status=published", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "controller-publish-updated") || strings.Contains(rec.Body.String(), "临时删除发布记录") {
		t.Fatalf("materials published status = %d body=%s, want publish record", rec.Code, rec.Body.String())
	}

	var reward models.AnalystRewardRecord
	if err := db.Where("reward_type = ?", "adoption").First(&reward).Error; err != nil {
		t.Fatalf("find adoption reward: %v", err)
	}
	if reward.Amount != 200 {
		t.Fatalf("reward amount = %.2f, want 200", reward.Amount)
	}
}

func TestOfficialAnalysisTaskAPIBatchCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, ctrl, _, _ := setupOfficialAnalysisTaskControllerTest(t)
	firstPlayer := models.User{Phone: "13918880003", Password: "test", Name: "批量球员A", Role: models.RoleUser, Status: models.StatusActive}
	secondPlayer := models.User{Phone: "13918880004", Password: "test", Name: "批量球员B", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&firstPlayer).Error; err != nil {
		t.Fatalf("create first player: %v", err)
	}
	if err := db.Create(&secondPlayer).Error; err != nil {
		t.Fatalf("create second player: %v", err)
	}
	router := gin.New()
	router.POST("/admin/official-analysis-tasks/batch", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.BatchCreateAdminTasks(c)
	})

	body := mustJSON(t, gin.H{
		"event_name":           "2034杯河南赛区八强",
		"publish_after_create": true,
		"common": gin.H{
			"age_group":            "U12",
			"authorization_status": "authorized",
			"video_source":         "官方赛事素材",
			"base_reward_amount":   0,
			"adoption_reward_rule": "按官方采用情况结算奖金",
			"requirements":         "重点分析目标球员的攻防表现",
			"max_accept_count":     2,
			"visible_level_min":    "L2",
			"priority_level_min":   "L3",
		},
		"tasks": []gin.H{
			{
				"title":                  "2034杯河南赛区八强 A队vsB队 10号",
				"match_name":             "2034杯河南赛区八强",
				"video_first_half_url":   "https://example.com/a-first.mp4",
				"video_second_half_url":  "https://example.com/a-second.mp4",
				"target_player_user_id":  firstPlayer.ID,
				"target_player_name":     "批量球员A",
				"target_player_team":     "河南少年A队",
				"target_jersey_color":    "红色",
				"target_jersey_number":   "10",
				"target_player_position": "前锋",
			},
			{
				"match_name":             "2034杯河南赛区八强",
				"video_first_half_url":   "https://example.com/b-first.mp4",
				"target_player_user_id":  secondPlayer.ID,
				"target_player_name":     "批量球员B",
				"target_player_team":     "河南少年B队",
				"target_jersey_color":    "蓝色",
				"target_jersey_number":   "8",
				"target_player_position": "中场",
			},
		},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/official-analysis-tasks/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("batch create status = %d body=%s", rec.Code, rec.Body.String())
	}

	var total int64
	if err := db.Model(&models.OfficialAnalysisTask{}).Count(&total).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if total != 2 {
		t.Fatalf("task total = %d, want 2", total)
	}

	var tasks []models.OfficialAnalysisTask
	if err := db.Order("id ASC").Find(&tasks).Error; err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if tasks[0].Status != models.OfficialAnalysisTaskPublished || tasks[1].Status != models.OfficialAnalysisTaskPublished {
		t.Fatalf("batch tasks should be published, got %s/%s", tasks[0].Status, tasks[1].Status)
	}
	if tasks[1].Title != "2034杯河南赛区八强｜批量球员B" {
		t.Fatalf("derived title = %q", tasks[1].Title)
	}
	if tasks[0].VideoSecondHalfURL == "" || tasks[1].TargetPlayerTeam != "河南少年B队" {
		t.Fatalf("batch task fields not persisted: %#v %#v", tasks[0], tasks[1])
	}
	if tasks[0].TaskNo == tasks[1].TaskNo {
		t.Fatalf("batch task numbers should be unique, got %s", tasks[0].TaskNo)
	}
}

func TestOfficialEventTopicConfigAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, ctrl, _, analyst := setupOfficialAnalysisTaskControllerTest(t)

	task := models.OfficialAnalysisTask{
		TaskNo:              "OFF-TOPIC-001",
		Title:               "2034杯专题运营",
		MatchName:           "2034杯",
		AgeGroup:            "U12",
		AuthorizationStatus: "authorized",
		TaskType:            "composite",
		MaxAcceptCount:      2,
		VisibleLevelMin:     "L1",
		Status:              models.OfficialAnalysisTaskPublished,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	adoption := models.OfficialContentAdoption{
		TaskID:         task.ID,
		SubmissionID:   1,
		AnalystID:      analyst.ID,
		AdoptionStatus: models.OfficialContentAdoptionOfficialPublished,
		Channel:        "douyin",
		WorkTitle:      "2034杯官方作品",
		WorkSummary:    "专题运营测试作品",
		IsPublic:       true,
	}
	if err := db.Create(&adoption).Error; err != nil {
		t.Fatalf("create adoption: %v", err)
	}

	router := gin.New()
	router.GET("/admin/official-event-topics", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.ListAdminEventTopics(c)
	})
	router.PUT("/admin/official-event-topics/:matchName", func(c *gin.Context) {
		c.Set("userId", uint(99))
		ctrl.SaveAdminEventTopic(c)
	})
	router.GET("/official-event-topics/:matchName", ctrl.GetPublicEventTopic)

	rec := httptest.NewRecorder()
	saveBody := mustJSON(t, gin.H{
		"display_name":       "2034杯官方专题",
		"summary":            "用于接口测试的专题简介",
		"alias_names":        []string{"2034杯小组赛"},
		"pinned_adoption_id": adoption.ID,
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/official-event-topics/2034%E6%9D%AF", bytes.NewReader(saveBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save topic config status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/official-event-topics", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list admin event topics status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/official-event-topics/2034%E6%9D%AF%E5%B0%8F%E7%BB%84%E8%B5%9B", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get public topic by alias status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func mustJSON(t *testing.T, payload interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}

func officialTaskIDString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
