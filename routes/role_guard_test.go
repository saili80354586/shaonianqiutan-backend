package routes_test

import (
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
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/routes"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleGuardRouteTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	t.Setenv("JWT_SECRET", "role-guard-route-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "role-guard-route.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Scout{},
		&models.Coach{},
		&models.ClubCoach{},
		&models.TeamCoach{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = oldDB
	})

	teamRepo := repositories.NewTeamRepository(db)
	weeklyReportRepo := repositories.NewWeeklyReportRepository(db)
	coachService := services.NewCoachService(db, teamRepo, weeklyReportRepo, repositories.NewMatchSummaryRepository(db))

	router := gin.New()
	api := router.Group("/api")
	routes.SetupScoutRoutes(api, controllers.NewScoutController(services.NewScoutService(db)))
	routes.SetupCoachRoutes(
		api,
		controllers.NewCoachController(coachService),
		teamRepo,
		controllers.NewFootballExperienceController(coachService),
		controllers.NewWeeklyReportController(nil, db),
		db,
	)

	return router, db
}

func createRoleGuardRouteTestUser(t *testing.T, db *gorm.DB, phone string, role models.UserRole) models.User {
	t.Helper()

	user := models.User{
		Phone:       phone,
		Password:    "hashed-password",
		Role:        role,
		CurrentRole: role,
		Status:      models.StatusActive,
		Name:        "路由权限测试用户",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func performRoleGuardRouteRequest(t *testing.T, router *gin.Engine, method, path string, user models.User) *httptest.ResponseRecorder {
	t.Helper()

	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestScoutProfileRouteRejectsPlayerBeforeAutoCreate(t *testing.T) {
	router, db := setupRoleGuardRouteTestRouter(t)
	user := createRoleGuardRouteTestUser(t, db, "13900007101", models.RoleUser)

	rec := performRoleGuardRouteRequest(t, router, http.MethodGet, "/api/scout/profile", user)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var scoutCount int64
	if err := db.Model(&models.Scout{}).Where("user_id = ?", user.ID).Count(&scoutCount).Error; err != nil {
		t.Fatalf("count scouts: %v", err)
	}
	if scoutCount != 0 {
		t.Fatalf("expected no scout profile to be created, got %d", scoutCount)
	}
}

func TestCoachProfileRouteRejectsCoachProfileWithoutMembership(t *testing.T) {
	router, db := setupRoleGuardRouteTestRouter(t)
	user := createRoleGuardRouteTestUser(t, db, "13900007102", models.RoleUser)
	if err := db.Create(&models.Coach{UserID: user.ID}).Error; err != nil {
		t.Fatalf("create coach profile: %v", err)
	}

	rec := performRoleGuardRouteRequest(t, router, http.MethodGet, "/api/coach/profile", user)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCoachProfileRouteAllowsTeamCoachMembership(t *testing.T) {
	router, db := setupRoleGuardRouteTestRouter(t)
	user := createRoleGuardRouteTestUser(t, db, "13900007103", models.RoleUser)
	if err := db.Create(&models.TeamCoach{
		TeamID:   1,
		UserID:   user.ID,
		Role:     models.CoachRoleAssistant,
		Status:   "active",
		JoinedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create team coach: %v", err)
	}

	rec := performRoleGuardRouteRequest(t, router, http.MethodGet, "/api/coach/profile", user)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
