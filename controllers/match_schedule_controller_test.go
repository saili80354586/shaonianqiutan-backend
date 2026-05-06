package controllers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMatchScheduleControllerTest(t *testing.T) (*gin.Engine, *gorm.DB, models.User, models.MatchSchedule) {
	t.Helper()

	t.Setenv("JWT_SECRET", "match-schedule-controller-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "match-schedule-controller.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.Team{},
		&models.TeamPlayer{},
		&models.MatchSchedule{},
		&models.MatchSummary{},
		&models.Notification{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = oldDB
	})

	clubUser := models.User{
		Phone:       "13900007301",
		Password:    "hashed-password",
		Role:        models.RoleClub,
		CurrentRole: models.RoleClub,
		Status:      models.StatusActive,
		Name:        "俱乐部管理员",
	}
	if err := db.Create(&clubUser).Error; err != nil {
		t.Fatalf("create club user: %v", err)
	}
	club := models.Club{UserID: clubUser.ID, Name: "赛程测试俱乐部"}
	if err := db.Create(&club).Error; err != nil {
		t.Fatalf("create club: %v", err)
	}
	team := models.Team{ClubID: club.ID, Name: "U12 测试队", AgeGroup: "U12", Status: models.TeamStatusActive}
	if err := db.Create(&team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	player := models.User{
		Phone:       "13900007302",
		Password:    "hashed-password",
		Role:        models.RoleUser,
		CurrentRole: models.RoleUser,
		Status:      models.StatusActive,
		Name:        "测试球员",
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := db.Create(&models.TeamPlayer{
		TeamID:   team.ID,
		UserID:   player.ID,
		Status:   "active",
		JoinedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create team player: %v", err)
	}

	homeScore := 2
	awayScore := 1
	schedule := models.MatchSchedule{
		ClubID:    club.ID,
		TeamID:    team.ID,
		Name:      "赛程闭环测试赛",
		MatchType: models.MatchScheduleTypeFriendly,
		Opponent:  "测试对手",
		MatchTime: time.Now().Add(-24 * time.Hour),
		Location:  "主场",
		HomeScore: &homeScore,
		AwayScore: &awayScore,
		Status:    models.MatchScheduleStatusCompleted,
		CreatedBy: clubUser.ID,
	}
	if err := db.Create(&schedule).Error; err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	router := gin.New()
	ctrl := controllers.NewMatchScheduleController(services.NewClubService(db), db)
	router.POST("/api/club/match-schedules/:id/summary", middleware.AuthMiddleware(), middleware.ClubRoleMiddleware(), ctrl.CreateMatchSummaryFromSchedule)

	return router, db, clubUser, schedule
}

func TestCreateMatchSummaryFromScheduleLinksSchedule(t *testing.T) {
	router, db, clubUser, schedule := setupMatchScheduleControllerTest(t)

	token, err := middleware.GenerateToken(clubUser.ID, clubUser.Phone)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	path := fmt.Sprintf("/api/club/match-schedules/%d/summary", schedule.ID)
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var updated models.MatchSchedule
	if err := db.First(&updated, schedule.ID).Error; err != nil {
		t.Fatalf("reload schedule: %v", err)
	}
	if updated.MatchSummaryID == nil {
		t.Fatalf("expected schedule to link generated match summary")
	}

	var summary models.MatchSummary
	if err := db.First(&summary, *updated.MatchSummaryID).Error; err != nil {
		t.Fatalf("load summary: %v", err)
	}
	if summary.MatchName != schedule.Name || summary.TeamID != schedule.TeamID || summary.PlayerCount != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.OurScore != 2 || summary.OppScore != 1 || summary.Result != "win" {
		t.Fatalf("unexpected score/result: %d-%d %s", summary.OurScore, summary.OppScore, summary.Result)
	}

	req = httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("second status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var summaryCount int64
	if err := db.Model(&models.MatchSummary{}).Where("team_id = ?", schedule.TeamID).Count(&summaryCount).Error; err != nil {
		t.Fatalf("count summaries: %v", err)
	}
	if summaryCount != 1 {
		t.Fatalf("expected one summary after duplicate request, got %d", summaryCount)
	}
}
