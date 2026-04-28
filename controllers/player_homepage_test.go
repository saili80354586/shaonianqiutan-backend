package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlayerHomepageTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.Team{},
		&models.TeamPlayer{},
		&models.ClubPlayer{},
		&models.PhysicalTestActivity{},
		&models.PhysicalTestRecord{},
		&models.WeeklyReport{},
		&models.MatchSummary{},
		&models.PlayerReview{},
		&models.Report{},
		&models.Scout{},
		&models.ScoutReport{},
		&models.Post{},
		&models.Follow{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	router := gin.New()
	api := router.Group("/api")
	routes.SetupPlayerPublicRoutes(api, controllers.NewPlayerController(db))

	return router, db
}

func seedPlayerHomepageData(t *testing.T, db *gorm.DB, visible bool) {
	t.Helper()

	now := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	privacy := `{"profileVisible":true,"showRealName":true,"searchable":true}`
	if !visible {
		privacy = `{"profileVisible":false,"showRealName":true,"searchable":false}`
	}
	player := models.User{
		ID:              1001,
		Phone:           "13900001001",
		Password:        "hashed",
		Role:            models.RoleUser,
		CurrentRole:     models.RoleUser,
		Status:          models.StatusActive,
		Name:            "测试球员",
		Nickname:        "子墨",
		Age:             12,
		Position:        "边锋",
		CurrentTeam:     "U12 精英队",
		Province:        "上海",
		City:            "上海",
		Height:          152,
		Weight:          42,
		DominantFoot:    "right",
		School:          "测试小学",
		TechnicalTags:   `["速度快","传球准"]`,
		MentalTags:      `["专注","团队协作"]`,
		PrivacySettings: privacy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	coach := models.User{
		ID:        1002,
		Phone:     "13900001002",
		Password:  "hashed",
		Role:      models.RoleCoach,
		Status:    models.StatusActive,
		Name:      "测试教练",
		Nickname:  "王指导",
		CreatedAt: now,
		UpdatedAt: now,
	}
	scout := models.User{
		ID:        1003,
		Phone:     "13900001003",
		Password:  "hashed",
		Role:      models.RoleScout,
		Status:    models.StatusActive,
		Name:      "测试球探",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := db.Create(&coach).Error; err != nil {
		t.Fatalf("create coach: %v", err)
	}
	if err := db.Create(&scout).Error; err != nil {
		t.Fatalf("create scout: %v", err)
	}
	scoutProfile := models.Scout{ID: 6001, UserID: scout.ID, CurrentOrganization: "测试球探机构", Verified: true, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&scoutProfile).Error; err != nil {
		t.Fatalf("create scout profile: %v", err)
	}

	club := models.Club{ID: 2001, UserID: 2001, Name: "测试青训俱乐部", Province: "上海", City: "上海", CreatedAt: now, UpdatedAt: now}
	team := models.Team{ID: 3001, ClubID: club.ID, Name: "U12 精英队", AgeGroup: "U12", Status: models.TeamStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&club).Error; err != nil {
		t.Fatalf("create club: %v", err)
	}
	if err := db.Create(&team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := db.Create(&models.TeamPlayer{
		TeamID: team.ID, UserID: player.ID, JerseyNumber: "7", Position: "边锋", Status: "active", JoinedAt: now, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team player: %v", err)
	}

	activity := models.PhysicalTestActivity{ID: 4001, ClubID: club.ID, Name: "春季体测", StartDate: now, CreatedBy: coach.ID, Status: models.PTStatusCompleted, CreatedAt: now, UpdatedAt: now}
	sprint := 4.8
	jump := 185.0
	pushUp := 24
	if err := db.Create(&activity).Error; err != nil {
		t.Fatalf("create activity: %v", err)
	}
	if err := db.Create(&models.PhysicalTestRecord{
		ActivityID:       activity.ID,
		PlayerID:         player.ID,
		ClubID:           club.ID,
		TestDate:         now,
		Sprint30m:        &sprint,
		StandingLongJump: &jump,
		PushUp:           &pushUp,
		RecorderID:       coach.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error; err != nil {
		t.Fatalf("create physical record: %v", err)
	}

	if err := db.Create(&models.WeeklyReport{
		TeamID: team.ID, PlayerID: player.ID, CoachID: coach.ID, WeekStart: now, WeekEnd: now.AddDate(0, 0, 6),
		TrainingCount: 3, TrainingDuration: 240, KnowledgeSummary: "边路突破训练", SubmitStatus: "submitted",
		ReviewStatus: "approved", CoachAttitudeRating: 5, CoachTechniqueRating: 4, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create weekly report: %v", err)
	}

	match := models.MatchSummary{ID: 5001, TeamID: team.ID, CoachID: coach.ID, MatchName: "测试杯小组赛", MatchDate: "2026-04-19", Opponent: "晨星 U12", OurScore: 3, OppScore: 1, Result: "win", Status: "completed", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&match).Error; err != nil {
		t.Fatalf("create match: %v", err)
	}
	if err := db.Create(&models.PlayerReview{MatchID: match.ID, TeamID: team.ID, PlayerID: player.ID, Performance: "优秀", Goals: 1, Assists: 1, Highlights: "边路制造威胁", Status: "coach_reviewed", SubmittedAt: now, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create player review: %v", err)
	}

	if err := db.Create(&models.Report{OrderID: 1, UserID: player.ID, AnalystID: coach.ID, PlayerName: player.Name, PlayerPosition: player.Position, Content: "报告正文", Status: models.ReportStatusCompleted, OverallRating: 4.6, Summary: "速度优势明显", Strengths: `["启动快"]`, Weaknesses: `["对抗需提升"]`, Suggestions: "加强核心力量", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
	if err := db.Create(&models.ScoutReport{ScoutID: scoutProfile.ID, PlayerID: player.ID, Status: "published", OverallRating: 88, PotentialRating: "A", Strengths: `["爆发力强"]`, Weaknesses: `["逆足需提升"]`, TechnicalSkills: `{"speed":88}`, Summary: "具备边路突破潜力", Recommendation: "增加对抗训练", PublishedAt: &now, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create scout report: %v", err)
	}
	if err := db.Create(&models.Post{UserID: player.ID, Content: "今天完成边路训练", RoleTag: "player", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	if err := db.Create(&models.Follow{FollowerID: scout.ID, FollowingID: player.ID, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create follow: %v", err)
	}
}

func TestGetHomepageReturnsAggregateData(t *testing.T) {
	router, db := setupPlayerHomepageTestRouter(t)
	seedPlayerHomepageData(t, db, true)

	req := httptest.NewRequest(http.MethodGet, "/api/players/1001/homepage", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Profile struct {
				ID          uint   `json:"id"`
				DisplayName string `json:"displayName"`
				School      string `json:"school"`
			} `json:"profile"`
			Stats struct {
				PhysicalTestCount int `json:"physicalTestCount"`
				WeeklyReportCount int `json:"weeklyReportCount"`
				MatchCount        int `json:"matchCount"`
				ReportsCount      int `json:"reportsCount"`
				ScoutReportsCount int `json:"scoutReportsCount"`
				PostCount         int `json:"postCount"`
				FollowersCount    int `json:"followersCount"`
			} `json:"stats"`
			ScoutReports struct {
				Total int                      `json:"total"`
				List  []map[string]interface{} `json:"list"`
			} `json:"scoutReports"`
			Timeline []map[string]interface{} `json:"timeline"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatalf("success = false")
	}
	if body.Data.Profile.ID != 1001 || body.Data.Profile.DisplayName != "子墨" {
		t.Fatalf("unexpected profile: %+v", body.Data.Profile)
	}
	if body.Data.Profile.School != "" {
		t.Fatalf("anonymous response should hide school, got %q", body.Data.Profile.School)
	}
	if body.Data.Stats.PhysicalTestCount != 1 ||
		body.Data.Stats.WeeklyReportCount != 1 ||
		body.Data.Stats.MatchCount != 1 ||
		body.Data.Stats.ReportsCount != 1 ||
		body.Data.Stats.ScoutReportsCount != 1 ||
		body.Data.Stats.PostCount != 1 ||
		body.Data.Stats.FollowersCount != 1 {
		t.Fatalf("unexpected stats: %+v", body.Data.Stats)
	}
	if body.Data.ScoutReports.Total != 1 || len(body.Data.ScoutReports.List) != 1 {
		t.Fatalf("unexpected scout reports: %+v", body.Data.ScoutReports)
	}
	if len(body.Data.Timeline) < 6 {
		t.Fatalf("timeline length = %d, want at least 6", len(body.Data.Timeline))
	}
	foundScoutReport := false
	for _, item := range body.Data.Timeline {
		if item["type"] == "scout_report" {
			foundScoutReport = true
			break
		}
	}
	if !foundScoutReport {
		t.Fatalf("timeline should include scout_report: %+v", body.Data.Timeline)
	}
}

func TestGetHomepageRejectsPrivateProfile(t *testing.T) {
	router, db := setupPlayerHomepageTestRouter(t)
	seedPlayerHomepageData(t, db, false)

	req := httptest.NewRequest(http.MethodGet, "/api/players/1001/homepage", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
